package otpsvc

import "time"

type OtpConfig struct {
	TempID    string
	// RequestLimit is the maximum number of OTP requests allowed in the Window
	RequestLimit int64
	// Window is the duration for the rate limit (e.g. 1 * time.Minute)
	Window time.Duration
}
