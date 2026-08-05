package token

import (
	"crypto/rand"
	"encoding/hex"
)

const defaultTokenBytes = 32

func GenerateDefaultToken() (string, error) {
	return GenerateToken(defaultTokenBytes)
}

func GenerateToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
