package domain

import "errors"

var (
	ErrNotFound        = errors.New("record not found")
	ErrRateLimit       = errors.New("too many request")
	ErrInvalidToken    = errors.New("invalid or expired token")
	ErrTooManyAttempts = errors.New("too many failed attempts, please try again later")
)
