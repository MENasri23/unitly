package domain

import (
	"context"
	"time"
)

type TokenType string

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

	TokenType TokenType
}

type AccessTokenClaims struct {
	TokenClaims
	UserID string
}

type TokenService interface {
	GenerateAccessToken(ctx context.Context, claims AccessTokenClaims) (string, error)
	ValidateAccessToken(ctx context.Context, token string) (AccessTokenClaims, error)
}
