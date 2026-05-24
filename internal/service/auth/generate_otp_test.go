package authsvc

import (
	"strconv"
	"testing"
	"unicode"
)

func TestGenerateOTP_ValidLengths(t *testing.T) {
	lengths := []int{1, 2, 4, 5, 6, 10, 18}

	for _, length := range lengths {
		t.Run("length_"+strconv.FormatInt(int64(length), 10), func(t *testing.T) {
			otp, err := generateOTP(length)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(otp) != length {
				t.Errorf("expected length %d, got %d (%s)", length, len(otp), otp)
			}

			for _, r := range otp {
				if !unicode.IsDigit(r) {
					t.Errorf("otp contains non-digit character: %q", r)
				}
			}
		})
	}
}

func TestGenerateOTP_InvalidLengths(t *testing.T) {
	tests := []int{0, -1, 19, 100}

	for _, length := range tests {
		t.Run("invalid_"+strconv.FormatInt(int64(length), 10), func(t *testing.T) {
			otp, err := generateOTP(length)

			if err == nil {
				t.Errorf("expected error for length %d, got nil (otp=%s)", length, otp)
			}

			if err != ErrInvalidLength {
				t.Errorf("expected ErrInvalidLength, got %v", err)
			}
		})
	}
}

func TestPow10(t *testing.T) {
	tests := []struct {
		input    int
		expected int64
	}{
		{1, 10},
		{2, 100},
		{3, 1000},
		{6, 1000000},
		{18, 1000000000000000000},
	}

	for _, tt := range tests {
		got := pow10(tt.input)
		if got != tt.expected {
			t.Errorf("pow10(%d) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}
