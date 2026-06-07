package rand

import (
	"testing"
)

// Test valid generation fills the correct range
func TestGenerateCode_ValidRange(t *testing.T) {
	out := make([]byte, 10)

	err := GenerateCodeRange(out, 2, 8, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ensure only requested range is modified
	for i := range out {
		if i >= 2 && i < 8 {
			if out[i] >= 10 {
				t.Errorf("value out of range at %d: %d", i, out[i])
			}
		}
	}
}

// Test full slice generation
func TestGenerateCode_FullRange(t *testing.T) {
	out := make([]byte, 100)

	err := GenerateCodeRange(out, 0, len(out), 255)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, v := range out {
		if v == 255 {
			t.Errorf("invalid value at %d: %d", i, v)
		}
	}
}

// Test invalid range: start < 0
func TestGenerateCode_InvalidStart(t *testing.T) {
	out := make([]byte, 10)

	err := GenerateCodeRange(out, -1, 5, 10)
	if err != ErrInvalidRange {
		t.Fatal("expected error for negative start")
	}
}

// Test invalid range: end > len(out)
func TestGenerateCode_InvalidEnd(t *testing.T) {
	out := make([]byte, 10)

	err := GenerateCodeRange(out, 0, 11, 10)
	if err != ErrInvalidRange {
		t.Fatal("expected error for end out of bounds")
	}
}

// Test invalid range: end < start
func TestGenerateCode_EndLessThanStart(t *testing.T) {
	out := make([]byte, 10)

	err := GenerateCodeRange(out, 5, 2, 10)
	if err != ErrInvalidRange {
		t.Fatal("expected error when end < start")
	}
}

// Test invalid max values
func TestGenerateCode_InvalidMax(t *testing.T) {
	out := make([]byte, 10)

	tests := []int{0, -1, 256}

	for _, max := range tests {
		err := GenerateCodeRange(out, 0, 5, max)
		if err != ErrInvalidMax {
			t.Errorf("expected error for max=%d", max)
		}
	}
}

// Smoke test: ensure deterministic bounds (not randomness quality test)
func TestGenerateCode_BoundsOnly(t *testing.T) {
	out := make([]byte, 1000)

	err := GenerateCode(out, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, v := range out {
		if v >= 42 {
			t.Fatalf("out of range value at %d: %d", i, v)
		}
	}
}
