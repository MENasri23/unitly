package domain

import (
	"fmt"
	"strings"
)

var (
	ErrInternal     = NewError(ErrorCode("INTERNAL"), "something went wrong")
	ErrAlreadyExist = NewError(ErrorCode("ALREADY_EXIST"), "record already exist")

	ErrNotFound = NewNotFoundError("record")

	ErrInvalidToken       = NewError(ErrorCode("INVALID_TOKEN"), "invalid or expired token")
	ErrInvalidPhone       = NewError(ErrorCode("INVALID_PHONE"), "invalid phone number")
	ErrInvalidPassword    = NewError(ErrorCode("INVALID_PASSWORD"), "invalid password")
	ErrInvalidCredentials = NewError(ErrorCode("INVALID_CREDENTIALS"), "invalid username or password")

	ErrRateLimit       = NewError(ErrorCode("TOO_MANY_REQUESTS"), "too many request")
	ErrTooManyAttempts = NewError(ErrorCode("TOO_MANY_ATTEMPTS"), "too many failed attempts, please try again later")
)

type ErrorCode string

type Error interface {
	error
	Code() ErrorCode
}

type errorCode struct {
	code    ErrorCode
	message string
}

func (e *errorCode) Error() string {
	return e.message
}

func (e *errorCode) Code() ErrorCode {
	return e.code
}

func NewError(code ErrorCode, msg string) error {
	return &errorCode{code: code, message: msg}
}

func NewNotFoundError(entity string) error {
	entityName := strings.TrimSpace(strings.ToLower(entity))
	if entityName == "" {
		entityName = "record"
	}

	code := ErrorCode(strings.ToUpper(entityName) + "_NOT_FOUND")
	return &errorCode{code: code, message: fmt.Sprintf("%s not found", entityName)}
}
