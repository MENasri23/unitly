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
const bearerPrefix = "Bearer "

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

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if !startWithBearer(authHeader) {
			unauthorized(w)
			return
		}

		token := trimBearer(authHeader)

		claims, err := m.tokenService.ValidateAccessToken(r.Context(), token)
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

		r = r.WithContext(context.WithValue(r.Context(), userCtxKey, user))
		next.ServeHTTP(w, r)
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
