package controllers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal JSON response", "error", err)

		fallback := api.ErrorResponse{
			Title:  "Internal Server Error",
			Status: http.StatusInternalServerError,
			Detail: "Failed to encode response",
		}

		data, err = json.Marshal(fallback)
		if err != nil {
			slog.Error("failed to marshal fallback JSON response", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(data); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}
