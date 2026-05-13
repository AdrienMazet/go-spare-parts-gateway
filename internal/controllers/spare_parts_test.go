package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/service"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/xerrors"
	"github.com/golang/mock/gomock"
)

func TestSparePartHandler(t *testing.T) {
	tests := []struct {
		name               string
		inputID            string
		expectedStatusCode int
		expectedID         string
		expectedError      *api.ErrorResponse
	}{
		{
			name:               "spare part found",
			inputID:            "sp-001",
			expectedStatusCode: http.StatusOK,
			expectedID:         "sp-001",
		},
		{
			name:               "spare part not found",
			inputID:            "sp-999",
			expectedStatusCode: http.StatusNotFound,
			expectedError: &api.ErrorResponse{
				Title:  "Not Found",
				Status: http.StatusNotFound,
				Detail: "spare part with id sp-999 not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/spare-part/"+tt.inputID, nil)
			req.SetPathValue("id", tt.inputID)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := service.NewMockSparePartsService(ctrl)

			if tt.expectedID != "" {
				validSparePart := &api.SparePart{
					ID:          tt.inputID,
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
				mockService.EXPECT().Retrieve(tt.inputID).Return(validSparePart, nil)
			} else {
				mockService.EXPECT().Retrieve(tt.inputID).Return(nil, xerrors.ErrorEntityNotFound.Msgf("spare part with id %s not found", tt.inputID))
			}

			handler := NewSparePartHandler(mockService)
			handler.GetSparePart(w, req)

			res := w.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()

			if res.StatusCode != tt.expectedStatusCode {
				t.Fatalf("expected status %d, got %d", tt.expectedStatusCode, res.StatusCode)
			}

			contentType := res.Header.Get("Content-Type")
			if !strings.HasPrefix(contentType, "application/json") {
				t.Fatalf("expected Content-Type application/json, got %q", contentType)
			}

			if tt.expectedID != "" {
				var sp api.SparePart
				if err := json.NewDecoder(res.Body).Decode(&sp); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}

				if sp.ID != tt.expectedID {
					t.Fatalf("expected ID %q, got %q", tt.expectedID, sp.ID)
				}
			} else if tt.expectedError != nil {
				var errBody api.ErrorResponse
				if err := json.NewDecoder(res.Body).Decode(&errBody); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}

				if errBody.Title != tt.expectedError.Title {
					t.Errorf("expected error title %q, got %q", tt.expectedError.Title, errBody.Title)
				}

				if errBody.Status != tt.expectedError.Status {
					t.Errorf("expected error status %d, got %d", tt.expectedError.Status, errBody.Status)
				}

				if errBody.Detail != tt.expectedError.Detail {
					t.Errorf("expected error detail %q, got %q", tt.expectedError.Detail, errBody.Detail)
				}
			}
		})
	}
}
