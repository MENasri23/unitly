package controller

import (
	"net/http"
	"unitly/internal/domain"
)

var errToStatusCode = map[error]int{
	domain.ErrUserNotFound: http.StatusNotFound,
	domain.ErrNotFound:     http.StatusNotFound,

	domain.ErrTooManyAttempts:    http.StatusTooManyRequests,
	domain.ErrInvalidOTP:         http.StatusUnauthorized,
	domain.ErrInvalidToken:       http.StatusUnauthorized,
	domain.ErrInvalidPhone:       http.StatusBadRequest,
	domain.ErrInvalidCredentials: http.StatusUnauthorized,

	domain.ErrRateLimit:      http.StatusTooManyRequests,
	domain.ErrOTPAlreadyUsed: http.StatusBadRequest,
}

func mapErrToStatusCode(err error) int {
	if code, ok := errToStatusCode[err]; ok {
		return code
	}

	return http.StatusInternalServerError
}
