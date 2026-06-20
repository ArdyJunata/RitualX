# S1-04 Review Fixes — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Apply 2 Important fixes from code review: TOCTOU race handling and response helper extraction.

**Architecture:** Extract response formatting into shared helpers, add duplicate-key error detection in service layer.

**Tech Stack:** Go Fiber v2, GORM, PostgreSQL

## Global Constraints

- Go module: `github.com/ArdyJunata/RitualX/backend`
- All code under `backend/` directory
- All existing tests must continue to pass
- No new dependencies

---

### Task 1: Response Helpers & Handler Refactor

**Files:**
- Create: `backend/internal/handler/response.go`
- Modify: `backend/internal/handler/auth.go`

**Interfaces:**
- Consumes: `service.ServiceError` from `internal/service`
- Produces:
  - `handler.success(c *fiber.Ctx, status int, data interface{}) error`
  - `handler.errorResponse(c *fiber.Ctx, status int, code string, message string) error`
  - `handler.handleServiceError(c *fiber.Ctx, err error) error` (moved here)
  - `handler.mapErrorCodeToStatus(code string) int` (moved here)

- [ ] **Step 1: Create `response.go` with helpers**

Create file `backend/internal/handler/response.go`:

```go
package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

func success(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func errorResponse(c *fiber.Ctx, status int, code string, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    code,
			"message": message,
		},
	})
}

func handleServiceError(c *fiber.Ctx, err error) error {
	svcErr, ok := err.(*service.ServiceError)
	if !ok {
		return errorResponse(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "unexpected error")
	}

	status := mapErrorCodeToStatus(svcErr.Code)

	errResp := fiber.Map{
		"code":    svcErr.Code,
		"message": svcErr.Message,
	}
	if len(svcErr.Details) > 0 {
		errResp["details"] = svcErr.Details
	}

	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error":   errResp,
	})
}

func mapErrorCodeToStatus(code string) int {
	switch code {
	case "VALIDATION_ERROR", "INVALID_REQUEST":
		return fiber.StatusBadRequest
	case "EMAIL_TAKEN", "USERNAME_TAKEN":
		return fiber.StatusConflict
	default:
		return fiber.StatusInternalServerError
	}
}
```

- [ ] **Step 2: Refactor `auth.go` to use helpers**

Replace `backend/internal/handler/auth.go` with:

```go
package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

func Register(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req service.RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		}

		resp, err := authService.Register(req)
		if err != nil {
			return handleServiceError(c, err)
		}

		return success(c, fiber.StatusCreated, resp)
	}
}
```

- [ ] **Step 3: Run tests to verify nothing broke**

```bash
cd backend
go test ./internal/handler/ -v -count=1
```

Expected: All 6 handler tests PASS (health + auth).

- [ ] **Step 4: Commit**

```bash
git add backend/internal/handler/response.go backend/internal/handler/auth.go
git commit -m "refactor(backend): extract response helpers into handler/response.go"
```

---

### Task 2: TOCTOU Race Fix — Handle Duplicate Key on Create

**Files:**
- Modify: `backend/internal/service/auth.go`
- Create: `backend/internal/service/auth_race_test.go`

**Interfaces:**
- Consumes: `s.userRepo.Create()` error containing PostgreSQL duplicate key message
- Produces: Proper `EMAIL_TAKEN`/`USERNAME_TAKEN` error even when pre-check passed but DB constraint caught it

- [ ] **Step 1: Write test for duplicate key handling**

Create file `backend/internal/service/auth_race_test.go`:

```go
package service

import (
	"testing"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
)

func TestRegister_DuplicateKeyOnCreate_Email(t *testing.T) {
	svc, db := newAuthService(t)
	defer cleanupUser(t, db, "race-email@test.com")

	// Insert user directly to simulate race condition
	db.Create(&model.User{
		Email:        "race-email@test.com",
		PasswordHash: "hash",
		Username:     "raceuser1",
	})

	// Service pre-check would pass if called with different timing,
	// but Create will hit the constraint
	_, err := svc.Register(RegisterRequest{
		Email:    "race-email@test.com",
		Username: "raceuser2",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	svcErr, ok := err.(*ServiceError)
	if !ok {
		t.Fatalf("expected *ServiceError, got %T", err)
	}
	if svcErr.Code != "EMAIL_TAKEN" {
		t.Errorf("code = %q, want EMAIL_TAKEN", svcErr.Code)
	}
}

func TestRegister_DuplicateKeyOnCreate_Username(t *testing.T) {
	svc, db := newAuthService(t)
	defer cleanupUser(t, db, "race-uname@test.com")
	defer cleanupUser(t, db, "race-uname2@test.com")

	// Insert user directly
	db.Create(&model.User{
		Email:        "race-uname@test.com",
		PasswordHash: "hash",
		Username:     "samerace",
	})

	// Register with same username but different email
	_, err := svc.Register(RegisterRequest{
		Email:    "race-uname2@test.com",
		Username: "samerace",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	svcErr, ok := err.(*ServiceError)
	if !ok {
		t.Fatalf("expected *ServiceError, got %T", err)
	}
	if svcErr.Code != "USERNAME_TAKEN" {
		t.Errorf("code = %q, want USERNAME_TAKEN", svcErr.Code)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend
go test ./internal/service/ -v -run "TestRegister_DuplicateKeyOnCreate" -count=1
```

Expected: FAIL — returns `INTERNAL_ERROR` instead of `EMAIL_TAKEN`/`USERNAME_TAKEN`.

- [ ] **Step 3: Fix the Create error handling in auth.go**

In `backend/internal/service/auth.go`, replace the Create block:

```go
// Before:
if err := s.userRepo.Create(user); err != nil {
    return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
}

// After:
if err := s.userRepo.Create(user); err != nil {
    return nil, mapCreateError(err)
}
```

Add this helper function at the bottom of the file (before `validateRegister`):

```go
func mapCreateError(err error) *ServiceError {
	msg := err.Error()
	if strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint") {
		if strings.Contains(msg, "email") {
			return &ServiceError{Code: "EMAIL_TAKEN", Message: "email already registered"}
		}
		if strings.Contains(msg, "username") {
			return &ServiceError{Code: "USERNAME_TAKEN", Message: "username already taken"}
		}
	}
	return &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend
go test ./internal/service/ -v -count=1
```

Expected: All service tests PASS including the new race condition tests.

- [ ] **Step 5: Run full project tests**

```bash
cd backend
go test ./... -count=1
```

Expected: All packages pass.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/service/auth.go backend/internal/service/auth_race_test.go
git commit -m "fix(backend): handle duplicate key race condition on user create"
```
