package auth

import (
	"context"
	"errors"
	"time"
	"unitly/internal/cache"
	"unitly/internal/domain"
	"unitly/internal/service/sms"
	"unitly/internal/util/validate"

	"github.com/google/uuid"
)

const jwtExpiry = 15 * time.Minute

var (
	ErrInvalidPhone = errors.New("invalid phone number")
)

var (
	authPurpose = domain.OTPPurpose("register-login")
)

type AuthService struct {
	otpRepo      domain.OTPRepository
	userRepo     domain.UserRepository
	smssvc       sms.SMSService
	tokenService domain.TokenService
}

func NewAuthService(
	otpRepo domain.OTPRepository,
	userRepo domain.UserRepository,
	smssvc sms.SMSService,
	cache cache.Cache,
	jwtSec string,
) *AuthService {
	tokenService := NewJWTService([]byte(jwtSec), jwtExpiry)
	return &AuthService{
		otpRepo:      otpRepo,
		userRepo:     userRepo,
		smssvc:       smssvc,
		tokenService: tokenService,
	}
}

func (s *AuthService) RequestAuthByOTP(
	ctx context.Context,
	phone string,
) (string, error) {
	return s.requestOTP(ctx, phone, authPurpose)
}

func (s *AuthService) VerifyAuthByOTP(
	ctx context.Context,
	token string,
	inputOTP string,
) (string, error) {
	return s.verifyOTP(ctx, token, inputOTP, authPurpose)
}

func (s *AuthService) requestOTP(
	ctx context.Context,
	phone string,
	purpose domain.OTPPurpose,
) (string, error) {
	if !s.isValidPhone(phone) {
		return "", ErrInvalidPhone
	}

	blocked, err := s.otpRepo.IsBlocked(ctx, phone, purpose)
	if err != nil {
		return "", err
	}

	if blocked {
		return "", domain.ErrTooManyAttempts
	}

	ratelimited, err := s.otpRepo.IsRateLimited(ctx, phone, purpose)
	if err != nil {
		return "", err
	}

	if ratelimited {
		return "", domain.ErrRateLimit
	}

	otp, token, err := s.otpRepo.Create(ctx, phone, purpose)
	if err != nil {
		s.otpRepo.DeleteRateLimit(ctx, phone, purpose)
		return "", err
	}

	s.smssvc.SendOTP(phone, otp)

	return token, nil
}

func (s *AuthService) verifyOTP(
	ctx context.Context,
	token string,
	inputOTP string,
	purpose domain.OTPPurpose,
) (string, error) {
	phone, err := s.otpRepo.GetPhoneByToken(ctx, token, purpose)
	if err != nil {
		return "", domain.ErrInvalidToken
	}

	blocked, err := s.otpRepo.IsBlocked(ctx, phone, purpose)
	if err != nil {
		return "", err
	}

	if blocked {
		return "", domain.ErrTooManyAttempts
	}

	req := &domain.VerifyOTPReq{
		Phone:    phone,
		Token:    token,
		InputOTP: inputOTP,
		Purpose:  purpose,
	}
	if err = s.otpRepo.Verify(ctx, req); err != nil {
		return "", err
	}

	userID, err := s.userRepo.GetUserIDByPhone(ctx, phone)
	if err != nil {
		if isNotFound(err) {
			return s.registerUser(ctx, phone)
		}

		return "", err
	}

	return s.generateToken(ctx, userID)
}

func (s *AuthService) isValidPhone(phone string) bool {
	return validate.IsValidPhone(phone)
}

func isNotFound(err error) bool {
	return errors.Is(err, domain.ErrNotFound)
}

func (s *AuthService) registerUser(ctx context.Context, phone string) (string, error) {
	user, err := s.createUser(ctx, phone)
	if err != nil {
		return "", err
	}
	return s.generateToken(ctx, user.ID)
}

func (s *AuthService) createUser(ctx context.Context, phone string) (*domain.User, error) {
	guid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	user := &domain.User{
		GUID:  guid,
		Phone: phone,
	}

	if err = s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) generateToken(ctx context.Context, userID int64) (string, error) {
	claims := domain.TokenClaims{
		UserID: userID,
		Expiry: jwtExpiry,
	}

	return s.tokenService.GenerateToken(ctx, claims)
}
