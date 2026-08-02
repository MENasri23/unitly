package auth

import (
	"context"
	"errors"
	"unicode"
	"unitly/internal/domain"
	"unitly/internal/util/eql"
	"unitly/internal/util/hash"
	"unitly/internal/util/validate"

	"github.com/google/uuid"
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
	otpService     domain.OTPService
	userRepo       domain.UserRepository
	sessionService domain.SessionService
	smsService     domain.SmsService
	tokenService   domain.TokenService
	uow            domain.UnitOfWork
}

func NewAuthService(
	otpSvc domain.OTPService,
	userRepo domain.UserRepository,
	sesSvc domain.SessionService,
	smsService domain.SmsService,
	tokenService domain.TokenService,
	uow domain.UnitOfWork,
) *AuthService {
	return &AuthService{
		otpService:     otpSvc,
		userRepo:       userRepo,
		sessionService: sesSvc,
		smsService:     smsService,
		tokenService:   tokenService,
		uow:            uow,
	}
}

func (s *AuthService) RequestAuthByOTP(ctx context.Context, phone string) (string, error) {
	if !validate.IsValidPhone(phone) {
		return "", domain.ErrInvalidPhone
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

	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		if !isUserNotFound(err) {
			return nil, err
		}

		user, err = s.registerUser(ctx, phone)
		if err != nil {
			return nil, err
		}
	}

	if user != nil {
		accessToken, refreshToken, err := s.sessionService.CreateSession(ctx, user.GUID)
		if err != nil {
			return nil, err
		}

		return &AuthResult{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}, nil
	}

	return nil, domain.ErrInternal
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

	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		if isUserNotFound(err) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.IsActive || user.HashedPassword == nil || *user.HashedPassword == "" {
		return nil, domain.ErrInvalidCredentials
	}

	if !eql.IsEqual(hash.SHA256(password), *user.HashedPassword) {
		return nil, domain.ErrInvalidCredentials
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

	userExists, err := s.userRepo.ExistsByPhone(ctx, phone)
	if err != nil {
		return "", err
	}

	return s.otpService.RequestOTP(ctx, phone, resetPassPurpose, userExists)
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

	user, err := s.userRepo.GetByPhone(ctx, phone)
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

func (s *AuthService) isValidPhone(phone string) bool {
	return validate.IsValidPhone(phone)
}

func isUserNotFound(err error) bool {
	return errors.Is(err, domain.ErrUserNotFound)
}

func (s *AuthService) registerUser(ctx context.Context, phone string) (*domain.User, error) {
	user, err := s.createUser(ctx, phone)
	if err == nil {
		return user, nil
	}

	// Possible race condition: user may have been created by another goroutine.
	// Attempt to retrieve the user. If successful, proceed; otherwise return original error.
	if errors.Is(err, domain.ErrAlreadyExist) {
		user, lookupErr := s.userRepo.GetByPhone(ctx, phone)
		if lookupErr == nil {
			return user, nil
		}
	}

	return nil, err
}

func (s *AuthService) createUser(ctx context.Context, phone string) (*domain.User, error) {
	guid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	user := &domain.User{
		GUID:     guid,
		Phone:    phone,
		IsActive: true,
	}

	if err = s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) FinalizeResetPasswordByOTP(
	ctx context.Context, sign, newPassword string,
) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	userUUID, err := s.sessionService.ValidateResetToken(ctx, sign)
	if err != nil {
		return err
	}

	return s.uow.Do(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.UpdatePassword(txCtx, userUUID, newPassword); err != nil {
			return err
		}

		return s.sessionService.RevokeAllSessions(txCtx, userUUID)
	})
}

func validatePassword(password string) error {
	const (
		minPassLength = 8
		maxPassLength = 64
	)

	if len(password) < minPassLength || len(password) > maxPassLength {
		return domain.ErrInvalidPassword
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSymbol := false

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSymbol {
		return domain.ErrInvalidPassword
	}

	return nil
}
