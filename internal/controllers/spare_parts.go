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

// GetSparePart handles GET requests to retrieve a spare part by reference.
func (h SparePartsHandler) GetSparePart(w http.ResponseWriter, r *http.Request) {
	reference := r.PathValue("reference")

	sp, err := h.sparePartsService.Retrieve(r.Context(), reference)
	if err != nil {
		slog.Warn("failed to retrieve spare part", "reference", reference, "error", err)
		HandleError(w, err)
		return
	}

	if err := Validator.Struct(sp); err != nil {
		slog.Warn("invalid spare part with reference", "reference", reference)
		HandleError(w, xerrors.ErrorInvalidEntity.Msgf("spare part with reference %s is invalid", reference))
		return
	}

	writeJSON(w, http.StatusOK, sp)
}
