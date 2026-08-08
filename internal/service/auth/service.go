package auth

import (
	"context"
	"errors"
	"unitly/internal/domain"
	identitysvc "unitly/internal/service/identity"
	"unitly/internal/util/validate"
)

var (
	authPurpose      = domain.OTPPurpose("register-login")
	resetPassPurpose = domain.OTPPurpose("password-reset")
)

type AuthResult struct {
	AccessToken  string
	RefreshToken string
}

type AuthService struct {
	otpService      domain.OTPService
	identityService identitysvc.IdentityService
	sessionService  domain.SessionService
	smsService      domain.SmsService
	tokenService    domain.TokenService
	uow             domain.UnitOfWork
}

func NewAuthService(
	otpSvc domain.OTPService,
	identityService identitysvc.IdentityService,
	sessionService domain.SessionService,
	smsService domain.SmsService,
	tokenService domain.TokenService,
	uow domain.UnitOfWork,
) *AuthService {
	return &AuthService{
		otpService:      otpSvc,
		identityService: identityService,
		sessionService:  sessionService,
		smsService:      smsService,
		tokenService:    tokenService,
		uow:             uow,
	}
}

func (s *AuthService) RequestAuthByOTP(ctx context.Context, phone string) (string, error) {
	if !validate.IsValidPhone(phone) {
		return "", domain.ErrInvalidPhone
	}

	if err := s.otpService.EnsureAllowed(ctx, phone, authPurpose); err != nil {
		return "", err
	}

	return s.otpService.RequestOTP(ctx, phone, authPurpose, true)
}

func (s *AuthService) VerifyAuthByOTP(
	ctx context.Context,
	token string,
	inputOTP string,
) (*AuthResult, error) {
	phone, err := s.otpService.VerifyOTP(ctx, token, inputOTP, authPurpose)
	if err != nil {
		return nil, err
	}

	user, err := s.identityService.GetOrCreateUser(ctx, phone)
	if err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := s.sessionService.CreateSession(ctx, user.GUID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) AuthenticateByPassword(
	ctx context.Context,
	phone string,
	password string,
) (*AuthResult, error) {
	if !s.isValidPhone(phone) {
		return nil, domain.ErrInvalidPhone
	}

	if password == "" {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := s.identityService.Authenticate(ctx, phone, password)
	if err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := s.sessionService.CreateSession(ctx, user.GUID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RequestResetPasswordByOTP(
	ctx context.Context,
	phone string,
) (string, error) {
	if !validate.IsValidPhone(phone) {
		return "", domain.ErrInvalidPhone
	}

	if err := s.otpService.EnsureAllowed(ctx, phone, resetPassPurpose); err != nil {
		return "", err
	}

	hasPassword, err := s.identityService.HasPassword(ctx, phone)
	if err != nil {
		return "", err
	}

	return s.otpService.RequestOTP(ctx, phone, resetPassPurpose, hasPassword)
}

func (s *AuthService) VerifyResetPasswordByOTP(
	ctx context.Context,
	token string,
	inputOTP string,
) (string, error) {
	phone, err := s.otpService.VerifyOTP(ctx, token, inputOTP, resetPassPurpose)
	if err != nil {
		return "", err
	}

	user, err := s.identityService.GetByPhone(ctx, phone)
	if err != nil || !user.IsActive {
		return "", domain.ErrInvalidOTP
	}

	return s.sessionService.GenerateResetToken(ctx, user.GUID)
}

func (s *AuthService) RefreshAuthByToken(
	ctx context.Context,
	refreshToken string,
) (*AuthResult, error) {
	accessToken, newRefreshToken, err := s.sessionService.RefreshSession(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return domain.ErrInvalidToken
	}

	err := s.sessionService.RevokeSession(ctx, refreshToken)
	if err != nil && !errors.Is(err, domain.ErrInvalidToken) {
		return err
	}

	return nil
}

func (s *AuthService) isValidPhone(phone string) bool {
	return validate.IsValidPhone(phone)
}

func (s *AuthService) FinalizeResetPasswordByOTP(
	ctx context.Context, sign, newPassword string,
) error {
	userUUID, err := s.sessionService.ValidateResetToken(ctx, sign)
	if err != nil {
		return err
	}

	return s.uow.Do(ctx, func(txCtx context.Context) error {
		if err := s.identityService.UpdatePassword(txCtx, userUUID, newPassword); err != nil {
			return err
		}

		return s.sessionService.RevokeAllSessions(txCtx, userUUID)
	})
}
