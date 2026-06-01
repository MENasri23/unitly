package domain

type SMSService interface {
	SendOTP(phone string, code string) error
}
