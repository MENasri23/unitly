package authsvc

import (
	"errors"
	"unitly/internal/util/rand"
)

var ErrInvalidLength = errors.New("invalid otp length")

func generateOTP(length int) (string, error) {
	if length <= 0 {
		return "", ErrInvalidLength
	}

	out := make([]byte, length)

	// First digit: 1-9
	if err := rand.GenerateCodeRange(out, 0, 1, 9); err != nil {
		return "", err
	}
	out[0] = out[0] + '1'

	// Remaining digits: 0-9
	if err := rand.GenerateCodeRange(out, 1, length, 10); err != nil {
		return "", nil
	}

	for i := 1; i < length; i++ {
		out[i] = out[i] + '0'
	}

	return string(out), nil
}
