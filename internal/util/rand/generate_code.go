package rand

import (
	"crypto/rand"
	"unitly/internal/domain"
)

var (
	ErrInvalidRange = domain.NewError(domain.ErrorCode("INVALID_RANGE"), "invalid range")
	ErrInvalidMax   = domain.NewError(domain.ErrorCode("INVALID_MAX"), "max must be between 1 and 255")
)

func GenerateCode(
	codes []byte,
	max int,
) error {
	return GenerateCodeRange(codes, 0, len(codes), max)
}

// GenerateCodeRange fills a portion of the provided byte slice with cryptographically
// secure random values in the range [0, max).
//
// The function writes into out[start:end] and ensures:
//
//   - All values are generated using crypto/rand (secure RNG)
//   - Values are uniformly distributed in [0, max)
//   - Modulo bias is avoided using rejection sampling
//   - The output slice is not reallocated (caller provides buffer)
//
// Parameters:
//
//   - out   : destination buffer where random values are written
//   - start : starting index (inclusive)
//   - end   : ending index (exclusive)
//   - max   : upper bound (exclusive) for generated values (must be 1–255)
//
// Returns an error if:
//   - start/end range is invalid
//   - max is outside allowed range
//   - crypto/rand fails
//
// Example usage:
//
//	out := make([]byte, 10)
//	err := GenerateCodeRange(out, 0, len(out), 10) // values in [0..9]
func GenerateCodeRange(
	codes []byte,
	start int,
	end int,
	max int,
) error {
	if start < 0 || end < start || end > len(codes) {
		return ErrInvalidRange
	}

	if max < 1 || max > 255 {
		return ErrInvalidMax
	}

	reject := byte(255 - (255 % max))
	maxByte := byte(max)
	buf := make([]byte, end-start)

	for i := start; i < end; {
		if _, err := rand.Read(buf); err != nil {
			return err
		}

		for _, b := range buf {
			if b >= reject {
				continue
			}

			codes[i] = b % maxByte
			i++

			if i == end {
				break
			}
		}
	}

	return nil
}
