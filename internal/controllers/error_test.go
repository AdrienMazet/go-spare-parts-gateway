package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/xerrors"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedTitle  string
		expectedDetail string
	}{
		{
			name:           "entity not found",
			err:            xerrors.ErrorEntityNotFound.Msgf("spare part with reference %s not found", "UNKNOWN-REF"),
			expectedStatus: http.StatusNotFound,
			expectedTitle:  "Not Found",
			expectedDetail: "spare part with reference UNKNOWN-REF not found",
		},
		{
			name:           "wrapped entity not found",
			err:            xerrors.ErrorEntityNotFound.Msgf("spare part with reference %s not found", "UNKNOWN-REF").Wrap(errors.New("repository lookup failed")),
			expectedStatus: http.StatusNotFound,
			expectedTitle:  "Not Found",
			expectedDetail: "spare part with reference UNKNOWN-REF not found",
		},
		{
			name:           "invalid entity",
			err:            xerrors.ErrorInvalidEntity.Msgf("spare part with reference %s is invalid", "BRK-PAD-4521"),
			expectedStatus: http.StatusInternalServerError,
			expectedTitle:  "Internal Server Error",
			expectedDetail: "spare part with reference BRK-PAD-4521 is invalid",
		},
		{
			name:           "unknown error",
			err:            errors.New("database unavailable"),
			expectedStatus: http.StatusInternalServerError,
			expectedTitle:  "Internal Server Error",
			expectedDetail: "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()

			HandleError(w, tt.err)

			res := w.Result()
			defer func() {
				assert.NoError(t, res.Body.Close())
			}()

			require.Equal(t, tt.expectedStatus, res.StatusCode)

			var body api.ErrorResponse
			require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

			assert.Equal(t, tt.expectedTitle, body.Title)
			assert.Equal(t, tt.expectedStatus, body.Status)
			assert.Equal(t, tt.expectedDetail, body.Detail)
		})
	}
}

func TestOpenAPIErrorHandler(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedStatus int
		expectedTitle  string
	}{
		{
			name:           "uses validation middleware status",
			statusCode:     http.StatusUnprocessableEntity,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedTitle:  "Unprocessable Entity",
		},
		{
			name:           "defaults empty status to bad request",
			statusCode:     0,
			expectedStatus: http.StatusBadRequest,
			expectedTitle:  "Bad Request",
		},
		{
			name:           "uses fallback title for unknown status text",
			statusCode:     499,
			expectedStatus: 499,
			expectedTitle:  "Bad Request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			expectedErr := errors.New("request does not match OpenAPI schema")
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/spare-part/BRK-PAD-4521", nil)

			OpenAPIErrorHandler(
				context.Background(),
				expectedErr,
				w,
				r,
				nethttpmiddleware.ErrorHandlerOpts{StatusCode: tt.statusCode},
			)

			res := w.Result()
			defer func() {
				assert.NoError(t, res.Body.Close())
			}()

			require.Equal(t, tt.expectedStatus, res.StatusCode)
			assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

			var body api.ErrorResponse
			require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

			assert.Equal(t, tt.expectedTitle, body.Title)
			assert.Equal(t, tt.expectedStatus, body.Status)
			assert.Equal(t, expectedErr.Error(), body.Detail)
		})
	}
}
