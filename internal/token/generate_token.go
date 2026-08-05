package token

import (
	"crypto/rand"
	"encoding/base64"
)

const defaultTokenBytes = 48

func GenerateDefaultToken() (string, error) {
	return GenerateToken(defaultTokenBytes)
}

func GenerateToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
