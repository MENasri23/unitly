package middleware

import (
	"net/http"
	"unitly/internal/domain"
	"unitly/internal/http/reqctx"
	"unitly/pkg/jsonutil"
)

// RequireMinRoleLevel guards a route behind a minimum role level, e.g. RequireMinRoleLevel(domain.LevelAdmin).
func RequireMinRoleLevel(min int16) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m := reqctx.GetMembership(r.Context())
			if m == nil || !m.HasMinLevel(min) {
				jsonutil.WriteErrJSON(w, http.StatusForbidden, domain.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
