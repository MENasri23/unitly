package token

import (
	"crypto/rand"
	"encoding/hex"
	apprand "unitly/internal/util/rand"
)

const defaultTokenLen = 52 // ⌈log₃₂ 2^256⌉ = 52 chars

const base32alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

func GenerateDefaultToken() (string, error) {
	return apprand.TextN(defaultTokenLen)
}

func GenerateToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
