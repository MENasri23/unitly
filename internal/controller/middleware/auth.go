package middleware

import (
	"context"
	"net/http"
	"strings"
	"unitly/internal/domain"
	apphttp "unitly/internal/http"
	"unitly/internal/http/header"
	"unitly/pkg/jsonutil"

	"github.com/google/uuid"
)

const userCtxKey = contextKey("userCtxKey")
const sessionCtxKey = contextKey("sessionCtxKey")
const bearerPrefix = "Bearer "

type contextKey string

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

		user, err := m.userRepo.GetByGUID(r.Context(), userID)
		if err != nil {
			unauthorized(w)
			return
		}

		ctx := context.WithValue(r.Context(), userCtxKey, user)
		ctx = context.WithValue(ctx, sessionCtxKey, claims.SessionID)
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

func GetUserFromCtx(ctx context.Context) *domain.User {
	user, _ := ctx.Value(userCtxKey).(*domain.User)
	return user
}

func GetSessionIdFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(sessionCtxKey).(string)
	return id
}
