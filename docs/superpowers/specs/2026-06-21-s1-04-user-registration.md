# S1-04: User Registration Endpoint

> **Date:** 2026-06-21
> **Status:** Approved
> **Story Points:** 5
> **Sprint:** Sprint 1 — Foundation & Auth

---

## Overview

Implement `POST /api/v1/auth/register` — creates a new user account with email, username, and password. Returns JWT access and refresh tokens on success. Establishes the handler→service→repository pattern used by all subsequent endpoints.

## Goals

- User can register with email, username, and password
- Password hashed with bcrypt before storage
- Returns JWT access token (15 min) and refresh token (7 days)
- Input validation with clear error messages
- Duplicate email/username returns 409 Conflict
- Establishes handler→service→repository architecture pattern

## Non-Goals

- Login endpoint (S1-05)
- Refresh token server-side storage/revocation (S1-05)
- Auth middleware for protected routes (S1-05)
- Email verification
- Rate limiting on auth endpoints

## Dependencies

- S1-01: Backend scaffolding (config, logger, Fiber app)
- S1-03: Users table migration + GORM model
- S1-08: Trace middleware (already active)

---

## Design Details

### 1. New Files

```
backend/internal/
├── handler/auth.go          # HTTP handler for POST /auth/register
├── service/auth.go          # Business logic: validate, hash, create, sign tokens
├── repository/user.go       # DB queries: create user, find by email/username
└── pkg/jwt.go               # JWT token generation utility
```

### 2. Request/Response Contract

**Request:** `POST /api/v1/auth/register`

```json
{
  "email": "user@example.com",
  "username": "ritualist",
  "password": "mypassword123"
}
```

**Validation Rules:**
| Field | Rule | Error Code |
|-------|------|------------|
| email | Required, valid email format (contains `@` and `.`) | `VALIDATION_ERROR` |
| username | Required, 3–20 chars, alphanumeric + underscore only | `VALIDATION_ERROR` |
| password | Required, min 8 chars | `VALIDATION_ERROR` |

**Success Response (201):**

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "username": "ritualist",
      "display_name": "",
      "xp": 0,
      "level": 1,
      "coins": 0,
      "title": "Novice",
      "created_at": "2026-06-21T00:00:00Z"
    },
    "access_token": "eyJ...",
    "refresh_token": "eyJ..."
  }
}
```

**Error Responses:**

| Status | Code | When |
|--------|------|------|
| 400 | `VALIDATION_ERROR` | Missing/invalid fields |
| 409 | `EMAIL_TAKEN` | Email already registered |
| 409 | `USERNAME_TAKEN` | Username already taken |
| 500 | `INTERNAL_ERROR` | Unexpected server error |

**Validation error format:**

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {"field": "email", "message": "invalid email format"},
      {"field": "password", "message": "must be at least 8 characters"}
    ]
  }
}
```

### 3. Repository Layer

**File:** `backend/internal/repository/user.go`

```go
type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository

func (r *UserRepository) Create(user *model.User) error
func (r *UserRepository) FindByEmail(email string) (*model.User, error)
func (r *UserRepository) FindByUsername(username string) (*model.User, error)
```

**Behavior:**
- `Create`: inserts user, returns error if unique constraint violated
- `FindByEmail`: returns user or nil (not error) if not found
- `FindByUsername`: returns user or nil if not found

### 4. Service Layer

**File:** `backend/internal/service/auth.go`

```go
type AuthService struct {
    userRepo  *repository.UserRepository
    jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService

func (s *AuthService) Register(req RegisterRequest) (*RegisterResponse, error)
```

**RegisterRequest struct:**

```go
type RegisterRequest struct {
    Email    string `json:"email"`
    Username string `json:"username"`
    Password string `json:"password"`
}
```

**RegisterResponse struct:**

```go
type RegisterResponse struct {
    User         model.User `json:"user"`
    AccessToken  string     `json:"access_token"`
    RefreshToken string     `json:"refresh_token"`
}
```

