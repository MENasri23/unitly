package sessionsvc

import (
	"context"
	"errors"
	"time"

	"unitly/internal/domain"
	"unitly/internal/token"
	"unitly/internal/util/hash"

	"github.com/google/uuid"
)

type sessionService struct {
	sessionRepo        domain.SessionRepository
	userRepo           domain.UserRepository
	tokenService       domain.TokenService
	uow                domain.UnitOfWork
	refreshTokenExpiry time.Duration
	resetTokenExpiry   time.Duration
}

func NewSessionService(
	sessionRepo domain.SessionRepository,
	userRepo domain.UserRepository,
	tokenService domain.TokenService,
	uow domain.UnitOfWork,
	refreshTokenExpiry time.Duration,
	resetTokenExpiry time.Duration,
) domain.SessionService {
	return &sessionService{
		sessionRepo:        sessionRepo,
		userRepo:           userRepo,
		tokenService:       tokenService,
		uow:                uow,
		refreshTokenExpiry: refreshTokenExpiry,
		resetTokenExpiry:   resetTokenExpiry,
	}
}

func (s *sessionService) CreateSession(ctx context.Context, userID uuid.UUID) (string, string, error) {
	refreshToken, err := token.GenerateDefaultToken()
	if err != nil {
		return "", "", err
	}

	session := &domain.Session{
		ID:               uuid.New(),
		UserID:           userID,
		RefreshTokenHash: hash.SHA256(refreshToken),
		ExpiresAt:        time.Now().Add(s.refreshTokenExpiry),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", "", err
	}

	claims := domain.AccessTokenClaims{
		UserID:    userID.String(),
		SessionID: session.ID.String(),
	}
	accessToken, err := s.tokenService.GenerateAccessToken(ctx, claims)

	return accessToken, refreshToken, err
}

func (s *sessionService) RefreshSession(ctx context.Context, oldRefreshToken string) (string, string, error) {
	if oldRefreshToken == "" {
		return "", "", domain.ErrInvalidToken
	}

	tokenHash := hash.SHA256(oldRefreshToken)
	session, err := s.sessionRepo.GetByRefreshToken(ctx, tokenHash)
	if err != nil {
		return "", "", domain.ErrInvalidToken
	}

	if session.IsCompromised() {
		context := context.WithoutCancel(ctx)
		_ = s.RevokeAllSessions(context, session.UserID)
		return "", "", domain.ErrInvalidToken
	}

	if session.IsExpired() || session.IsRevoked() {
		return "", "", domain.ErrInvalidToken
	}

	claims := domain.AccessTokenClaims{UserID: session.UserID.String()}
	accessToken, err := s.tokenService.GenerateAccessToken(ctx, claims)
	if err != nil {
		return "", "", err
	}

	var newRefreshToken string

	err = s.uow.Do(ctx, func(txCtx context.Context) error {
		revokedAt := time.Now()
		session.RevokedAt = &revokedAt
		if err = s.sessionRepo.Revoke(txCtx, session); err != nil {
			return err
		}

		newRefreshToken, err = token.GenerateDefaultToken()
		if err != nil {
			return err
		}

		newSession := &domain.Session{
			ID:               uuid.New(),
			UserID:           session.UserID,
			RefreshTokenHash: hash.SHA256(newRefreshToken),
			ExpiresAt:        time.Now().Add(s.refreshTokenExpiry),
		}

		if err := s.sessionRepo.Create(txCtx, newSession); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

func (s *sessionService) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.RevokeAll(ctx, userID)
}

func (s *sessionService) GenerateResetToken(ctx context.Context, userID uuid.UUID) (string, error) {
	claims := domain.AccessTokenClaims{
		UserID: userID.String(),
		TokenClaims: domain.TokenClaims{
			Expiry: s.resetTokenExpiry,
		},
	}
	return s.tokenService.GenerateResetPasswordToken(ctx, claims)
}

func (s *sessionService) ValidateResetToken(ctx context.Context, token string) (uuid.UUID, error) {
	claims, err := s.tokenService.ValidateResetPasswordToken(ctx, token)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(claims.UserID)
}

func (s *sessionService) RevokeSession(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return domain.ErrInvalidToken
	}

	tokenHash := hash.SHA256(refreshToken)
	session, err := s.sessionRepo.GetByRefreshToken(ctx, tokenHash)
	if err != nil {
		return domain.ErrInvalidToken
	}

	if session.IsExpired() || session.IsRevoked() {
		return nil
	}

	revokedAt := time.Now()
	session.RevokedAt = &revokedAt
	return s.sessionRepo.Revoke(ctx, session)
}

func (s *sessionService) RevokeAllSessionsExcept(ctx context.Context, userID, sessionID uuid.UUID) error {
	return s.sessionRepo.RevokeAllExcept(ctx, userID, sessionID)
}
