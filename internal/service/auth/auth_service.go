package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
	"unitly/internal/domain"
	"unitly/internal/util/rand"
	"unitly/internal/util/validate"

	"github.com/google/uuid"
)

const (
	sendOtpTimeout      = 15 * time.Second
	resetPasswordExpiry = 5 * time.Minute
)

var (
	authPurpose      = domain.OTPPurpose("register-login")
	resetPassPurpose = domain.OTPPurpose("password-reset")

	tokenTypeAccess    = domain.TokenType("access")
	tokenTypeResetPass = domain.TokenType("reset-pass")
)

type AuthService struct {
	otpRepo        domain.OTPRepository
	userRepo       domain.UserRepository
	smsService     domain.SmsService
	tokenService   domain.TokenService
	otpRateLimiter domain.RateLimiter
}

func NewAuthService(
	otpRepo domain.OTPRepository,
	userRepo domain.UserRepository,
	smsService domain.SmsService,
	tokenService domain.TokenService,
	otpRateLimiter domain.RateLimiter,
) *AuthService {
	return &AuthService{
		otpRepo:        otpRepo,
		userRepo:       userRepo,
		smsService:     smsService,
		tokenService:   tokenService,
		otpRateLimiter: otpRateLimiter,
	}
}

func (s *AuthService) RequestAuthByOTP(
	ctx context.Context,
	phone string,
) (string, error) {
	if !s.isValidPhone(phone) {
		return "", domain.ErrInvalidPhone
	}

	if err := s.ensureOTPRequestAllowed(ctx, phone, authPurpose); err != nil {
		return "", err
	}

	return s.requestOTP(ctx, phone, authPurpose)
}

func (s *AuthService) VerifyAuthByOTP(
	ctx context.Context,
	token string,
	inputOTP string,
) (string, error) {
	req := &domain.VerifyOTPReq{
		Token:    token,
		InputOTP: inputOTP,
		Purpose:  authPurpose,
	}

	result, err := s.otpRepo.Verify(ctx, req)
	if err != nil {
		return "", err
	}

	user, err := s.userRepo.GetByPhone(ctx, result.Phone)
	if err != nil {
		if isNotFound(err) {
			return s.registerUser(ctx, result.Phone)
		}

		return "", err
	}

	return s.generateToken(ctx, user.GUID)
}

func (s *AuthService) RequestResetPasswordByOTP(
	ctx context.Context,
	phone string,
) (string, error) {
	if !validate.IsValidPhone(phone) {
		return "", domain.ErrInvalidPhone
	}

	if err := s.ensureOTPRequestAllowed(ctx, phone, resetPassPurpose); err != nil {
		return "", err
	}

	_, err := s.userRepo.GetActiveUserIDByPhone(ctx, phone)
	if err == nil {
		return s.requestOTP(ctx, phone, resetPassPurpose)
	}

	// If the phone is not associated with an active user, do not reveal that fact.
	// Create an OTP record so timing/behaviour is consistent, but do not send an SMS.
	if !isNotFound(err) {
		return "", err
	}

	token, err := rand.GenereteDefaultToken()
	if err != nil {
		return "", err
	}

	req := &domain.CreateOTPReq{
		Phone:   phone,
		Token:   token,
		Purpose: resetPassPurpose,
	}

	if _, err = s.otpRepo.Create(ctx, req); err != nil {
		s.reset(ctx, phone, resetPassPurpose)
		return "", err
	}

	// Return the generated token (same shape as success) without indicating existence.
	return token, nil
}

func (s *AuthService) VerifyResetPasswordByOTP(
	ctx context.Context,
	token string,
	inputOTP string,
) (string, error) {
	req := &domain.VerifyOTPReq{
		Token:    token,
		InputOTP: inputOTP,
		Purpose:  resetPassPurpose,
	}

	result, err := s.otpRepo.Verify(ctx, req)
	if err != nil {
		return "", err
	}

	user, err := s.userRepo.GetByPhone(ctx, result.Phone)
	if err != nil {
		// Do not reveal whether the phone maps to a user. Return a generic OTP error for not-found.
		if isNotFound(err) {
			return "", domain.ErrInvalidOTP
		}
		return "", err
	}

	if !user.IsActive {
		return "", domain.ErrInvalidOTP
	}

	return s.generateResetPassToken(ctx, user.GUID)
}

func (s *AuthService) requestOTP(
	ctx context.Context,
	phone string,
	purpose domain.OTPPurpose,
) (string, error) {
	token, err := rand.GenereteDefaultToken()
	if err != nil {
		return "", err
	}

	req := &domain.CreateOTPReq{
		Phone:   phone,
		Token:   token,
		Purpose: purpose,
	}
	otp, err := s.otpRepo.Create(ctx, req)
	if err != nil {
		s.reset(ctx, phone, purpose)
		return "", err
	}

	s.sendOtpAsync(ctx, phone, otp)

	return token, nil
}

func (s *AuthService) ensureOTPRequestAllowed(
	ctx context.Context,
	phone string,
	purpose domain.OTPPurpose,
) error {
	blocked, err := s.otpRepo.IsBlocked(ctx, phone, purpose)
	if err != nil {
		return err
	}

	if blocked {
		return domain.ErrTooManyAttempts
	}

	if err = s.checkRatelimit(ctx, phone, purpose); err != nil {
		return err
	}

	return nil
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

func (s *AuthService) reset(ctx context.Context, phone string, purpose domain.OTPPurpose) {
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
		if err := s.smsService.SendOTP(bgCtx, phone, otp); err != nil {
			log.Printf("failed to send OTP to %s: %v", phone, err)
		}
	}()
}

func (s *AuthService) isValidPhone(phone string) bool {
	return validate.IsValidPhone(phone)
}

func isNotFound(err error) bool {
	return errors.Is(err, domain.ErrNotFound)
}

func (s *AuthService) registerUser(ctx context.Context, phone string) (string, error) {
	user, err := s.createUser(ctx, phone)
	if err == nil {
		return s.generateToken(ctx, user.GUID)
	}

	// Possible race condition: user may have been created by another goroutine.
	// Attempt to retrieve the user. If successful, proceed; otherwise return original error.
	if err == domain.ErrAleardyExist {
		user, lookupErr := s.userRepo.GetByPhone(ctx, phone)
		if lookupErr == nil {
			return s.generateToken(ctx, user.GUID)
		}
	}

	return "", err
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

func (s *AuthService) generateToken(ctx context.Context, userID uuid.UUID) (string, error) {
	claims := domain.AccessTokenClaims{
		UserID: userID.String(),
		TokenClaims: domain.TokenClaims{
			TokenType: tokenTypeAccess,
		},
	}
	return s.tokenService.GenerateAccessToken(ctx, claims)
}

func (s *AuthService) generateResetPassToken(ctx context.Context, userID uuid.UUID) (string, error) {
	claims := domain.AccessTokenClaims{
		UserID: userID.String(),
		TokenClaims: domain.TokenClaims{
			Expiry:    resetPasswordExpiry,
			TokenType: tokenTypeResetPass,
		},
	}
	return s.tokenService.GenerateAccessToken(ctx, claims)
}

func ratelimitKey(phone string, purpose domain.OTPPurpose) string {
	return fmt.Sprintf("auth:%s:otp:ratelimit:%s", purpose, phone)
}
