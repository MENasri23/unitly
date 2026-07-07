package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID     uuid.UUID
	UserID uuid.UUID

	RefreshTokenHash string

	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

func (s *Session) IsRevoked() bool {
	return s.RevokedAt != nil
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*Session, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAll(ctx context.Context, userID uuid.UUID) error
	GetByRefreshToken(ctx context.Context, refreshTokenHash string) (*Session, error)
}
