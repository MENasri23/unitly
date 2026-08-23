package reqctx

import (
	"context"
	"unitly/internal/domain"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type ctxKey int

const (
	sessionIDKey ctxKey = iota
	userIDKey
	userKey
)

func WithSessionID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, sessionIDKey, id)
}

func GetSessionID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(sessionIDKey).(uuid.UUID)
	return id, ok
}

func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

func WithUser(ctx context.Context, user *domain.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func GetUser(ctx context.Context) *domain.User {
	user, _ := ctx.Value(userKey).(*domain.User)
	return user
}

func GetClientIP(ctx context.Context) string {
	return middleware.GetClientIP(ctx)
}
