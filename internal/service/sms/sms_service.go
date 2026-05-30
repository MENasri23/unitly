package sms

type SMSService interface {
	SendOTP(phone string, code string) error
}
