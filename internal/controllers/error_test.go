package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/xerrors"
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
			err:            xerrors.ErrorEntityNotFound.Msgf("spare part with id %s not found", "sp-999"),
			expectedStatus: http.StatusNotFound,
			expectedTitle:  "Not Found",
			expectedDetail: "spare part with id sp-999 not found",
		},
		{
			name:           "invalid entity",
			err:            xerrors.ErrorInvalidEntity.Msgf("spare part with id %s is invalid", "sp-001"),
			expectedStatus: http.StatusInternalServerError,
			expectedTitle:  "Internal Server Error",
			expectedDetail: "spare part with id sp-001 is invalid",
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
				if err := res.Body.Close(); err != nil {
					t.Errorf("failed to close response body: %v", err)
				}
			}()

			if res.StatusCode != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			var body api.ErrorResponse
			if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			if body.Title != tt.expectedTitle {
				t.Errorf("expected title %q, got %q", tt.expectedTitle, body.Title)
			}

			if body.Status != tt.expectedStatus {
				t.Errorf("expected body status %d, got %d", tt.expectedStatus, body.Status)
			}

			if body.Detail != tt.expectedDetail {
				t.Errorf("expected detail %q, got %q", tt.expectedDetail, body.Detail)
			}
		})
	}
}
