# S1-08 Review Fixes — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Apply 3 minor fixes from code review on the trace middleware.

**Architecture:** Small targeted edits to existing files — one defensive copy, one new test, one comment.

**Tech Stack:** Go Fiber v2, slog

## Global Constraints

- Go module: `github.com/ArdyJunata/RitualX/backend`
- All code under `backend/` directory
- All existing tests must continue to pass after changes
- No new dependencies

---

### Task 1: All Three Review Fixes

**Files:**
- Modify: `backend/internal/middleware/helpers.go`
- Modify: `backend/internal/middleware/trace_integration_test.go`

**Interfaces:**
- Consumes: existing `sanitizeHeaders()`, existing `Trace()` middleware
- Produces: same functions with defensive improvement + 1 new test + 1 comment

- [ ] **Step 1: Fix defensive header value copy in `helpers.go`**

In file `backend/internal/middleware/helpers.go`, replace line 17:

```go
// Before:
sanitized[k] = v

// After:
sanitized[k] = append([]string{}, v...)
```

The full function should look like:

```go
func sanitizeHeaders(headers map[string][]string) map[string][]string {
	sanitized := make(map[string][]string, len(headers))
	for k, v := range headers {
		if strings.EqualFold(k, "Authorization") {
			sanitized[k] = []string{"Bearer [REDACTED]"}
		} else {
			sanitized[k] = append([]string{}, v...)
		}
	}
	return sanitized
}
```

- [ ] **Step 2: Add parallel-safety comment on `captureLogOutput`**

In file `backend/internal/middleware/trace_integration_test.go`, replace the existing comment + function signature:

```go
// Before:
func captureLogOutput(t *testing.T) *bytes.Buffer {

// After:
// captureLogOutput redirects slog to a buffer for test assertions.
// NOT safe for parallel tests — mutates global logger state.
func captureLogOutput(t *testing.T) *bytes.Buffer {
```

- [ ] **Step 3: Add test for non-auth endpoint NOT redacting passwords**

Append this test to the end of `backend/internal/middleware/trace_integration_test.go`:

```go
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
```

- [ ] **Step 4: Run all tests to verify nothing broke**

```bash
cd backend
go test ./internal/middleware/ -v -count=1
```

Expected: All 26 tests PASS (14 helper + 12 trace integration).

- [ ] **Step 5: Run full project tests**

```bash
cd backend
go test ./... -count=1
```

Expected: All packages pass.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/middleware/helpers.go backend/internal/middleware/trace_integration_test.go
git commit -m "fix(backend): apply review fixes — defensive copy, test coverage, comments"
```
