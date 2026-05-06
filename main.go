package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
)

func main() {
	swagger, err := api.GetSwagger()
	if err != nil {
		log.Fatal(err)
	}

	swagger.Servers = nil

	mux := http.NewServeMux()
	mux.HandleFunc("GET /spare-part/{id}", sparePartHandler)

	middleware := nethttpmiddleware.OapiRequestValidatorWithOptions(
		swagger,
		&nethttpmiddleware.Options{
			ErrorHandlerWithOpts: openAPIErrorHandler,
		},
	)

	handler := middleware(mux)

	log.Fatal(http.ListenAndServe(":8080", handler))
}

func openAPIErrorHandler(
	ctx context.Context,
	err error,
	w http.ResponseWriter,
	r *http.Request,
	opts nethttpmiddleware.ErrorHandlerOpts,
) {
	slog.Warn(
		"openapi request validation failed",
		"error", err,
		"method", r.Method,
		"path", r.URL.Path,
		"status", opts.StatusCode,
	)

	status := opts.StatusCode
	if status == 0 {
		status = http.StatusBadRequest
	}

	title := http.StatusText(status)
	if title == "" {
		title = "Bad Request"
	}

	writeJSON(w, status, api.ErrorResponse{
		Title:  title,
		Status: status,
		Detail: err.Error(),
	})
}

func sparePartHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	spareParts, err := getSpareParts()
	if err != nil {
		slog.Error("failed to retrieve spare parts", "error", err)

		internalServerError := api.ErrorResponse{
			Title:  "Internal Server Error",
			Status: http.StatusInternalServerError,
			Detail: "Failed to retrieve spare parts",
		}

		writeJSON(w, http.StatusInternalServerError, internalServerError)
		return
	}

	sp, found := findSparePartByID(spareParts, id)
	if !found {
		slog.Warn("spare part not found", "id", id)

		notFoundErr := api.ErrorResponse{
			Title:  "Not Found",
			Status: http.StatusNotFound,
			Detail: fmt.Sprintf("Spare part with id %s not found", id),
		}

		writeJSON(w, http.StatusNotFound, notFoundErr)
		return
	}

	writeJSON(w, http.StatusOK, sp)
}

func findSparePartByID(spareParts []api.SparePart, sparePartID string) (*api.SparePart, bool) {
	for i := range spareParts {
		if spareParts[i].ID == sparePartID {
			return &spareParts[i], true
		}
	}

	return nil, false
}

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

// TODO : retrieve spare parts from external data sources
func getSpareParts() ([]api.SparePart, error) {
	spareParts := []api.SparePart{
		{
			ID:          "sp-001",
			Reference:   "BRK-PAD-4521",
			Label:       "Front Brake Pads",
			Brand:       "Brembo",
			Category:    api.BRAKING,
			Description: "High performance front brake pads for urban and highway use",
			Offers: []api.Offer{
				{
					ID:            "off-001",
					Supplier:      "PartsPro",
					Price:         4599,
					Currency:      api.EUR,
					StockQuantity: 42,
					DeliveryDelay: "PT48H",
				},
				{
					ID:            "off-002",
					Supplier:      "AutoStock",
					Price:         3999,
					Currency:      api.EUR,
					StockQuantity: 8,
					DeliveryDelay: "P3D",
				},
			},
		},
		{
			ID:          "sp-002",
			Reference:   "ENG-FLT-7823",
			Label:       "Oil Filter",
			Brand:       "Mann",
			Category:    api.FILTERS,
			Description: "Standard oil filter compatible with most 4-cylinder engines",
			Offers: []api.Offer{
				{
					ID:            "off-003",
					Supplier:      "PartsPro",
					Price:         899,
					Currency:      api.EUR,
					StockQuantity: 150,
					DeliveryDelay: "PT2H",
				},
			},
		},
		{
			ID:          "sp-003",
			Reference:   "SUS-SPR-3341",
			Label:       "Rear Coil Spring",
			Brand:       "KYB",
			Category:    api.SUSPENSION,
			Description: "OEM replacement rear coil spring for medium sedans",
			Offers:      []api.Offer{},
		},
	}

	for i := range spareParts {
		if err := Validator.Struct(spareParts[i]); err != nil {
			return nil, fmt.Errorf("invalid spare part with id %q: %w", spareParts[i].ID, err)
		}
	}

	return spareParts, nil
}
