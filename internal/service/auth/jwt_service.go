package auth

import (
	"context"
	"time"
	"unitly/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

const issuer = "unitly"

type JWTService struct {
	secret  []byte
	expiry  time.Duration
	issuer  string
	options []jwt.ParserOption
}

type accessTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewJWTService(secret string, expiry time.Duration) *JWTService {
	return &JWTService{
		secret:  []byte(secret),
		expiry:  expiry,
		issuer:  issuer,
		options: jwtOptions(issuer),
	}
}

func (s *JWTService) GenerateAccessToken(_ context.Context, claims domain.AccessTokenClaims) (string, error) {
	jwtClaims := accessTokenClaims{
		UserID:           claims.UserID,
		RegisteredClaims: s.registeredClaims(claims.TokenClaims),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *JWTService) ValidateAccessToken(_ context.Context, tokenString string) (domain.AccessTokenClaims, error) {
	claims := &accessTokenClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		s.parseKey,
		s.options...,
	)

	if err != nil {
		return domain.AccessTokenClaims{}, domain.ErrInvalidToken
	}

	if !token.Valid {
		return domain.AccessTokenClaims{}, domain.ErrInvalidToken
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

	return domain.AccessTokenClaims{
		TokenClaims: domain.TokenClaims{
			Issuer:    claims.RegisteredClaims.Issuer,
			Subject:   claims.RegisteredClaims.Subject,
			Audience:  claims.RegisteredClaims.Audience,
			ID:        claims.RegisteredClaims.ID,
			IssuedAt:  issuedAt,
			ExpiresAt: expiresAt,
			Expiry:    expiresAt.Sub(issuedAt),
			NotBefore: notBefore,
		},
		UserID: claims.UserID,
	}, nil
}

func (s *JWTService) registeredClaims(claims domain.TokenClaims) jwt.RegisteredClaims {
	now := time.Now()

	expiresAt := claims.ExpiresAt
	if expiresAt.IsZero() {
		if claims.Expiry > 0 {
			expiresAt = now.Add(claims.Expiry)
		} else {
			expiresAt = now.Add(s.expiry)
		}
	}

	issuer := s.issuer
	issuedAt := claims.IssuedAt
	if issuedAt.IsZero() {
		issuedAt = now
	}

	notBefore := claims.NotBefore
	if notBefore.IsZero() {
		notBefore = now
	}

	return jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   claims.Subject,
		Audience:  jwt.ClaimStrings(claims.Audience),
		ID:        claims.ID,
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		NotBefore: jwt.NewNumericDate(notBefore),
	}
}

func (s *JWTService) parseKey(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, domain.ErrInvalidToken
	}
	return s.secret, nil
}

func jwtOptions(issuer string) []jwt.ParserOption {
	return []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithExpirationRequired(),
		jwt.WithNotBeforeRequired(),
		jwt.WithIssuedAt(),
		jwt.WithIssuer(issuer),
	}
}
