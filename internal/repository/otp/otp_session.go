package otprepo

import "strconv"

const (
	otpPhoneField    = "phone"
	otpHashField     = "otp_hash"
	otpAttemptsField = "attempts"
)

type otpSession struct {
	Phone    string
	otpHash  string
	Attempts int64
}

func (s *otpSession) serialize() map[string]any {
	return map[string]any{
		otpPhoneField:    s.Phone,
		otpHashField:     s.otpHash,
		otpAttemptsField: s.Attempts,
	}
}

func deserializeOtpSession(data map[string]string) (*otpSession, error) {
	attempts, err := strconv.ParseInt(data[otpAttemptsField], 10, 64)
	if err != nil {
		return nil, err
	}

	return &otpSession{
		Phone:    data[otpPhoneField],
		otpHash:  data[otpHashField],
		Attempts: attempts,
	}, nil
}
