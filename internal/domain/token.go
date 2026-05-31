package domain

import (
	"context"
	"time"
)

// TokenClaims represents the application's view of claims contained in a JWT.
// It includes the standard JWT registered claims plus common custom fields
// used across the app.
type TokenClaims struct {
	// Standard / registered claims
	Issuer    string
	Subject   string
	Audience  []string
	Expiry    time.Duration
	ExpiresAt time.Time
	IssuedAt  time.Time
	NotBefore time.Time
	ID        string

	// Application-specific claims
	UserID int64
}

type TokenService interface {
	GenerateToken(ctx context.Context, claims TokenClaims) (string, error)
	ValidateToken(ctx context.Context, token string) (TokenClaims, error)
}
