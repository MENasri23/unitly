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
		apphttp.InternalServerError(w, errors.New("failed to write response"))
	}
}
