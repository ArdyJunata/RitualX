package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/ArdyJunata/RitualX/backend/internal/logger"
)

// captureLogOutput redirects slog to a buffer for test assertions.
// NOT safe for parallel tests — mutates global logger state.
func captureLogOutput(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger.SetLogger(slog.New(handler))
	return &buf
}

func TestTrace_SetsTraceIDHeader(t *testing.T) {
	_ = captureLogOutput(t)

	app := fiber.New()
	app.Use(Trace())
	app.Get("/api/v1/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true})
	})

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	traceID := resp.Header.Get("X-Trace-ID")
	if traceID == "" {
		t.Error("X-Trace-ID header not set")
	}
	if len(traceID) != 36 {
		t.Errorf("trace_id length = %d, want 36 (UUID format)", len(traceID))
	}
}

func TestTrace_LogsRequestFields(t *testing.T) {
	buf := captureLogOutput(t)

	app := fiber.New()
	app.Use(Trace())
	app.Post("/api/v1/test", func(c *fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{"success": true})
	})

	body := `{"name":"test"}`
	req := httptest.NewRequest("POST", "/api/v1/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TestAgent/1.0")

	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log: %v\nraw: %s", err, buf.String())
	}

	if entry["msg"] != "request completed" {
		t.Errorf("msg = %v, want 'request completed'", entry["msg"])
	}
	if entry["method"] != "POST" {
		t.Errorf("method = %v, want POST", entry["method"])
	}
	if entry["path"] != "/api/v1/test" {
		t.Errorf("path = %v, want /api/v1/test", entry["path"])
	}
	if entry["request_body"] != body {
		t.Errorf("request_body = %v, want %v", entry["request_body"], body)
	}
	status := entry["response_status"].(float64)
	if int(status) != 201 {
		t.Errorf("response_status = %v, want 201", status)
	}
	if entry["trace_id"] == nil || entry["trace_id"] == "" {
		t.Error("trace_id missing from log entry")
	}
	if entry["duration_ms"] == nil {
		t.Error("duration_ms missing from log entry")
	}
}

func TestTrace_SkipsHealthCheck(t *testing.T) {
	buf := captureLogOutput(t)

	app := fiber.New()
	app.Use(Trace())
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no log output for health check, got: %s", buf.String())
	}
	if resp.Header.Get("X-Trace-ID") != "" {
		t.Error("X-Trace-ID should not be set for health check")
	}
}

func TestTrace_RedactsAuthorizationHeader(t *testing.T) {
	buf := captureLogOutput(t)

	app := fiber.New()
	app.Use(Trace())
	app.Get("/api/v1/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true})
	})

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer super-secret-token")

	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	logOutput := buf.String()
	if strings.Contains(logOutput, "super-secret-token") {
		t.Error("Authorization token value leaked into logs")
	}
	if !strings.Contains(logOutput, "Bearer [REDACTED]") {
		t.Error("Authorization header not redacted in logs")
	}
}

func TestTrace_RedactsPasswordOnAuthEndpoints(t *testing.T) {
	buf := captureLogOutput(t)

	app := fiber.New()
	app.Use(Trace())
	app.Post("/api/v1/auth/register", func(c *fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{"success": true})
	})

	body := `{"email":"test@x.com","password":"mysecret123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	logOutput := buf.String()
	if strings.Contains(logOutput, "mysecret123") {
		t.Error("password value leaked into logs")
	}
	// slog JSON-encodes the string value, so inner quotes are escaped
	if !strings.Contains(logOutput, `\"password\":\"[REDACTED]\"`) {
		t.Errorf("password not redacted in logs, got: %s", logOutput)
	}
}

func TestTrace_4xxLogsWarn(t *testing.T) {
	buf := captureLogOutput(t)

	app := fiber.New()
	app.Use(Trace())
	app.Get("/api/v1/test", func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{"success": false})
	})

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log: %v", err)
	}
	if entry["level"] != "WARN" {
		t.Errorf("level = %v, want WARN for 404", entry["level"])
	}
}

func TestTrace_5xxLogsError(t *testing.T) {
	buf := captureLogOutput(t)

	app := fiber.New()
	app.Use(Trace())
	app.Get("/api/v1/test", func(c *fiber.Ctx) error {
		return c.Status(500).JSON(fiber.Map{"success": false})
	})

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log: %v", err)
	}
	if entry["level"] != "ERROR" {
		t.Errorf("level = %v, want ERROR for 500", entry["level"])
	}
}

func TestTrace_MultipartBodyNotCaptured(t *testing.T) {
	buf := captureLogOutput(t)

	app := fiber.New()
	app.Use(Trace())
	app.Post("/api/v1/upload", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true})
	})

	// Use a properly-terminated multipart body to avoid parse errors
	boundary := "----TestBoundary"
	body := "--" + boundary + "\r\n" +
		"Content-Disposition: form-data; name=\"file\"; filename=\"test.txt\"\r\n" +
		"Content-Type: text/plain\r\n\r\n" +
		"file-content-here\r\n" +
		"--" + boundary + "--\r\n"

	req := httptest.NewRequest("POST", "/api/v1/upload", strings.NewReader(body))
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)

	_, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	logOutput := buf.String()
	if strings.Contains(logOutput, "file-content-here") {
		t.Error("multipart body content leaked into logs")
	}
	if !strings.Contains(logOutput, "[multipart/form-data]") {
		t.Error("multipart body not replaced with placeholder")
	}
}

func TestTrace_SetsLocalsTraceID(t *testing.T) {
	_ = captureLogOutput(t)

	var capturedTraceID string
	app := fiber.New()
	app.Use(Trace())
	app.Get("/api/v1/test", func(c *fiber.Ctx) error {
		capturedTraceID = c.Locals("trace_id").(string)
		return c.JSON(fiber.Map{"success": true})
	})

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if capturedTraceID == "" {
		t.Error("trace_id not set in c.Locals")
	}
	if capturedTraceID != resp.Header.Get("X-Trace-ID") {
		t.Error("c.Locals trace_id does not match X-Trace-ID header")
	}
}

func TestTrace_LargeBodyTruncated(t *testing.T) {
	buf := captureLogOutput(t)

	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024,
	})
	app.Use(Trace())
	app.Post("/api/v1/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true})
	})

	largeBody := strings.Repeat("x", 15000)
	req := httptest.NewRequest("POST", "/api/v1/test", strings.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")

	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "...[truncated]") {
		t.Error("large body not truncated in logs")
	}
}

func TestTrace_TransparentPassthrough(t *testing.T) {
	_ = captureLogOutput(t)

	app := fiber.New()
	app.Use(Trace())
	app.Post("/api/v1/test", func(c *fiber.Ctx) error {
		body := c.Body()
		return c.Status(200).JSON(fiber.Map{"received": len(body)})
	})

	originalBody := `{"key":"value","data":"important"}`
	req := httptest.NewRequest("POST", "/api/v1/test", strings.NewReader(originalBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	received := int(result["received"].(float64))
	if received != len(originalBody) {
		t.Errorf("handler received %d bytes, want %d", received, len(originalBody))
	}
}


func TestTrace_DoesNotRedactPasswordOnNonAuthEndpoints(t *testing.T) {
	buf := captureLogOutput(t)

	app := fiber.New()
	app.Use(Trace())
	app.Post("/api/v1/routines", func(c *fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{"success": true})
	})

	body := `{"title":"workout","password":"not-a-real-password"}`
	req := httptest.NewRequest("POST", "/api/v1/routines", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	_, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "not-a-real-password") {
		t.Error("password was redacted on non-auth endpoint — should only redact on /api/v1/auth/*")
	}
}
