package controllers

import (
	"log/slog"
	"net/http"
	"time"
)

// LogRequests logs one structured line for each HTTP request.
func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		slog.Info(
			"http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration", time.Since(startedAt).String(),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}

	r.status = status
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(status)
}
