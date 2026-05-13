package controllers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

func OpenAPIErrorHandler(
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
