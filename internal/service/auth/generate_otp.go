package authsvc

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strconv"
)

var maxOTPLength = 18

var ErrInvalidLength = errors.New("invalid otp length")

func generateOTP(length int) (string, error) {
	if length <= 0 || length > maxOTPLength {
		return "", ErrInvalidLength
	}

	max := pow10(length)
	min := max / 10

	n, err := rand.Int(rand.Reader, big.NewInt(max-min))
	if err != nil {
		return "", err
	}

	val := n.Int64() + min

	return strconv.FormatInt(val, 10), nil
}

func pow10(n int) int64 {
	v := int64(1)
	for range n {
		v *= 10
	}

	return v
}
