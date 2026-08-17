package controller

import (
	"net/http"
	"strconv"
	"unitly/pkg/jsonutil"
)

// bindAndValidate collapses the decode+validate steps that used to be
// hand-repeated at the top of every handler, and always returns the real
func bindAndValidate[T any](r *http.Request) (T, error) {
	var req T
	if err := jsonutil.ReadJSON(r, &req); err != nil {
		return req, err
	}
	if err := validateStruct(req); err != nil {
		return req, err
	}
	return req, nil
}

// queryInt reads a positive integer query param, falling back to def if
// it's missing, not parseable, or not positive.
func queryInt(r *http.Request, key string, def int) int {
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
