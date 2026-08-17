package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Pagination bounds for ListSessions/ListActiveByUserID. PageSize is
// clamped into (0, MaxSessionPageSize] by the service regardless of what a
// caller asks for.
const (
	DefaultSessionPageSize = 10
	MaxSessionPageSize     = 20
)

var (
	ErrSessionNotFound = NewNotFoundError("session")
	// Returned by RevokeSessionByID when the session exists but belongs to
	// a different user than the one making the request.
	ErrSessionForbidden = NewError("SESSION_FORBIDDEN", "session does not belong to this user")
	// Returned when a user attempts to delete their current active session.
	ErrSessionInUse = NewError("SESSION_IN_USE", "cannot delete your current session")
	
)

type DeviceType string

const (
	DeviceTypeDesktop DeviceType = "Desktop"
	DeviceTypeAndroid DeviceType = "Android"
	DeviceTypeIOS     DeviceType = "iOS"
	DeviceTypeBot     DeviceType = "Bot"
	DeviceTypeUnknown DeviceType = "Unknown"
)

type Session struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	FamilyID        uuid.UUID
	ParentSessionID *uuid.UUID

	RefreshTokenHash string

	IPAddress  string
	UserAgent  string
	DeviceType DeviceType
	OSName     string
	OSVersion  string

	Label string

	CreatedAt  time.Time 
	LastUsedAt time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
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

// IsActive reports whether the session is currently usable: not revoked and
// not expired. This is the filter used for the "your active sessions" list.
func (s *Session) IsActive() bool {
	return !s.IsRevoked() && !s.IsExpired()
}

// SessionPage is one page of a user's active sessions, most recently used
// first, plus the total count across all pages.
type SessionPage = Page[*Session]

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*Session, error)
	Revoke(ctx context.Context, session *Session) error
	RevokeAll(ctx context.Context, userID uuid.UUID) error
	RevokeAllExcept(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error
	RevokeFamily(ctx context.Context, familyID uuid.UUID) error
	GetByRefreshToken(ctx context.Context, refreshTokenHash string) (*Session, error)
	CleanupOldSessions(ctx context.Context, expiredThreshold, revokedThreshold time.Duration) (int64, error)

	// ListActiveByUserID returns one page (limit/offset, most-recently-used
	// first) of userID's non-revoked, non-expired sessions, plus the total
	// count across all pages. This backs the "your devices / active
	// sessions" listing endpoint.
	ListActiveByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) (sessions []*Session, total int, err error)
}

// SessionService isolates token management, token validation, and session auditing.
type SessionService interface {
	CreateSession(ctx context.Context, userID uuid.UUID, meta ClientMetadata) (accessToken, refreshToken string, err error)
	RefreshSession(ctx context.Context, oldRefreshToken string, meta ClientMetadata) (accessToken, newRefreshToken string, err error)
	
	GenerateResetToken(ctx context.Context, userID uuid.UUID) (string, error)
	ValidateResetToken(ctx context.Context, token string) (uuid.UUID, error)
	
	RevokeAllSessions(ctx context.Context, userID uuid.UUID) error
	RevokeAllSessionsExcept(ctx context.Context, userID, sessionID uuid.UUID) error
	RevokeSession(ctx context.Context, refreshToken string) error

	// RevokeSessionByID revokes a single session by ID on behalf of userID
	// (the "log out this device" action from the sessions list). It's
	// distinct from RevokeSession: the caller knows the target session's
	// ID from the list, not that device's raw refresh token, which is
	// never retrievable once hashed.
	RevokeSessionByID(ctx context.Context, userID, currSessionID, sessionID uuid.UUID) error

	// ListSessions returns a page of the caller's active sessions, most
	// recently used first. page is 1-indexed; both page and pageSize are
	// clamped to sane bounds (DefaultSessionPageSize/MaxSessionPageSize)
	// regardless of what's passed in.
	ListSessions(ctx context.Context, userID uuid.UUID, page, pageSize int) (SessionPage, error)
}
