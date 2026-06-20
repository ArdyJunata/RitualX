# S1-04 Code Review — User Registration

> **Date:** 2026-06-21
> **Reviewer:** Code Reviewer Agent
> **Status:** Approved with fixes
> **Verdict:** Ready to merge after addressing Important items

---

## Summary

Clean handler->service->repository architecture established. All 14 acceptance criteria verified. JWT, bcrypt, validation all correct. Tests pass with PostgreSQL.

## Fixes to Apply

### 1. TOCTOU race on email/username uniqueness (Important)

**File:** `backend/internal/service/auth.go:70`
**Issue:** Pre-check then Create has a race window. Under concurrency, two requests pass FindByEmail/FindByUsername checks, then one hits DB unique constraint. Error returned as generic INTERNAL_ERROR instead of proper EMAIL_TAKEN/USERNAME_TAKEN.
**Fix:** After `s.userRepo.Create(user)` fails, check error string for `duplicate key` and map to correct error code.

### 2. Inline fiber.Map responses — add response helpers (Important)

**File:** `backend/internal/handler/auth.go`
**Issue:** Every response manually constructs `fiber.Map{"success": true/false, ...}`. Violates DRY, risks typos, inconsistent format across future handlers.
**Fix:** Create `backend/internal/handler/response.go` with `success(c, status, data)` and `errorResponse(c, status, code, message)` helpers. Refactor auth handler to use them.

### 3. No repository test file (Minor)

**File:** `backend/internal/repository/`
**Issue:** Plan included repository tests but they weren't created. Coverage exists via service integration tests.
**Fix:** Optional — add `user_test.go` for direct repository testing.

### 4. `handleServiceError` placement (Minor)

**File:** `backend/internal/handler/auth.go`
**Issue:** `handleServiceError` and `mapErrorCodeToStatus` live in auth.go but will be shared by all future handlers.
**Fix:** Extract to `handler/response.go` alongside the response helpers (combined with fix #2).

---

## Not Fixing (Acknowledged)

| Item | Reason |
|------|--------|
| `pkg` package name is generic | Matches project convention from S1-01 |
| No request body size limit | Fiber default 4MB is fine for JSON auth payloads |
