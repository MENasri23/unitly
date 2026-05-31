package auth

import (
	"context"
	"time"
	"unitly/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

const issuer = "unitly"

type JWTService struct {
	secret []byte
	expiry time.Duration
	issuer string
}

type customClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func NewJWTService(secret []byte, expiry time.Duration) *JWTService {
	return &JWTService{
		secret: secret,
		expiry: expiry,
		issuer: issuer,
	}
}

func (s *JWTService) GenerateToken(_ context.Context, claims domain.TokenClaims) (string, error) {
	now := time.Now()

	issuer := claims.Issuer
	if issuer == "" {
		issuer = s.issuer
	}

	expiresAt := claims.ExpiresAt
	if expiresAt.IsZero() {
		if claims.Expiry > 0 {
			expiresAt = now.Add(claims.Expiry)
		} else {
			expiresAt = now.Add(s.expiry)
		}
	}

	issuedAt := claims.IssuedAt
	if issuedAt.IsZero() {
		issuedAt = now
	}

	notBefore := claims.NotBefore
	if notBefore.IsZero() {
		notBefore = now
	}

	jwtClaims := customClaims{
		UserID: claims.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   claims.Subject,
			Audience:  jwt.ClaimStrings(claims.Audience),
			ID:        claims.ID,
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(notBefore),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *JWTService) ValidateToken(_ context.Context, tokenString string) (domain.TokenClaims, error) {
	claims := &customClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidToken
		}
		return s.secret, nil
	})

	if err != nil {
		return domain.TokenClaims{}, domain.ErrInvalidToken
	}

	if !token.Valid {
		return domain.TokenClaims{}, domain.ErrInvalidToken
	}

	var issuedAt, expiresAt, notBefore time.Time

	if claims.RegisteredClaims.IssuedAt != nil {
		issuedAt = claims.RegisteredClaims.IssuedAt.Time
	}
	if claims.RegisteredClaims.ExpiresAt != nil {
		expiresAt = claims.RegisteredClaims.ExpiresAt.Time
	}
	if claims.RegisteredClaims.NotBefore != nil {
		notBefore = claims.RegisteredClaims.NotBefore.Time
	}

	return domain.TokenClaims{
		UserID:    claims.UserID,
		Issuer:    claims.RegisteredClaims.Issuer,
		Subject:   claims.RegisteredClaims.Subject,
		Audience:  claims.RegisteredClaims.Audience,
		ID:        claims.RegisteredClaims.ID,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
		Expiry:    expiresAt.Sub(issuedAt),
		NotBefore: notBefore,
	}, nil
}
