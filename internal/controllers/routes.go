package controllers

import (
	"net/http"
)

// SetupRoutes configures the API endpoints.
func SetupRoutes(r *http.ServeMux, handler SparePartsHandler) {
	r.HandleFunc("GET /spare-part/{reference}", handler.GetSparePart)
}
