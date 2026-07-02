package jsonutil

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"unitly/internal/domain"
)

func TestWriteErrJSONIncludesCodeAndMessage(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := WriteErrJSON(recorder, 400, domain.NewError(domain.ErrorCode("INVALID_PASSWORD"), "password is too weak"))
	if err != nil {
		t.Fatalf("WriteErrJSON returned error: %v", err)
	}

	var payload map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if payload["code"] != "INVALID_PASSWORD" {
		t.Fatalf("expected code %q, got %q", "INVALID_PASSWORD", payload["code"])
	}

	if payload["message"] != "password is too weak" {
		t.Fatalf("expected message %q, got %q", "password is too weak", payload["message"])
	}
}
