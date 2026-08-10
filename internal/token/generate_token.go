package token

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

const defaultTokenBytes = 32

func GenerateDefaultToken() (token string, err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("failed to generate token")
		}
	}()

	token = rand.Text()
	return
}

func GenerateToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
