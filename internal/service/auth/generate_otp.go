package authsvc

import (
	"crypto/rand"
	"errors"
)

var maxOTPLength = 18

var ErrInvalidLength = errors.New("invalid otp length")

func generateOTP(length int) (string, error) {
	if length <= 0 {
		return "", ErrInvalidLength
	}

	out := make([]byte, length)
	buf := make([]byte, 1)

	// First digit: 1-9
	for {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}

		// 252 divisible by 9 to avoid module bias
		if buf[0] >= 252 {
			continue
		}

		out[0] = (buf[0] % 9) + '1'
		break
	}

	// Remaining digits: 0-9
	for i := 1; i < length; i++ {
		for {
			if _, err := rand.Read(buf); err != nil {
				return "", err
			}

			// 250 divisible by 10 to avoid module bias
			if buf[0] >= 250 {
				continue
			}

			out[i] = (buf[0] % 10) + '0'
			break
		}
	}

	return string(out), nil
}
