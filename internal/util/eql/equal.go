package eql

import "crypto/subtle"

func IsEqual(s1, s2 string) bool {
	return subtle.ConstantTimeCompare([]byte(s1), []byte(s2)) == 1
}
