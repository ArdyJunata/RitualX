package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestInit_SetsLogLevel(t *testing.T) {
	Init("warn")
	l := Get()
	if l == nil {
		t.Fatal("Get() returned nil after Init()")
	}
	if !l.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("expected WARN level to be enabled")
	}
	if l.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("expected INFO level to be disabled at WARN threshold")
	}
}

func TestInit_InvalidLevelDefaultsToInfo(t *testing.T) {
	Init("invalid")
	l := Get()
	if !l.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("expected INFO level enabled for invalid input")
	}
	if l.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("expected DEBUG disabled when defaulting to INFO")
	}
}

func TestGet_BeforeInit_ReturnsFallback(t *testing.T) {
	defaultLogger = nil
	l := Get()
	if l == nil {
		t.Fatal("Get() returned nil before Init()")
	}
}

func TestFromContext_WithTraceID(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	defaultLogger = slog.New(handler)

	ctx := WithTraceID(context.Background(), "test-trace-123")
	l := FromContext(ctx)
	l.Info("test message")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log JSON: %v", err)
	}
	if entry["trace_id"] != "test-trace-123" {
		t.Errorf("trace_id = %v, want %q", entry["trace_id"], "test-trace-123")
	}
}

func TestFromContext_WithoutTraceID(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	defaultLogger = slog.New(handler)

	l := FromContext(context.Background())
	l.Info("no trace")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log JSON: %v", err)
	}
	if _, exists := entry["trace_id"]; exists {
		t.Error("expected no trace_id in log entry")
	}
}
