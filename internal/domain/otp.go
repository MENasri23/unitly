package domain

import "context"

var (
	ErrInvalidOTP     = NewError(ErrorCode("INVALID_OTP"), "invalid or expired OTP")
	ErrOTPAlreadyUsed = NewError(ErrorCode("OTP_ALREADY_USED"), "OTP has already been used")
)

type OTPPurpose string

type CreateOTPReq struct {
	Phone   string
	Token   string
	Purpose OTPPurpose
}

type VerifyOTPReq struct {
	Token    string
	InputOTP string
	Purpose  OTPPurpose
}

type VerifyOTPResult struct {
	Phone string
}

type OTPRepository interface {
	IsBlocked(ctx context.Context, phone string, purpose OTPPurpose) (bool, error)
	Create(ctx context.Context, req *CreateOTPReq) (string, error)
	Verify(ctx context.Context, verifyReq *VerifyOTPReq) (*VerifyOTPResult, error)
}

// OTPService handles rate limiting, lifecycle, and transmission of one-time tokens.
type OTPService interface {
	EnsureAllowed(ctx context.Context, phone string, purpose OTPPurpose) error
	RequestOTP(ctx context.Context, phone string, purpose OTPPurpose, executeDelivery bool) (string, error)
	VerifyOTP(ctx context.Context, token, inputOTP string, purpose OTPPurpose) (string, error)
}
