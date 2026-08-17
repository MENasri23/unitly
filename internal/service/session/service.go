package sessionsvc

import (
	"context"
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

func (s *sessionService) ListSessions(ctx context.Context, userID uuid.UUID, page, pageSize int) (domain.SessionPage, error) {
	if page < 1 {
		page = 1
	}
	switch {
	case pageSize <= 0:
		pageSize = domain.DefaultSessionPageSize
	case pageSize > domain.MaxSessionPageSize:
		pageSize = domain.MaxSessionPageSize
	}
	offset := (page - 1) * pageSize

	sessions, total, err := s.sessionRepo.ListActiveByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return domain.SessionPage{}, err
	}

	return domain.SessionPage{
		Data:       sessions,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
	}, nil
}

func (s *sessionService) CreateSession(ctx context.Context, userID uuid.UUID, meta domain.ClientMetadata) (accessToken, refreshToken string, err error) {
	refreshToken, err = token.GenerateDefaultToken()
	if err != nil {
		return "", "", err
	}

	sessionID := uuid.New()
	device := resolveDevice(meta.UserAgent)
	now := time.Now()
	session := &domain.Session{
		ID:               sessionID,
		UserID:           userID,
		FamilyID:         sessionID, // Root session sets FamilyID to its own ID
		ParentSessionID:  nil,       // Root session has no parent
		RefreshTokenHash: hash.SHA256(refreshToken),

		IPAddress:  meta.IPAddress,
		UserAgent:  meta.UserAgent,
		DeviceType: device.DeviceType,
		OSName:     device.OSName,
		OSVersion:  device.OSVersion,
		Label:      device.ClientLabel,

		CreatedAt:  now,
		LastUsedAt: now,
		ExpiresAt:  now.Add(s.refreshTokenExpiry),
	}

	claims := domain.AccessTokenClaims{
		UserID:    userID.String(),
		SessionID: session.ID.String(),
	}
	accessToken, err = s.tokenService.GenerateAccessToken(ctx, claims)
	if err != nil {
		return "", "", err
	}

	if err = s.sessionRepo.Create(ctx, session); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *sessionService) RefreshSession(ctx context.Context, oldRefreshToken string, meta domain.ClientMetadata) (accessToken, newRefreshToken string, err error) {
	if oldRefreshToken == "" {
		return "", "", domain.ErrInvalidToken
	}

	tokenHash := hash.SHA256(oldRefreshToken)
	session, err := s.sessionRepo.GetByRefreshToken(ctx, tokenHash)
	if err != nil {
		return "", "", domain.ErrInvalidToken
	}

	// If reuse is detected outside the grace period, revoke the entire family lineage
	if session.IsCompromised() {
		detachedCtx := context.WithoutCancel(ctx)
		_ = s.revokeFamily(detachedCtx, session.FamilyID)
		return "", "", domain.ErrInvalidToken
	}

	if !session.IsActive() {
		return "", "", domain.ErrInvalidToken
	}

	claims := domain.AccessTokenClaims{UserID: session.UserID.String()}
	accessToken, err = s.tokenService.GenerateAccessToken(ctx, claims)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err = token.GenerateDefaultToken()
	if err != nil {
		return "", "", err
	}

	now := time.Now()

	revokedAt := now
	session.RevokedAt = &revokedAt

	device := resolveDevice(meta.UserAgent)
	parentID := session.ID
	newSession := &domain.Session{
		ID:               uuid.New(),
		UserID:           session.UserID,
		FamilyID:         session.FamilyID,
		ParentSessionID:  &parentID,
		RefreshTokenHash: hash.SHA256(newRefreshToken),

		IPAddress:  meta.IPAddress,
		UserAgent:  meta.UserAgent,
		DeviceType: device.DeviceType,
		OSName:     device.OSName,
		OSVersion:  device.OSVersion,
		Label:      device.ClientLabel,

		CreatedAt:  now,
		LastUsedAt: now,
		ExpiresAt:  now.Add(s.refreshTokenExpiry),
	}

	err = s.uow.Do(ctx, func(txCtx context.Context) error {
		if err = s.sessionRepo.Revoke(txCtx, session); err != nil {
			return err
		}

		// Database-level UNIQUE constraint on parent_session_id prevents race conditions
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

func (s *sessionService) revokeFamily(ctx context.Context, familyID uuid.UUID) error {
	return s.sessionRepo.RevokeFamily(ctx, familyID)
}

func (s *sessionService) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.RevokeAll(ctx, userID)
}

func (s *sessionService) RevokeAllSessionsExcept(ctx context.Context, userID, sessionID uuid.UUID) error {
	return s.sessionRepo.RevokeAllExcept(ctx, userID, sessionID)
}

func (s *sessionService) RevokeSessionByID(ctx context.Context, userID, sessionID uuid.UUID) error {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if session.UserID != userID {
		return domain.ErrSessionForbidden
	}
	if !session.IsActive() {
		return nil
	}

	now := time.Now()
	session.RevokedAt = &now
	return s.sessionRepo.Revoke(ctx, session)
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
