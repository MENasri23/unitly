package middleware

import (
	"net/http"
	"strings"
	"unitly/internal/domain"
	apphttp "unitly/internal/http"
	"unitly/internal/http/header"
	"unitly/internal/http/reqctx"
	"unitly/pkg/jsonutil"

	"github.com/google/uuid"
)

const bearerPrefix = "Bearer "

type AuthMiddleware struct {
	tokenValidator domain.TokenValidator
	userRepo       domain.UserRepository
}

func NewAuthMiddleware(
	tokenValidator domain.TokenValidator,
	userRepo domain.UserRepository,
) *AuthMiddleware {
	return &AuthMiddleware{
		tokenValidator: tokenValidator,
		userRepo:       userRepo,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		var token string

		if apphttp.IsWebPlatform(r) {
			atCookie, err := r.Cookie("access_token")
			if err != nil {
				jsonutil.WriteErrJSON(w, http.StatusUnauthorized, domain.ErrInvalidToken)
				return
			}
			token = atCookie.Value
		} else {
			authHeader := r.Header.Get(header.Authorization)
			if !startWithBearer(authHeader) {
				unauthorized(w)
				return
			}
			token = trimBearer(authHeader)
		}

		claims, err := m.tokenValidator.ValidateAccessToken(r.Context(), token)
		if err != nil {
			unauthorized(w)
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			unauthorized(w)
			return
		}
		sessionID, err := uuid.Parse(claims.SessionID)
		if err != nil {
			unauthorized(w)
			return
		}

		user, err := m.userRepo.GetByGUID(r.Context(), userID)
		if err != nil {
			unauthorized(w)
			return
		}

		ctx := reqctx.WithUser(r.Context(), user)
		ctx = reqctx.WithSessionID(ctx, sessionID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(f)
}

func startWithBearer(header string) bool {
	return strings.HasPrefix(header, bearerPrefix)
}

func trimBearer(header string) string {
	return strings.TrimPrefix(header, bearerPrefix)
}

func unauthorized(w http.ResponseWriter) {
	jsonutil.WriteErrJSON(w, http.StatusUnauthorized, domain.ErrInvalidToken)
}
