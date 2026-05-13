package controllers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/adrienmazet/go-spare-parts-gateway/internal/xerrors"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

// HandleError writes an API error response for err.
func HandleError(w http.ResponseWriter, err error) {
	status, response := errorResponse(err)
	writeJSON(w, status, response)
}

func errorResponse(err error) (int, api.ErrorResponse) {
	response := api.ErrorResponse{
		Title:  http.StatusText(http.StatusInternalServerError),
		Status: http.StatusInternalServerError,
		Detail: http.StatusText(http.StatusInternalServerError),
	}

	var appError *xerrors.Error
	if errors.As(err, &appError) {
		response.Detail = appError.Msg
	}

	switch {
	case errors.Is(err, xerrors.ErrorEntityNotFound):
		response.Status = http.StatusNotFound
		response.Title = http.StatusText(http.StatusNotFound)
	}

	return response.Status, response
}

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
