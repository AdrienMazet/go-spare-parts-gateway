package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/service"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/xerrors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSparePartHandler(t *testing.T) {
	tests := []struct {
		name               string
		inputID            string
		serviceResult      *api.SparePart
		serviceError       error
		expectedStatusCode int
		expectedID         string
		expectedError      *api.ErrorResponse
	}{
		{
			name:               "spare part found",
			inputID:            "sp-001",
			serviceResult:      getValidSparePart("sp-001"),
			expectedStatusCode: http.StatusOK,
			expectedID:         "sp-001",
		},
		{
			name:               "spare part not found",
			inputID:            "sp-999",
			serviceError:       xerrors.ErrorEntityNotFound.Msgf("spare part with id %s not found", "sp-999"),
			expectedStatusCode: http.StatusNotFound,
			expectedError: &api.ErrorResponse{
				Title:  "Not Found",
				Status: http.StatusNotFound,
				Detail: "spare part with id sp-999 not found",
			},
		},
		{
			name:               "unknown service error",
			inputID:            "sp-001",
			serviceError:       errors.New("repository unavailable"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedError: &api.ErrorResponse{
				Title:  "Internal Server Error",
				Status: http.StatusInternalServerError,
				Detail: "Internal Server Error",
			},
		},
		{
			name:    "invalid spare part returned by service",
			inputID: "sp-001",
			serviceResult: func() *api.SparePart {
				sp := getValidSparePart("sp-001")
				sp.Reference = "invalid-reference"

				return sp
			}(),
			expectedStatusCode: http.StatusInternalServerError,
			expectedError: &api.ErrorResponse{
				Title:  "Internal Server Error",
				Status: http.StatusInternalServerError,
				Detail: "spare part with id sp-001 is invalid",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/spare-part/"+tt.inputID, nil)
			req.SetPathValue("id", tt.inputID)

			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockService := service.NewMockSparePartsService(mockController)

			mockService.EXPECT().Retrieve(tt.inputID).Return(tt.serviceResult, tt.serviceError)

			handler := NewSparePartHandler(mockService)
			handler.GetSparePart(w, req)

			res := w.Result()
			defer func() {
				assert.NoError(t, res.Body.Close())
			}()

			require.Equal(t, tt.expectedStatusCode, res.StatusCode)
			assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

			if tt.expectedID != "" {
				var sp api.SparePart
				require.NoError(t, json.NewDecoder(res.Body).Decode(&sp))

				assert.Equal(t, tt.expectedID, sp.ID)
			} else if tt.expectedError != nil {
				var errBody api.ErrorResponse
				require.NoError(t, json.NewDecoder(res.Body).Decode(&errBody))

				assert.Equal(t, *tt.expectedError, errBody)
			}
		})
	}
}

func getValidSparePart(id string) *api.SparePart {
	return &api.SparePart{
		ID:          id,
		Reference:   "BRK-PAD-001",
		Label:       "Front Brake Pads",
		Brand:       "Brembo",
		Category:    "BRAKING",
		Description: "High performance front brake pads",
		Offers: []api.Offer{
			{
				ID:            "off-001",
				Supplier:      "PartsPro",
				Price:         4599,
				Currency:      "EUR",
				StockQuantity: 42,
				DeliveryDelay: "PT48H",
			},
		},
	}
}
