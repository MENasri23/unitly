package domain

import (
	"testing"
)

func TestNewErrorExposesCode(t *testing.T) {
	err := NewError(ErrorCode("USER_NOT_FOUND"), "user not found")
	codedErr, ok := err.(Error)
	if !ok {
		t.Fatalf("expected domain.Error implementation, got %T", err)
	}

	if codedErr.Code() != ErrorCode("USER_NOT_FOUND") {
		t.Fatalf("expected USER_NOT_FOUND code, got %q", codedErr.Code())
	}
}
