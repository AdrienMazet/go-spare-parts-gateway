package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adrienmazet/go-spare-parts-gateway/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	payload := api.ErrorResponse{
		Title:  "Not Found",
		Status: http.StatusNotFound,
		Detail: "spare part not found",
	}

	w := httptest.NewRecorder()

	writeJSON(w, http.StatusNotFound, payload)

	res := w.Result()
	defer func() {
		assert.NoError(t, res.Body.Close())
	}()

	require.Equal(t, http.StatusNotFound, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var body api.ErrorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	assert.Equal(t, payload, body)
}

func TestWriteJSONMarshalFailure(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	writeJSON(w, http.StatusOK, make(chan int))

	res := w.Result()
	defer func() {
		assert.NoError(t, res.Body.Close())
	}()

	require.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var body api.ErrorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	assert.Equal(t, api.ErrorResponse{
		Title:  "Internal Server Error",
		Status: http.StatusInternalServerError,
		Detail: "Failed to encode response",
	}, body)
}
