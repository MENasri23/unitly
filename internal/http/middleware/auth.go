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

func getTokenFromRequest(r *http.Request) (string, error) {
	if apphttp.IsWebPlatform(r) {
		atCookie, err := r.Cookie("access_token")
		if err != nil {
			return "", err
		}
		return atCookie.Value, nil
	} else {
		authHeader := r.Header.Get(header.Authorization)
		if !startWithBearer(authHeader) {
			return "", domain.ErrInvalidToken
		}
		return trimBearer(authHeader), nil
	}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		token, err := getTokenFromRequest(r)
		if err != nil {
			unauthorized(w)
			return
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

		ctx := reqctx.WithUserID(r.Context(), userID)
		ctx = reqctx.WithSessionID(ctx, sessionID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(f)
}

func (m *AuthMiddleware) RequireUser(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		userID, ok := reqctx.GetUserID(r.Context())
		if !ok {
			unauthorized(w)
			return
		}

		user, err := m.userRepo.GetByGUID(r.Context(), userID)
		if err != nil {
			unauthorized(w)
			return
		}

		ctx := reqctx.WithUser(r.Context(), user)
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
