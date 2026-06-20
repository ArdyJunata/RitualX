# S1-08 Code Review — Trace Middleware

> **Date:** 2026-06-20
> **Reviewer:** Code Reviewer Agent
> **Status:** Approved with minor fixes
> **Verdict:** Ready to merge after addressing items below

---

## Summary

Implementation matches spec. 25/25 tests pass. All 13 acceptance criteria verified. Security redaction works correctly. Code is idiomatic Go.

## Fixes to Apply

### 1. Defensive header value copy in `sanitizeHeaders`

**File:** `backend/internal/middleware/helpers.go:17`
**Issue:** `sanitized[k] = v` shares the slice backing array with original headers map.
**Fix:** Copy the slice to prevent potential mutation issues.

### 2. Add test for non-auth endpoint NOT redacting passwords

**File:** `backend/internal/middleware/trace_integration_test.go`
**Issue:** No test confirming password fields are left alone on non-auth endpoints.
**Fix:** Add test case verifying `/api/v1/routines` does NOT redact password fields.

### 3. Add parallel-safety comment on `captureLogOutput`

**File:** `backend/internal/middleware/trace_integration_test.go:18`
**Issue:** `captureLogOutput` mutates global logger state — would break with `t.Parallel()`.
**Fix:** Add comment noting it is NOT safe for parallel tests.

---

## Not Fixing (Acknowledged)

| Item | Reason |
|------|--------|
| `"error":""` always present in log | Simpler code, consistent field count for log parsing tools |
| `c.Set("X-Trace-ID")` after `c.Next()` | Works with Fiber buffered responses; no streaming in Phase 1 |
| `SetLogger` not thread-safe | Test-only function, tests run sequentially within package |
