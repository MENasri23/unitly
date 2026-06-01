package domain

type SmsService interface {
	SendOTP(phone string, code string) error
}
