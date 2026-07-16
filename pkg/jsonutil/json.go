// Package response provides small HTTP helper utilities for working with JSON APIs.
package jsonutil

import (
	"encoding/json"
	"errors"
	"net/http"
	"unitly/internal/domain"
	"unitly/internal/http/header"
)

// errorResponse defines the standard structure used for all JSON error responses.
// This ensures consistent API error formatting across the application.
type errorResponse struct {
	Code    domain.ErrorCode `json:"code"`
	Message string           `json:"message"`
}

// WriteErrJSON writes a JSON-formatted error response to the client.
//
// It sets the HTTP status code and returns a JSON object in the form:
//
//	{
//	  "code": "INVALID_REQUEST",
//	  "message": "error description"
//	}
//
// Example usage:
//
//	return jsonutil.WriteErrJSON(w, http.StatusBadRequest, domain.NewError("INVALID_REQUEST", "invalid input"))
//
// Returns an error if JSON encoding fails (rare, but possible if the writer fails).
func WriteErrJSON(
	w http.ResponseWriter,
	statusCode int,
	err error,
) error {
	if err == nil {
		err = domain.ErrInternal
	}

	payload := errorResponse{Message: err.Error()}
	var domainErr domain.Error
	if errors.As(err, &domainErr) {
		payload.Code = domainErr.Code()
	}

	return WriteJSON(w, statusCode, &payload)
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
	w.Header().Set(header.ContentType, "application/json")
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
//   - Unknown JSON fields will cause an error (decoder.DisallowUnknownFields()).
//
// Parameters:
//   - r: HTTP request containing JSON body
//   - data: pointer to struct where decoded JSON will be stored
//
// Returns:
//   - error if decoding fails, body is too large, or JSON is invalid
func ReadJSON(
	r *http.Request,
	data any,
) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(data); err != nil {
		return domain.NewError(domain.ErrorCode("INVALID_JSON"), err.Error())
	}

	return nil
}
