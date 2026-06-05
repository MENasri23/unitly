package controller

import (
	"errors"
	"net/http"
	"unitly/internal/domain"
)

var (
	ErrInvalidReqest = errors.New("Invalid request body")
)

func mapErrToStatusCode(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound

	case errors.Is(err, domain.ErrTooManyAttempts):
		return http.StatusTooManyRequests

	case errors.Is(err, domain.ErrInvalidOTP):
		return http.StatusUnauthorized

	case errors.Is(err, domain.ErrInvalidToken):
		return http.StatusUnauthorized

	case errors.Is(err, domain.ErrRateLimit):
		return http.StatusTooManyRequests

	case errors.Is(err, domain.ErrInvalidPhone):
		return http.StatusBadRequest

	case errors.Is(err, domain.ErrOTPAlreadyUsed):
		return http.StatusBadRequest

	default:
		return http.StatusInternalServerError
	}
}
