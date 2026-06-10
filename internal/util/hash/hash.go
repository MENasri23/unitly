package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
