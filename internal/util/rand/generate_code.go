package rand

import (
	"crypto/rand"
	"errors"
	"fmt"
)

func GenerateCode(
	out []byte,
	start int,
	end int,
	max int,
) error {
	if start < 0 || end < start || end > len(out) {
		return fmt.Errorf(
			"invalid range: start=%d end=%d len=%d",
			start,
			end,
			len(out),
		)
	}

	if max < 1 || max > 255 {
		return errors.New("max must be between 1 and 255")
	}

	reject := byte(255 - (255 % max))
	buf := make([]byte, 1)
	maxByte := byte(max)

	for i := start; i < end; i++ {
		for {
			if _, err := rand.Read(buf); err != nil {
				return err
			}

			if buf[0] >= reject {
				continue
			}

			out[i] = buf[0] % maxByte
			break
		}
	}

	return nil
}
