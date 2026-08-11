package otprepo

import (
	"unitly/internal/domain"
	"unitly/internal/util/rand"
)

var ErrInvalidLength = domain.NewError("INVALID_LENGTH", "invalid otp length")

func generateOTP(length int) (string, error) {
	if length <= 0 {
		return "", ErrInvalidLength
	}

	out := make([]byte, length)

	if err := rand.GenerateCode(out, 10); err != nil {
		return "", err
	}

	for i := range length {
		out[i] = out[i] + '0'
	}

	return string(out), nil
}
