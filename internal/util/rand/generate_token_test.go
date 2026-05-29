package rand

import (
	"encoding/base64"
	"strconv"
	"strings"
	"testing"
)

var validChars = "1234567890qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM_-"

func TestGenerateToken_ValidLength(t *testing.T) {
	nBytesSlice := []int{16, 32, 48, 64}

	for _, nBytes := range nBytesSlice {
		t.Run("bytes_"+strconv.FormatInt(int64(nBytes), 10), func(t *testing.T) {
			token, err := GenerateToken(nBytes)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			decodedBytes, err := decodeToken(token)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(decodedBytes) != nBytes {
				t.Errorf("expected number of bytes %d, got %d (%s)", nBytes, len(token), token)
			}

			for _, char := range token {
				if !strings.ContainsRune(validChars, char) {
					t.Errorf("token contains invalid character: %q", char)
				}
			}
		})
	}
}

func decodeToken(token string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(token)
}
