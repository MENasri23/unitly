package auth

import (
	"context"
	"time"
	"unitly/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const issuer = "unitly"

const (
	tokenTypeAccess    = TokenType("access")
	tokenTypeResetPass = TokenType("reset-pass")
)

type JWTService struct {
	secret  []byte
	expiry  time.Duration
	issuer  string
	options []jwt.ParserOption
}

type accessTokenClaims struct {
	UserID    string    `json:"user_id"`
	TokenType TokenType `json:"token_type"`
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
		TokenType:        tokenTypeAccess,
		RegisteredClaims: s.registeredClaims(claims.TokenClaims),
	}

	return s.generateAccessToken(jwtClaims)
}

func (s *JWTService) GenerateResetPasswordToken(_ context.Context, claims domain.AccessTokenClaims) (string, error) {
	jwtClaims := accessTokenClaims{
		UserID:           claims.UserID,
		TokenType:        tokenTypeResetPass,
		RegisteredClaims: s.registeredClaims(claims.TokenClaims),
	}

	return s.generateAccessToken(jwtClaims)
}

func (s *JWTService) generateAccessToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
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

	if !token.Valid || claims.TokenType != tokenTypeAccess {
		return domain.AccessTokenClaims{}, domain.ErrInvalidToken
	}

	return s.accessTokenClaims(claims), nil
}

func (s *JWTService) ValidateResetPasswordToken(_ context.Context, tokenString string) (domain.AccessTokenClaims, error) {
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

	if !token.Valid || claims.TokenType != tokenTypeResetPass {
		return domain.AccessTokenClaims{}, domain.ErrInvalidToken
	}

	return s.accessTokenClaims(claims), nil
}

func (s *JWTService) accessTokenClaims(claims *accessTokenClaims) domain.AccessTokenClaims {
	var issuedAt, expiresAt, notBefore time.Time

	if claims.IssuedAt != nil {
		issuedAt = claims.IssuedAt.Time
	}
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}
	if claims.NotBefore != nil {
		notBefore = claims.NotBefore.Time
	}

	return domain.AccessTokenClaims{
		TokenClaims: domain.TokenClaims{
			Issuer:    claims.Issuer,
			Subject:   claims.Subject,
			Audience:  claims.Audience,
			ID:        claims.ID,
			IssuedAt:  issuedAt,
			ExpiresAt: expiresAt,
			Expiry:    expiresAt.Sub(issuedAt),
			NotBefore: notBefore,
		},
		UserID: claims.UserID,
	}
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
