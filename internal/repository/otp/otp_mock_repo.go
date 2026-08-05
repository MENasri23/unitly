package otprepo

import (
	"context"
	"unitly/internal/domain"
)

type otpMockRepo struct {
	originalRepo *otpRepository
}

func NewOTPMockRepo(cache domain.Cache) domain.OTPRepository {
	return &otpMockRepo{
		originalRepo: NewOtpRepository(cache).(*otpRepository),
	}
}

func (o *otpMockRepo) IsBlocked(ctx context.Context, phone string, purpose domain.OTPPurpose) (bool, error) {
	return o.originalRepo.IsBlocked(ctx, phone, purpose)
}

func (o *otpMockRepo) Create(ctx context.Context, req *domain.CreateOTPReq) (string, error) {
	otp := "123456" // Mock OTP for testing

	if err := o.originalRepo.storeOTP(ctx, req, otp); err != nil {
		return "", err
	}

	return otp, nil
}

func (o *otpMockRepo) Verify(ctx context.Context, verifyReq *domain.VerifyOTPReq) (*domain.VerifyOTPResult, error) {
	return o.originalRepo.Verify(ctx, verifyReq)
}
