package domain

import (
	"context"
	"errors"
)

var (
	ErrInvalidOTP     = errors.New("invalid or expired OTP")
	ErrOTPAlreadyUsed = errors.New("OTP has already been used")
)

type OTPPurpose string

type VerifyOTPReq struct {
	Phone    string
	Token    string
	InputOTP string
	Purpose  OTPPurpose
}

type OTPRepository interface {
	IsBlocked(ctx context.Context, phone string, purpose OTPPurpose) (bool, error)
	Create(ctx context.Context, phone string, purpose OTPPurpose) (string, string, error)
	GetPhoneByToken(ctx context.Context, token string, purpose OTPPurpose) (string, error)
	Verify(ctx context.Context, verifyReq *VerifyOTPReq) error
}
