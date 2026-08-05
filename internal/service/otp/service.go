package otpsvc

import (
	"context"
	"fmt"
	"log"
	"time"

	"unitly/internal/domain"
	"unitly/internal/token"
)

type otpService struct {
	otpRepo        domain.OTPRepository
	smsService     domain.SmsService
	otpRateLimiter domain.RateLimiter
	sendTimeout    time.Duration
}

func NewOTPService(otpRepo domain.OTPRepository, smsService domain.SmsService, rateLimiter domain.RateLimiter) domain.OTPService {
	return &otpService{
		otpRepo:        otpRepo,
		smsService:     smsService,
		otpRateLimiter: rateLimiter,
		sendTimeout:    15 * time.Second,
	}
}

func (s *otpService) EnsureAllowed(ctx context.Context, phone string, purpose domain.OTPPurpose) error {
	blocked, err := s.otpRepo.IsBlocked(ctx, phone, purpose)
	if err != nil || blocked {
		return domain.ErrTooManyAttempts
	}

	key := ratelimitKey(phone, purpose)
	allowed, err := s.otpRateLimiter.Allow(ctx, key)
	if err != nil || !allowed {
		return domain.ErrRateLimit
	}

	return nil
}

func (s *otpService) RequestOTP(ctx context.Context, phone string, purpose domain.OTPPurpose, executeDelivery bool) (string, error) {
	token, err := token.GenerateDefaultToken()
	if err != nil {
		return "", err
	}

	otpCode, err := s.otpRepo.Create(ctx, &domain.CreateOTPReq{
		Phone:   phone,
		Token:   token,
		Purpose: purpose,
	})
	if err != nil {
		s.otpRateLimiter.Reset(ctx, ratelimitKey(phone, purpose))
		return "", err
	}

	if executeDelivery {
		s.sendOtpAsync(ctx, phone, otpCode)
	}

	return token, nil
}

func ratelimitKey(phone string, purpose domain.OTPPurpose) string {
	return fmt.Sprintf("auth:%s:otp:%s", purpose, phone)
}

func (s *otpService) VerifyOTP(ctx context.Context, token, inputOTP string, purpose domain.OTPPurpose) (string, error) {
	result, err := s.otpRepo.Verify(ctx, &domain.VerifyOTPReq{
		Token:    token,
		InputOTP: inputOTP,
		Purpose:  purpose,
	})
	if err != nil {
		return "", domain.ErrInvalidOTP
	}
	return result.Phone, nil
}

func (s *otpService) sendOtpAsync(ctx context.Context, phone, otp string) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.sendTimeout)
		defer cancel()
		if err := s.smsService.SendOTP(bgCtx, phone, otp); err != nil {
			log.Printf("failed to send OTP to %s: %v", phone, err)
		}
	}()
}
