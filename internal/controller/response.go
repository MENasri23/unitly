package controller

import (
	"errors"
	"net/http"
	"unitly/internal/log"
	"unitly/pkg/jsonutil"
)

func writeJSON(w http.ResponseWriter, status int, v any, l log.Logger) {
	if err := jsonutil.WriteJSON(w, status, v); err != nil {
		l.Error("failed to write response", "error", err)
		writeInternalServerError(w, errors.New("failed to write response"))
	}
}

func writeInternalServerError(w http.ResponseWriter, err error) {
	jsonutil.WriteErrJSON(w, http.StatusInternalServerError, err)
}
