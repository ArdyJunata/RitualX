# S1-08: Trace Middleware — Request/Response Telemetry

> **Date:** 2026-06-20
> **Status:** Approved
> **Story Points:** 5
> **Sprint:** Sprint 1 — Foundation & Auth

---

## Overview

A Fiber middleware that wraps every incoming HTTP request to capture full request/response telemetry as structured JSON logs. Generates a unique `trace_id` per request for end-to-end correlation across service layers.

## Goals

- Full request/response telemetry logged as a single structured JSON entry per request
- UUID v4 `trace_id` generated per request, propagated via `c.Locals` and context
- `X-Trace-ID` response header for client-side correlation
- Log level determined by response status code
- Sensitive data redacted (Authorization header, passwords on auth endpoints)
- Health check endpoint excluded from trace logging
- Body capture capped at 10KB to prevent log bloat

## Non-Goals

- Distributed tracing (OpenTelemetry, Jaeger) — out of scope
- Request rate limiting — separate middleware (future)
- Metrics collection (Prometheus) — not in Phase 1
- Log shipping to external services — handled by Promtail (S1-10)

## Dependencies

- S1-01: Backend scaffolding (logger package with `Get()`, `FromContext()`, `WithTraceID()`)
- `github.com/google/uuid` (already in go.mod)

---

## Design Details

### 1. File Location

```
backend/internal/middleware/trace.go
```

### 2. Function Signature

```go
package middleware

func Trace() fiber.Handler
```

Returns a Fiber handler function to be registered with `app.Use(middleware.Trace())` before all route groups.

### 3. Execution Flow

```
Request arrives
  │
  ├─ path == "/api/v1/health" → skip (return c.Next() immediately)
  │
  ├─ Generate trace_id = uuid.New().String()
  ├─ Set c.Locals("trace_id", traceID)
  ├─ Capture request metadata: method, path, host, user_agent, client_ip
  ├─ Capture request headers (redact Authorization)
  ├─ Capture request body (respect 10KB cap, skip multipart)
  ├─ Record start = time.Now()
  │
  ├─ chainErr := c.Next()
  │
  ├─ Calculate duration_ms = time.Since(start).Milliseconds()
  ├─ Capture response status
  ├─ Capture response body (respect 10KB cap)
  ├─ Set X-Trace-ID response header
  ├─ Redact passwords if path starts with "/api/v1/auth/"
  ├─ Determine log level: 2xx/3xx → INFO, 4xx → WARN, 5xx → ERROR
  ├─ Extract error message from chainErr or c.Locals("error")
  ├─ Log single structured entry via slog
  │
  └─ return chainErr
```

### 4. Captured Fields

| Field | Source | Type |
|-------|--------|------|
| `trace_id` | Generated UUID v4 | `string` |
| `method` | `c.Method()` | `string` |
| `path` | `c.Path()` | `string` |
| `host` | `c.Hostname()` | `string` |
| `user_agent` | `c.Get("User-Agent")` | `string` |
| `request_headers` | `c.GetReqHeaders()` sanitized | `map[string][]string` |
| `request_body` | `c.Body()` capped | `string` |
| `response_status` | `c.Response().StatusCode()` | `int` |
| `response_body` | `c.Response().Body()` capped | `string` |
| `duration_ms` | `time.Since(start).Milliseconds()` | `int64` |
| `client_ip` | `c.IP()` | `string` |
| `error` | `chainErr.Error()` or empty | `string` |

### 5. Log Output Format

```json
{
  "time": "2026-06-20T12:00:00.000Z",
  "level": "INFO",
  "msg": "request completed",
  "trace_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "method": "POST",
  "path": "/api/v1/routines/123/log",
  "host": "localhost:8080",
  "user_agent": "Mozilla/5.0...",
  "request_headers": {"Content-Type": ["application/json"], "Authorization": ["Bearer [REDACTED]"]},
  "request_body": "{\"date\":\"2026-06-20\",\"count\":1}",
  "response_status": 201,
  "response_body": "{\"success\":true,\"data\":{...}}",
  "duration_ms": 45,
  "client_ip": "192.168.1.1",
  "error": ""
}
```

### 6. Log Level Determination

```go
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

### 7. Security — Redaction Rules

**Authorization header:**
- In the sanitized headers map, replace any `Authorization` value with `"Bearer [REDACTED]"`
- Implementation: iterate headers, check key case-insensitively

```go
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
```

**Password redaction for auth endpoints:**
- Applies when path starts with `/api/v1/auth/`
- Regex replaces password field values in request body JSON
- Pattern: `"password"\s*:\s*"[^"]*"` → `"password":"[REDACTED]"`

```go
var passwordRegex = regexp.MustCompile(`"password"\s*:\s*"[^"]*"`)

func redactPasswords(body string) string {
    return passwordRegex.ReplaceAllString(body, `"password":"[REDACTED]"`)
}
```

### 8. Body Capture Rules

**Size cap:** 10,240 bytes (10KB)

```go
const maxBodySize = 10240

func captureBody(body []byte) string {
    if len(body) == 0 {
        return ""
    }
    if len(body) > maxBodySize {
        return string(body[:maxBodySize]) + "...[truncated]"
    }
    return string(body)
}
```

**Multipart/form-data skip:**
- Check `Content-Type` header for `multipart/form-data`
- If matched: log request body as `"[multipart/form-data]"` instead of actual content
- Response body and all other fields still captured normally

```go
func isMultipart(contentType string) bool {
    return strings.HasPrefix(contentType, "multipart/form-data")
}
```

### 9. Health Check Exclusion

First check in the middleware — zero-cost skip:

```go
if c.Path() == "/api/v1/health" {
    return c.Next()
}
```

### 10. Integration Point

In `cmd/server/main.go`, register before route groups:

```go
app := fiber.New(fiber.Config{})
app.Use(middleware.Trace())

api := app.Group("/api/v1")
api.Get("/health", handler.HealthCheck(db))
```

### 11. Error Capture

The middleware captures errors from two sources:

1. **Return value of `c.Next()`** — Fiber propagates handler errors here
2. **`c.Locals("error")`** — handlers can optionally set an error string for logging context

```go
var errMsg string
if chainErr != nil {
    errMsg = chainErr.Error()
}
```

---

## Acceptance Criteria

- [ ] Trace middleware registered on all routes via `app.Use(middleware.Trace())`
- [ ] UUID v4 `trace_id` generated per request
- [ ] `trace_id` set in `c.Locals("trace_id")` for downstream handler access
- [ ] `X-Trace-ID` response header set on every response (except health check)
- [ ] Single structured JSON log entry emitted per request containing all 12 fields
- [ ] Log level: INFO for 2xx/3xx, WARN for 4xx, ERROR for 5xx
- [ ] `Authorization` header value redacted as `"Bearer [REDACTED]"` in logged headers
- [ ] Request bodies on `/api/v1/auth/*` paths have `password` field values redacted
- [ ] Request/response bodies capped at 10KB with `...[truncated]` marker
- [ ] Multipart/form-data request bodies logged as `"[multipart/form-data]"` string
- [ ] `GET /api/v1/health` excluded from trace logging entirely
- [ ] No new dependencies required (uses `github.com/google/uuid` already in go.mod)
- [ ] Middleware does not alter request/response behavior — transparent passthrough

---

## Open Questions

None.
