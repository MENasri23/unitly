// Package response provides small HTTP helper utilities for working with JSON APIs.
package jsonutil

import (
	"encoding/json"
	"net/http"
)

var (
	// maxReadBytes defines the maximum allowed size of a JSON request body.
	// This prevents clients from sending excessively large payloads that could
	// exhaust memory or degrade service performance.
	//
	// Default: 1MB
	maxReadBytes int64 = 1024 * 1024
)

// errorResponse defines the standard structure used for all JSON error responses.
// This ensures consistent API error formatting across the application.
type errorResponse struct {
	Message string `json:"message"`
}

// WriteErrJSON writes a JSON-formatted error response to the client.
//
// It sets the HTTP status code and returns a JSON object in the form:
//
//	{
//	  "message": "error description"
//	}
//
// Example usage:
//
//	return jsonutil.WriteErrJSON(w, http.StatusBadRequest, "invalid input")
//
// Returns an error if JSON encoding fails (rare, but possible if the writer fails).
func WriteErrJSON(
	w http.ResponseWriter,
	statusCode int,
	message string,
) error {
	return WriteJSON(w, statusCode, &errorResponse{Message: message})
}

// WriteJSON writes any Go value as a JSON response.
//
// Example usage:
//
//	type User struct { Name string }
//	return jsonutil.WriteJSON(w, http.StatusOK, User{Name: "John"})
func WriteJSON(
	w http.ResponseWriter,
	statusCode int,
	data any,
) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// ReadJSON safely reads and decodes a JSON request body into the provided struct.
//
// Example usage:
//
//	var req CreateUserRequest
//	err := jsonutil.ReadJSON(w, r, &req)
//
// Security notes:
//   - Requests larger than maxReadBytes will be rejected automatically.
//   - Unknown JSON fields will cause an error (decoder.DisallowUnknownFields()).
//
// Parameters:
//   - w: ResponseWriter (used for MaxBytesReader limiting)
//   - r: HTTP request containing JSON body
//   - data: pointer to struct where decoded JSON will be stored
//
// Returns:
//   - error if decoding fails, body is too large, or JSON is invalid
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
