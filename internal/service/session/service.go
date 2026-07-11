package sessionsvc

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"unitly/internal/domain"
	"unitly/internal/util/hash"
	"unitly/internal/util/rand"
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
	refreshToken, err := rand.GenereteDefaultToken()
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

	claims := domain.AccessTokenClaims{UserID: userID.String()}
	accessToken, err := s.tokenService.GenerateAccessToken(ctx, claims)

	return accessToken, refreshToken, err
}

func (s *sessionService) RefreshSession(ctx context.Context, oldRefreshToken string) (string, string, error) {
	if oldRefreshToken == "" {
		return "", "", domain.ErrInvalidToken
	}

	tokenHash := hash.SHA256(oldRefreshToken)
	var activeUserID uuid.UUID
	var newRefreshToken string

	err := s.uow.Do(ctx, func(txCtx context.Context) error {
		session, err := s.sessionRepo.GetByRefreshToken(txCtx, tokenHash)
		if err != nil {
			return domain.ErrInvalidToken
		}

		if session.IsExpired() || session.IsRevoked() {
			return domain.ErrInvalidToken
		}

		if session.IsCompromised() {
			activeUserID = session.UserID
			return domain.ErrSessionCompromised
		}

		user, err := s.userRepo.GetByGUID(txCtx, session.UserID)
		if err != nil || !user.IsActive {
			return domain.ErrInvalidToken
		}

		revokedAt := time.Now()
		session.RevokedAt = &revokedAt
		if err = s.sessionRepo.Revoke(txCtx, session); err != nil {
			return err
		}

		newRefreshToken, err = rand.GenereteDefaultToken()
		if err != nil {
			return err
		}

		newSession := &domain.Session{
			ID:               uuid.New(),
			UserID:           user.GUID,
			RefreshTokenHash: hash.SHA256(newRefreshToken),
			ExpiresAt:        time.Now().Add(s.refreshTokenExpiry),
		}

		if err := s.sessionRepo.Create(txCtx, newSession); err != nil {
			return err
		}

		activeUserID = user.GUID
		return nil
	})

	if errors.Is(err, domain.ErrSessionCompromised) && activeUserID != uuid.Nil {
		_ = s.sessionRepo.RevokeAll(ctx, activeUserID)
		return "", "", domain.ErrInvalidToken
	}

	if err != nil {
		return "", "", err
	}

	claims := domain.AccessTokenClaims{UserID: activeUserID.String()}
	accessToken, err := s.tokenService.GenerateAccessToken(ctx, claims)

	return accessToken, newRefreshToken, err
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
