package controller

import (
	"net/http"
	"unitly/internal/domain"
	apphttp "unitly/internal/http"
	"unitly/internal/http/reqctx"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	user := reqctx.GetUser(r.Context())
	if user == nil {
		apphttp.WriteErrJSON(w, http.StatusInternalServerError, domain.ErrInternal)
		return
	}

	payload := map[string]any{
		"status": "authorized",
		"user":   user,
	}
	apphttp.WriteJSON(w, http.StatusOK, payload)
}
