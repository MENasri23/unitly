package pkg

import (
	"encoding/json"
	"net/http"
)

var (
	maxReadBytes int64 = 1024 * 1024 // 1MB
)

type errorResponse struct {
	Message string `json:"message"`
}

func WriteErrJSON(
	w http.ResponseWriter,
	statusCode int,
	message string,
) error {
	return WriteJSON(w, statusCode, &errorResponse{Message: message})
}

func WriteJSON(
	w http.ResponseWriter,
	statusCode int,
	data any,
) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

func ReadJSON(
	w http.ResponseWriter,
	r *http.Request,
	data any,
) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxReadBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}
