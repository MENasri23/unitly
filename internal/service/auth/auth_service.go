package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
	"unitly/internal/domain"
	"unitly/internal/util/validate"

	"github.com/google/uuid"
)

const (
	sendOtpTimeout = 15 * time.Second
)

var (
	authPurpose = domain.OTPPurpose("register-login")
)

type AuthService struct {
	otpRepo        domain.OTPRepository
	userRepo       domain.UserRepository
	smssvc         domain.SmsService
	tokenService   domain.TokenService
	otpRateLimiter domain.RateLimiter
}

func NewAuthService(
	otpRepo domain.OTPRepository,
	userRepo domain.UserRepository,
	smssvc domain.SmsService,
	otpRateLimiter domain.RateLimiter,
	accessTokenSecret string,
	accessTokenExpiry time.Duration,
) *AuthService {
	tokenService := NewJWTService([]byte(accessTokenSecret), accessTokenExpiry)
	return &AuthService{
		otpRepo:        otpRepo,
		userRepo:       userRepo,
		smssvc:         smssvc,
		tokenService:   tokenService,
		otpRateLimiter: otpRateLimiter,
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
		return "", domain.ErrInvalidPhone
	}

	blocked, err := s.otpRepo.IsBlocked(ctx, phone, purpose)
	if err != nil {
		return "", err
	}

	if blocked {
		return "", domain.ErrTooManyAttempts
	}

	if err = s.checkRatelimit(ctx, phone, purpose); err != nil {
		return "", err
	}

	otp, token, err := s.otpRepo.Create(ctx, phone, purpose)
	if err != nil {
		s.resetRatelimit(ctx, phone, purpose)
		return "", err
	}

	s.sendOtpAsync(ctx, phone, otp)

	return token, nil
}

func (s *AuthService) checkRatelimit(ctx context.Context, phone string, purpose domain.OTPPurpose) error {
	ratelimitAllowed, err := s.otpRateLimiter.Allow(ctx, ratelimitKey(phone, purpose))
	if err != nil {
		return err
	}

	if !ratelimitAllowed {
		return domain.ErrRateLimit
	}

	return nil
}

func (s *AuthService) resetRatelimit(ctx context.Context, phone string, purpose domain.OTPPurpose) {
	s.otpRateLimiter.Reset(ctx, ratelimitKey(phone, purpose))
}

func (s *AuthService) sendOtpAsync(
	ctx context.Context,
	phone string,
	otp string,
) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), sendOtpTimeout)
		defer cancel()
		if err := s.smssvc.SendOTP(bgCtx, phone, otp); err != nil {
			log.Printf("failed to send OTP to %s: %v", phone, err)
		}
	}()
}

func (s *AuthService) verifyOTP(
	ctx context.Context,
	token string,
	inputOTP string,
	purpose domain.OTPPurpose,
) (string, error) {
	req := &domain.VerifyOTPReq{
		Token:    token,
		InputOTP: inputOTP,
		Purpose:  purpose,
	}
	
	result, err := s.otpRepo.Verify(ctx, req)
	if err != nil {
		return "", err
	}

	userID, err := s.userRepo.GetUserIDByPhone(ctx, result.Phone)
	if err != nil {
		if isNotFound(err) {
			return s.registerUser(ctx, result.Phone)
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
	claims := domain.TokenClaims{UserID: userID}
	return s.tokenService.GenerateToken(ctx, claims)
}

func ratelimitKey(phone string, purpose domain.OTPPurpose) string {
	return fmt.Sprintf("auth:%s:otp:ratelimit:%s", purpose, phone)
}
