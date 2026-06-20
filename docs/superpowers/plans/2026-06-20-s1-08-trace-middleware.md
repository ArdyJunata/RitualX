# S1-08: Trace Middleware Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Fiber middleware that logs full request/response telemetry as structured JSON with trace_id correlation.

**Architecture:** Single middleware function `Trace()` registered via `app.Use()` before all routes. Generates UUID v4 trace_id, captures request metadata before `c.Next()`, captures response after, logs a single structured entry via the existing `slog` logger. Helper functions handle redaction, body capping, and level determination.

**Tech Stack:** Go Fiber v2, slog, github.com/google/uuid

## Global Constraints

- Go module: `github.com/ArdyJunata/RitualX/backend`
- All code under `backend/` directory
- JSON responses follow: `{"success": true, "data": {...}}` or `{"success": false, "error": {"code": "...", "message": "..."}}`
- No new dependencies — `github.com/google/uuid` already in go.mod
- Branch: `feat/s1-08-trace-middleware`
- Middleware must be transparent — does not alter request/response behavior

---


### Task 1: Helper Functions & Unit Tests

**Files:**
- Create: `backend/internal/middleware/helpers.go`
- Create: `backend/internal/middleware/helpers_test.go`
- Delete: `backend/internal/middleware/.gitkeep`

**Interfaces:**
- Consumes: nothing
- Produces:
  - `sanitizeHeaders(headers map[string][]string) map[string][]string`
  - `captureBody(body []byte) string`
  - `isMultipart(contentType string) bool`
  - `redactPasswords(body string) string`
  - `logLevel(status int) slog.Level`

- [ ] **Step 1: Delete .gitkeep placeholder**

```bash
cd backend
rm internal/middleware/.gitkeep
```

- [ ] **Step 2: Write failing tests for all helper functions**

Create file `backend/internal/middleware/helpers_test.go`:

```go
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
```

- [ ] **Step 3: Run tests to verify they fail**

```bash
cd backend
go test ./internal/middleware/ -v
```

Expected: FAIL — functions not defined.

- [ ] **Step 4: Implement helper functions**

Create file `backend/internal/middleware/helpers.go`:

```go
package middleware

import (
	"log/slog"
	"regexp"
	"strings"
)

const maxBodySize = 10240

var passwordRegex = regexp.MustCompile(`"password"\s*:\s*"[^"]*"`)

func sanitizeHeaders(headers map[string][]string) map[string][]string {
	sanitized := make(map[string][]string, len(headers))
	for k, v := range headers {
		if strings.EqualFold(k, "Authorization") {
			sanitized[k] = []string{"Bearer [REDACTED]"}
		} else {
			sanitized[k] = v
		}
	}
	return sanitized
}

func captureBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	if len(body) > maxBodySize {
		return string(body[:maxBodySize]) + "...[truncated]"
	}
	return string(body)
}

func isMultipart(contentType string) bool {
	return strings.HasPrefix(contentType, "multipart/form-data")
}

func redactPasswords(body string) string {
	return passwordRegex.ReplaceAllString(body, `"password":"[REDACTED]"`)
}

func logLevel(status int) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd backend
go test ./internal/middleware/ -v
```

Expected: All 12 tests PASS.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/middleware/helpers.go backend/internal/middleware/helpers_test.go
git rm backend/internal/middleware/.gitkeep
git commit -m "feat(backend): add trace middleware helper functions with tests"
```

---


### Task 2: Trace Middleware Function & Integration Tests

**Files:**
- Create: `backend/internal/middleware/trace.go`
- Create: `backend/internal/middleware/trace_test.go`

**Interfaces:**
- Consumes:
  - `logger.Get() *slog.Logger` from `internal/logger`
  - `sanitizeHeaders()`, `captureBody()`, `isMultipart()`, `redactPasswords()`, `logLevel()` from Task 1
- Produces:
  - `middleware.Trace() fiber.Handler`

- [ ] **Step 1: Write integration tests for Trace middleware**

Create file `backend/internal/middleware/trace_test.go` (append to existing test file — actually create as separate file `trace_integration_test.go`):

Create file `backend/internal/middleware/trace_integration_test.go`:

```go
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

// captureLogOutput redirects slog to a buffer for test assertions
func captureLogOutput(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger.SetLogger(slog.New(handler))
	return &buf
}

func TestTrace_SetsTraceIDHeader(t *testing.T) {
	buf := captureLogOutput(t)
	_ = buf

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
	if len(traceID) != 36 { // UUID format: 8-4-4-4-12
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
	if !strings.Contains(logOutput, `"password":"[REDACTED]"`) {
		t.Error("password not redacted in logs")
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

	req := httptest.NewRequest("POST", "/api/v1/upload", strings.NewReader("file-content-here"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=----boundary")

	_, err := app.Test(req)
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
		BodyLimit: 20 * 1024, // 20KB limit to allow large body
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
		// Verify the handler still receives the full body
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
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend
go test ./internal/middleware/ -v -run "TestTrace_"
```

Expected: FAIL — `Trace` not defined, `logger.SetLogger` not defined.

- [ ] **Step 3: Add `SetLogger` to logger package (test helper)**

Modify file `backend/internal/logger/logger.go` — add this function at the end:

```go
// SetLogger replaces the default logger (used for testing).
func SetLogger(l *slog.Logger) {
	defaultLogger = l
}
```

- [ ] **Step 4: Implement the Trace middleware**

Create file `backend/internal/middleware/trace.go`:

