package domain

import "context"

type SmsService interface {
	SendOTP(ctx context.Context, phone string, code string) error
}
