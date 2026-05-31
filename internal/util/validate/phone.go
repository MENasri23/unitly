package validate

import (
	"regexp"
	"strings"
)

var (
	re          = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
	iranPhoneRe = regexp.MustCompile(`^\+98\d{10}$`)
)

// IsValidPhone checks if the provided phone number is in a valid international format(E.164).
func IsValidPhone(phone string) bool {
	if strings.HasPrefix(phone, "+98") {
		return iranPhoneRe.MatchString(phone)
	}

	return re.MatchString(phone)
}
