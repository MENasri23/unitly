package controller

import (
	"net/http"
	"unitly/internal/domain"
	"unitly/internal/http/reqctx"
	"unitly/pkg/jsonutil"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	user := reqctx.GetUser(r.Context())
	if user == nil {
		jsonutil.WriteErrJSON(w, http.StatusInternalServerError, domain.ErrInternal)
		return
	}

	payload := map[string]any{
		"status": "authorized",
		"user":   user,
	}
	jsonutil.WriteJSON(w, http.StatusOK, payload)
}
