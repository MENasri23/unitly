package http

import (
	"net/http"
	"strconv"
	"unitly/internal/client"
	"unitly/internal/http/header"
)

func IsWebPlatform(r *http.Request) bool {
	return header.GetClientPlatform(r) == client.WebPlatform
}

// BindAndValidate collapses the decode+validate steps that used to be
// hand-repeated at the top of every handler, and always returns the real
func BindAndValidate[T any](r *http.Request) (T, error) {
	var req T
	if err := ReadJSON(r, &req); err != nil {
		return req, err
	}
	if err := ValidateStruct(req); err != nil {
		return req, err
	}
	return req, nil
}

// QueryInt reads a positive integer query param, falling back to def if
// it's missing, not parseable, or not positive.
func QueryInt(r *http.Request, key string, def int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return def
	}
	return n
}
