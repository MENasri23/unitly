package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrSessionCompromised = errors.New("session is compromised")

type Session struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	Version int

	RefreshTokenHash string

	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

func (s *Session) IsRevoked() bool {
	return s.RevokedAt != nil
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// A session is only suspicious if it was used AFTER it was already revoked
func (s *Session) IsCompromised() bool {
	var revokedAt = s.RevokedAt
	if revokedAt == nil {
		return false
	}

	// Allow a 5-second grace period for network retries
	const gracePeriod = 5 * time.Second
	return time.Since(*revokedAt) > gracePeriod
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*Session, error)
	Revoke(ctx context.Context, session *Session) error
	RevokeAll(ctx context.Context, userID uuid.UUID) error
	GetByRefreshToken(ctx context.Context, refreshTokenHash string) (*Session, error)
	CleanupOldSessions(ctx context.Context, expiredThreshold, revokedThreshold time.Duration) (int64, error)
}


// SessionService isolates token management, token validation, and session auditing.
type SessionService interface {
	CreateSession(ctx context.Context, userID uuid.UUID) (accessToken, refreshToken string, err error)
	RefreshSession(ctx context.Context, oldRefreshToken string) (accessToken, newRefreshToken string, err error)
	RevokeAllSessions(ctx context.Context, userID uuid.UUID) error
	GenerateResetToken(ctx context.Context, userID uuid.UUID) (string, error)
	ValidateResetToken(ctx context.Context, token string) (uuid.UUID, error)
	RevokeSession(ctx context.Context, refreshToken string) error
}