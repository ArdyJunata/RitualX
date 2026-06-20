package middleware

import (
	"log/slog"
	"testing"
)

func TestSanitizeHeaders_RedactsAuthorization(t *testing.T) {
	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer eyJhbGciOiJIUzI1NiJ9.secret"},
	}
	result := sanitizeHeaders(headers)
	if result["Authorization"][0] != "Bearer [REDACTED]" {
		t.Errorf("got %q, want %q", result["Authorization"][0], "Bearer [REDACTED]")
	}
	if result["Content-Type"][0] != "application/json" {
		t.Errorf("Content-Type modified unexpectedly")
	}
}

func TestSanitizeHeaders_CaseInsensitive(t *testing.T) {
	headers := map[string][]string{
		"authorization": {"Bearer token123"},
	}
	result := sanitizeHeaders(headers)
	if result["authorization"][0] != "Bearer [REDACTED]" {
		t.Errorf("got %q, want %q", result["authorization"][0], "Bearer [REDACTED]")
	}
}

func TestSanitizeHeaders_NoAuthHeader(t *testing.T) {
	headers := map[string][]string{
		"Content-Type": {"application/json"},
	}
	result := sanitizeHeaders(headers)
	if len(result) != 1 || result["Content-Type"][0] != "application/json" {
		t.Errorf("headers modified unexpectedly")
	}
}

func TestCaptureBody_Empty(t *testing.T) {
	result := captureBody([]byte{})
	if result != "" {
		t.Errorf("got %q, want empty string", result)
	}
}

func TestCaptureBody_UnderLimit(t *testing.T) {
	body := []byte(`{"key":"value"}`)
	result := captureBody(body)
	if result != `{"key":"value"}` {
		t.Errorf("got %q, want %q", result, `{"key":"value"}`)
	}
}

func TestCaptureBody_OverLimit(t *testing.T) {
	body := make([]byte, 11000)
	for i := range body {
		body[i] = 'x'
	}
	result := captureBody(body)
	if len(result) != 10240+len("...[truncated]") {
		t.Errorf("got len %d, want %d", len(result), 10240+len("...[truncated]"))
	}
	if result[len(result)-14:] != "...[truncated]" {
		t.Errorf("missing truncation marker")
	}
}

func TestIsMultipart_True(t *testing.T) {
	if !isMultipart("multipart/form-data; boundary=----") {
		t.Error("expected true for multipart/form-data")
	}
}

func TestIsMultipart_False(t *testing.T) {
	if isMultipart("application/json") {
		t.Error("expected false for application/json")
	}
}

func TestRedactPasswords(t *testing.T) {
	input := `{"email":"a@b.com","password":"secret123"}`
	result := redactPasswords(input)
	expected := `{"email":"a@b.com","password":"[REDACTED]"}`
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestRedactPasswords_NoPassword(t *testing.T) {
	input := `{"email":"a@b.com"}`
	result := redactPasswords(input)
	if result != input {
		t.Errorf("body modified when no password present")
	}
}

func TestLogLevel_2xx(t *testing.T) {
	if logLevel(200) != slog.LevelInfo {
		t.Error("expected INFO for 200")
	}
	if logLevel(201) != slog.LevelInfo {
		t.Error("expected INFO for 201")
	}
}

func TestLogLevel_3xx(t *testing.T) {
	if logLevel(301) != slog.LevelInfo {
		t.Error("expected INFO for 301")
	}
}

func TestLogLevel_4xx(t *testing.T) {
	if logLevel(400) != slog.LevelWarn {
		t.Error("expected WARN for 400")
	}
	if logLevel(404) != slog.LevelWarn {
		t.Error("expected WARN for 404")
	}
}

func TestLogLevel_5xx(t *testing.T) {
	if logLevel(500) != slog.LevelError {
		t.Error("expected ERROR for 500")
	}
	if logLevel(503) != slog.LevelError {
		t.Error("expected ERROR for 503")
	}
}
