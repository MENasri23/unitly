package controller

import (
	"net/http"
	"unitly/internal/domain"
	"unitly/internal/service/identity"
)

var errToStatusCode = map[error]int{
	domain.ErrUserNotFound: http.StatusNotFound,
	domain.ErrNotFound:     http.StatusNotFound,

	domain.ErrTooManyAttempts:    http.StatusTooManyRequests,
	domain.ErrInvalidOTP:         http.StatusUnauthorized,
	domain.ErrInvalidToken:       http.StatusUnauthorized,
	domain.ErrInvalidPhone:       http.StatusBadRequest,
	domain.ErrPhoneUnchanged:     http.StatusConflict,
	domain.ErrPhoneAlreadyInUse:  http.StatusConflict,
	domain.ErrInvalidCredentials: http.StatusUnauthorized,
	identity.ErrEmptyFullName:    http.StatusBadRequest,
	domain.ErrInvalidPassword:    http.StatusBadRequest,
	domain.ErrIncorrectPassword:  http.StatusBadRequest,

	domain.ErrSessionInUse:     http.StatusBadRequest,
	domain.ErrSessionNotFound:  http.StatusNotFound,
	domain.ErrSessionForbidden: http.StatusForbidden,

	domain.ErrRateLimit:      http.StatusTooManyRequests,
	domain.ErrOTPAlreadyUsed: http.StatusBadRequest,
}

func mapErrToStatusCode(err error) int {
	if code, ok := errToStatusCode[err]; ok {
		return code
	}

	return http.StatusInternalServerError
}
