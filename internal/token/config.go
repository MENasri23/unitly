package token

import "time"

type TokenConfig struct {
	AccessTokenSecret        string
	AccessTokenExpiry        time.Duration
	RefreshTokenExpiry       time.Duration
	ResetPasswordTokenExpiry time.Duration
}
