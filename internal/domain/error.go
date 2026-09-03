// domain/errors.go
package domain

import (
	"fmt"
	"strings"
)

type ErrorCode string

// Domain Error Interface
type Error interface {
	error
	Code() ErrorCode
	Params() map[string]any
}

type errorCode struct {
	code    ErrorCode
	message string
	params  map[string]any
}

func (e *errorCode) Error() string {
	return e.message
}

func (e *errorCode) Code() ErrorCode {
	return e.code
}

func (e *errorCode) Params() map[string]any {
	return e.params
}

// Sentinel Errors
var (
	ErrInternal           = NewError("INTERNAL", "something went wrong")
	ErrAlreadyExist       = NewError("ALREADY_EXIST", "record already exist")
	ErrInvalidToken       = NewError("INVALID_TOKEN", "invalid or expired token")
	ErrInvalidPhone       = NewError("INVALID_PHONE", "invalid phone number")
	ErrPhoneUnchanged     = NewError("PHONE_UNCHANGED", "phone number must be different from the current one")
	ErrPhoneAlreadyInUse  = NewError("PHONE_ALREADY_IN_USE", "this phone number is already in use")
	ErrInvalidPassword    = NewError("INVALID_PASSWORD", "invalid password")
	ErrIncorrectPassword  = NewError("INCORRECT_PASSWORD", "incorrect password")
	ErrInvalidCredentials = NewError("INVALID_CREDENTIALS", "invalid username or password")
	ErrRateLimit          = NewError("TOO_MANY_REQUESTS", "too many request")
	ErrTooManyAttempts    = NewError("TOO_MANY_ATTEMPTS", "too many failed attempts, please try again later")
	ErrNotFound           = NewNotFoundError("record")
)

func NewError(code ErrorCode, msg string) Error {
	return &errorCode{code: code, message: msg}
}

// NewErrorWithParams creates a domain error with dynamic template variables
func NewErrorWithParams(code ErrorCode, msg string, params map[string]any) Error {
	return &errorCode{code: code, message: msg, params: params}
}

func NewNotFoundError(entity string) Error {
	entityName := strings.TrimSpace(strings.ToLower(entity))
	if entityName == "" {
		entityName = "record"
	}

	code := ErrorCode(strings.ToUpper(entityName) + "_NOT_FOUND")
	return &errorCode{
		code:    code,
		message: fmt.Sprintf("%s not found", entityName),
		params:  map[string]any{"entity": entityName},
	}
}

func NewConflictError(entity string) Error {
	entityName := strings.TrimSpace(strings.ToLower(entity))
	if entityName == "" {
		entityName = "record"
	}

	code := ErrorCode(strings.ToUpper(entityName) + "_CONFLICT")
	return &errorCode{
		code:    code,
		message: fmt.Sprintf("%s conflict", entityName),
		params:  map[string]any{"entity": entityName},
	}
}
