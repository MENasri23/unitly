package controller

import (
	"errors"
	"net/http"
	apphttp "unitly/internal/http"
	"unitly/internal/log"
)

func writeJSON(w http.ResponseWriter, status int, v any, l log.Logger) {
	if err := apphttp.WriteJSON(w, status, v); err != nil {
		l.Error("failed to write response", "error", err)
		writeInternalServerError(w, errors.New("failed to write response"))
	}
}

func writeInternalServerError(w http.ResponseWriter, err error) {
	apphttp.WriteErrJSON(w, http.StatusInternalServerError, err)
}

func writeBadRequestError(w http.ResponseWriter, err error) {
	apphttp.WriteErrJSON(w, http.StatusBadRequest, err)
}
