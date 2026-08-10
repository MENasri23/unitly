package rand

import (
	"crypto/rand"
	"errors"
)

const base32alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

func TextN(n int) (t string, err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("failed to generate token")
		}
	}()

	src := make([]byte, n)
	rand.Read(src)
	for i := range src {
		src[i] = base32alphabet[src[i]%32]
	}
	t = string(src)
	return
}
