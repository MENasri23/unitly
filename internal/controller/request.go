package controller

import (
	"net/http"
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
