package controller

import (
	"net/http"
	"unitly/internal/domain"
)

func mapErrToStatusCode(err error) int {
	switch err {
	case domain.ErrNotFound:
		return http.StatusNotFound

	case domain.ErrTooManyAttempts:
		return http.StatusTooManyRequests

	case domain.ErrInvalidOTP:
		return http.StatusUnauthorized

	case domain.ErrInvalidToken:
		return http.StatusUnauthorized

	case domain.ErrRateLimit:
		return http.StatusTooManyRequests

	case domain.ErrInvalidPhone:
		return http.StatusBadRequest

	case domain.ErrOTPAlreadyUsed:
		return http.StatusBadRequest

	default:
		return http.StatusInternalServerError
	}
}
