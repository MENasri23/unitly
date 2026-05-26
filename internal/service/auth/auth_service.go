package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"unitly/internal/domain"
	"unitly/internal/service/sms"
)

type AuthService struct {
	otpRepo    domain.OTPRepository
	userRepo   domain.UserRepository
	smssvc     sms.
	jwtSecret  []byte
	hmacSecret []byte
}

func NewAuthService(
	otpRepo domain.OTPRepository,
	userRepo domain.UserRepository,
	smssvc smssvc.SMSService,
	jwtSec string,
	hmacSec string,
) *AuthService {
	return &AuthService{
		otpRepo:    otpRepo,
		userRepo:   userRepo,
		smssvc:     smssvc,
		jwtSecret:  []byte(jwtSec),
		hmacSecret: []byte(hmacSec),
	}
}

// hashOTP creates an HMAC-SHA256 hash of the OTP for secure storage
func (s *AuthService) hashOTP(otp string) string {
	h := hmac.New(sha256.New, s.hmacSecret)
	h.Write([]byte(otp))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *AuthService) RequestOTP(ctx context.Context, phone string) (string, error) {

	otp, err := generateOTP(6)
	if err != nil {
		return "", err
	}

	return "", nil
}
