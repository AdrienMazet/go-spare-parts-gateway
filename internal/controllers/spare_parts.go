package controllers

import (
	"log/slog"
	"net/http"

	"github.com/adrienmazet/go-spare-parts-gateway/internal/service"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/xerrors"
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
		slog.Warn("failed to retrieve spare part", "id", id, "error", err)
		HandleError(w, err)
		return
	}

	if err := Validator.Struct(sp); err != nil {
		slog.Warn("invalid spare part with id", "id", id)
		HandleError(w, xerrors.ErrorInvalidEntity.Msgf("spare part with id %s is invalid", id))
		return
	}

	writeJSON(w, http.StatusOK, sp)
}
