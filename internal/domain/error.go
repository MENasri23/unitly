package domain

import "errors"

var (
	ErrInternal        = errors.New("something went wrong")
	ErrNotFound        = errors.New("record not found")
	ErrRateLimit       = errors.New("too many request")
	ErrInvalidToken    = errors.New("invalid or expired token")
	ErrTooManyAttempts = errors.New("too many failed attempts, please try again later")
	ErrInvalidPhone    = errors.New("invalid phone number")
	ErrAleardyExist    = errors.New("record already exist")
)
