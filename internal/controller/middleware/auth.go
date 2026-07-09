package middleware

import (
	"context"
	"net/http"
	"strings"
	"unitly/internal/domain"
	"unitly/pkg/jsonutil"

	"github.com/google/uuid"
)

const userCtxKey = userContext("userCtxKey")

type userContext string

type AuthMiddleware struct {
	tokenService domain.TokenService
	userRepo     domain.UserRepository
}

func NewAuthMiddleware(
	tokenService domain.TokenService,
	userRepo domain.UserRepository,
) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
		userRepo:     userRepo,
	}
}

func (m *AuthMiddleware) AuthorizeUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if !strings.HasPrefix(authHeader, "Bearer ") {
				jsonutil.WriteErrJSON(w, http.StatusUnauthorized, domain.ErrInvalidToken)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := m.tokenService.ValidateAccessToken(r.Context(), token)
			if err != nil {
				jsonutil.WriteErrJSON(w, http.StatusUnauthorized, domain.ErrInvalidToken)
				return
			}

			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				jsonutil.WriteErrJSON(w, http.StatusUnauthorized, domain.ErrInvalidToken)
				return
			}

			user, err := m.userRepo.GetByGUID(r.Context(), userID)
			if err != nil {
				jsonutil.WriteErrJSON(w, http.StatusUnauthorized, domain.ErrInvalidToken)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), userCtxKey, user))

			next.ServeHTTP(w, r)
		})
}

func GetUserFromCtx(ctx context.Context) *domain.User {
	user, _ := ctx.Value(userCtxKey).(*domain.User)
	return user
}