**Register logic:**
1. Validate input (email format, username 3–20 chars alphanumeric+underscore, password ≥ 8)
2. Check if email exists → return `EMAIL_TAKEN` error
3. Check if username exists → return `USERNAME_TAKEN` error
4. Hash password with bcrypt (cost 10)
5. Create user in DB
6. Generate access token (15 min expiry)
7. Generate refresh token (7 days expiry)
8. Return user + tokens

### 5. JWT Token Generation

**File:** `backend/pkg/jwt.go`

```go
package pkg

func GenerateAccessToken(userID string, secret string) (string, error)
func GenerateRefreshToken(userID string, secret string) (string, error)
```

**Access token claims:**

```go
type Claims struct {
    UserID string `json:"user_id"`
    Type   string `json:"type"` // "access" or "refresh"
    jwt.RegisteredClaims
}
```

- Access token: `exp` = now + 15 minutes, `type` = "access"
- Refresh token: `exp` = now + 7 days, `type` = "refresh"
- Signing method: HS256
- Library: `github.com/golang-jwt/jwt/v5`

### 6. Handler Layer

**File:** `backend/internal/handler/auth.go`

```go
func Register(authService *service.AuthService) fiber.Handler
```

**Logic:**
1. Parse JSON body into `RegisterRequest`
2. Call `authService.Register(req)`
3. Map service errors to HTTP responses (400/409/500)
4. Return 201 with user + tokens on success

### 7. Error Handling Pattern

Define a custom service error type for typed error responses:

**File:** `backend/internal/service/errors.go`

```go
type ServiceError struct {
    Code    string
    Message string
    Details []FieldError
}

type FieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func (e *ServiceError) Error() string { return e.Message }
```

The handler maps `ServiceError.Code` to HTTP status:
- `VALIDATION_ERROR` → 400
- `EMAIL_TAKEN`, `USERNAME_TAKEN` → 409
- anything else → 500

### 8. Route Registration

In `cmd/server/main.go`:

```go
userRepo := repository.NewUserRepository(db)
authService := service.NewAuthService(userRepo, cfg.JWTSecret)

auth := api.Group("/auth")
auth.Post("/register", handler.Register(authService))
```

### 9. Password Hashing

- Library: `golang.org/x/crypto/bcrypt` (already indirect dependency, will need direct import)
- Cost: 10 (default, good balance of security and speed)
- `bcrypt.GenerateFromPassword([]byte(password), 10)`

### 10. Username Validation Regex

```go
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
```

Allows: letters, digits, underscore. Length 3–20.

### 11. Email Validation

Simple check — not a full RFC 5322 parser:

```go
func isValidEmail(email string) bool {
    // Must contain exactly one @ with content before and after
    // Must have a . after the @
    parts := strings.SplitN(email, "@", 2)
    if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
        return false
    }
    return strings.Contains(parts[1], ".")
}
```

---

## Acceptance Criteria

- [ ] `POST /api/v1/auth/register` creates user and returns 201 with access + refresh JWT tokens
- [ ] Password is hashed with bcrypt (cost 10) before storage
- [ ] Duplicate email returns 409 with code `EMAIL_TAKEN`
- [ ] Duplicate username returns 409 with code `USERNAME_TAKEN`
- [ ] Invalid email format returns 400 with field-level error
- [ ] Password less than 8 chars returns 400 with field-level error
- [ ] Username outside 3–20 chars or non-alphanumeric returns 400 with field-level error
- [ ] Missing required fields returns 400 with field-level errors
- [ ] Access token has 15 min expiry and type "access"
- [ ] Refresh token has 7 day expiry and type "refresh"
- [ ] Tokens contain user_id claim
- [ ] User object in response excludes password_hash (json:"-")
- [ ] Handler→service→repository architecture pattern established
- [ ] No new tables required (uses existing users table from S1-03)

---

## Open Questions

None.