```go
package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/ArdyJunata/RitualX/backend/internal/logger"
)

func Trace() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/api/v1/health" {
			return c.Next()
		}

		traceID := uuid.New().String()
		c.Locals("trace_id", traceID)

		method := c.Method()
		path := c.Path()
		host := c.Hostname()
		userAgent := c.Get("User-Agent")
		clientIP := c.IP()
		headers := sanitizeHeaders(c.GetReqHeaders())

		var reqBody string
		if isMultipart(c.Get("Content-Type")) {
			reqBody = "[multipart/form-data]"
		} else {
			reqBody = captureBody(c.Body())
		}

		start := time.Now()

		chainErr := c.Next()

		durationMs := time.Since(start).Milliseconds()
		status := c.Response().StatusCode()
		respBody := captureBody(c.Response().Body())

		c.Set("X-Trace-ID", traceID)

		if strings.HasPrefix(path, "/api/v1/auth/") {
			reqBody = redactPasswords(reqBody)
		}

		var errMsg string
		if chainErr != nil {
			errMsg = chainErr.Error()
		}

		level := logLevel(status)
		log := logger.Get()

		log.Log(c.UserContext(), level, "request completed",
			"trace_id", traceID,
			"method", method,
			"path", path,
			"host", host,
			"user_agent", userAgent,
			"request_headers", headers,
			"request_body", reqBody,
			"response_status", status,
			"response_body", respBody,
			"duration_ms", durationMs,
			"client_ip", clientIP,
			"error", errMsg,
		)

		return chainErr
	}
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd backend
go test ./internal/middleware/ -v
```

Expected: All tests PASS (helpers + trace integration).

- [ ] **Step 6: Commit**

```bash
git add backend/internal/middleware/trace.go backend/internal/middleware/trace_integration_test.go backend/internal/logger/logger.go
git commit -m "feat(backend): add Trace middleware with request/response telemetry"
```

---


### Task 3: Register Middleware in main.go & Verify End-to-End

**Files:**
- Modify: `backend/cmd/server/main.go`

**Interfaces:**
- Consumes: `middleware.Trace() fiber.Handler` from Task 2
- Produces: Running server with trace middleware active on all routes

- [ ] **Step 1: Update main.go to register Trace middleware**

Modify `backend/cmd/server/main.go` — add the import and middleware registration:

Replace the import block with:

```go
import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ArdyJunata/RitualX/backend/internal/config"
	"github.com/ArdyJunata/RitualX/backend/internal/handler"
	"github.com/ArdyJunata/RitualX/backend/internal/logger"
	"github.com/ArdyJunata/RitualX/backend/internal/middleware"
)
```

Replace the Fiber app setup section (after `log.Info("database connected")`) with:

```go
	app := fiber.New(fiber.Config{})

	// Register trace middleware before all routes
	app.Use(middleware.Trace())

	api := app.Group("/api/v1")
	api.Get("/health", handler.HealthCheck(db))
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend
go build ./cmd/server/
```

Expected: compiles without errors.

- [ ] **Step 3: Run all tests to ensure nothing broke**

```bash
cd backend
go test ./... -v
```

Expected: All tests pass (config: 3, logger: 5, handler: 2, middleware: ~22).

- [ ] **Step 4: Manual verification (requires running postgres)**

Start postgres and server:

```bash
cd backend
docker compose -f docker-compose.dev.yml up -d
cp .env.example .env
go run cmd/server/main.go
```

In another terminal, test health check produces NO trace log:

```bash
curl http://localhost:8080/api/v1/health
```

Expected: Response has NO `X-Trace-ID` header. Server stdout has NO "request completed" log.

Test a non-health route produces trace log:

```bash
curl -H "Authorization: Bearer test-token" http://localhost:8080/api/v1/nonexistent
```

Expected: Response has `X-Trace-ID` header. Server stdout shows JSON log with:
- `"msg":"request completed"`
- `"level":"WARN"` (404)
- `"trace_id":"<uuid>"`
- `"request_headers"` contains `"Bearer [REDACTED]"` (not the actual token)

Stop server with Ctrl+C.

- [ ] **Step 5: Commit**

```bash
git add backend/cmd/server/main.go
git commit -m "feat(backend): register Trace middleware in server entry point"
```

---

## Verification Checklist

After all tasks complete, verify each acceptance criterion:

| # | Criterion | How to Verify |
|---|-----------|---------------|
| 1 | Middleware registered on all routes | `main.go` has `app.Use(middleware.Trace())` |
| 2 | UUID v4 trace_id generated | `TestTrace_SetsTraceIDHeader` — 36 char UUID |
| 3 | trace_id in c.Locals | `TestTrace_SetsLocalsTraceID` |
| 4 | X-Trace-ID response header | `TestTrace_SetsTraceIDHeader` |
| 5 | 12 fields logged | `TestTrace_LogsRequestFields` |
| 6 | Log level by status | `TestTrace_4xxLogsWarn`, `TestTrace_5xxLogsError` |
| 7 | Authorization redacted | `TestTrace_RedactsAuthorizationHeader` |
| 8 | Password redacted on auth paths | `TestTrace_RedactsPasswordOnAuthEndpoints` |
| 9 | Body capped at 10KB | `TestTrace_LargeBodyTruncated` |
| 10 | Multipart body skipped | `TestTrace_MultipartBodyNotCaptured` |
| 11 | Health check excluded | `TestTrace_SkipsHealthCheck` |
| 12 | No new dependencies | `go.mod` unchanged |
| 13 | Transparent passthrough | `TestTrace_TransparentPassthrough` |
