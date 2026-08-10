package reqctx

import (
	"context"
	"github.com/google/uuid"
	"unitly/internal/domain"
)

type ctxKey int

const (
	sessionIDKey ctxKey = iota
	userKey
)

func WithSessionID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, sessionIDKey, id)
}

func GetSessionID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(sessionIDKey).(uuid.UUID)
	return id, ok
}

func WithUser(ctx context.Context, user *domain.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func GetUser(ctx context.Context) *domain.User {
	user, _ := ctx.Value(userKey).(*domain.User)
	return user
}
