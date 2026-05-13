package controllers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/adrienmazet/go-spare-parts-gateway/api"

	"github.com/adrienmazet/go-spare-parts-gateway/internal/service"
)

type SparePartsHandler struct {
	sparePartsService service.SparePartsService
}

func NewSparePartHandler(sparePartsService service.SparePartsService) SparePartsHandler {
	return SparePartsHandler{
		sparePartsService: sparePartsService,
	}
}

// GetSparePart handles GET requests to retrieve a spare part by ID.
func (h SparePartsHandler) GetSparePart(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	sp, err := h.sparePartsService.Retrieve(id)
	if err != nil {
		// TODO : error handling
		slog.Warn("spare part not found", "id", id)

		notFoundErr := api.ErrorResponse{
			Title:  "Not Found",
			Status: http.StatusNotFound,
			Detail: fmt.Sprintf("spare part with id %s not found", id),
		}

		writeJSON(w, http.StatusNotFound, notFoundErr)
		return
	}

	if err := Validator.Struct(sp); err != nil {
		slog.Warn("invalid spare part with id", "id", id)

		invalidSparePartError := api.ErrorResponse{
			Title:  "Invalid spare part",
			Status: http.StatusInternalServerError,
			Detail: fmt.Sprintf("spare part with id %s is invalid", id),
		}

		writeJSON(w, http.StatusInternalServerError, invalidSparePartError)
		return
	}

	writeJSON(w, http.StatusOK, sp)
}

// TODO : where does it belong ?
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
