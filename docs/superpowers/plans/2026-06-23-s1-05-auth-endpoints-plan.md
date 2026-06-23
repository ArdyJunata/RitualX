# S1-05 Auth Endpoints Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Login, Refresh, and Logout endpoints with database-backed refresh tokens and HttpOnly cookies.

**Architecture:** We use an auth-specific `RefreshToken` GORM model and raw SQL migrations. The `AuthService` handles core logic, `handler/auth.go` sets `HttpOnly` cookies, and `middleware/auth.go` handles route protection via JWT validation.

**Tech Stack:** Go Fiber, GORM, PostgreSQL, `golang-jwt/jwt/v5`

## Global Constraints

- JWT access token: 15 min expiry.
- Refresh token: 7 days expiry.
- Use `HttpOnly`, `Secure`, `SameSite=Strict` cookies for refresh tokens.
- Follow handler→service→repository pattern.
- Use `ServiceError` for typed HTTP responses.

---

### Task 1: Database Migration & Model

**Files:**
- Create: `backend/migrations/000002_create_refresh_tokens.up.sql`
- Create: `backend/migrations/000002_create_refresh_tokens.down.sql`
- Create: `backend/internal/model/refresh_token.go`

**Interfaces:**
- Produces: `model.RefreshToken` struct

- [ ] **Step 1: Write UP migration**
Create `backend/migrations/000002_create_refresh_tokens.up.sql`:
```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR NOT NULL UNIQUE,
    user_agent VARCHAR,
    ip_address VARCHAR,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
```

- [ ] **Step 2: Write DOWN migration**
Create `backend/migrations/000002_create_refresh_tokens.down.sql`:
```sql
DROP TABLE IF EXISTS refresh_tokens;
```

- [ ] **Step 3: Define GORM Model**
Create `backend/internal/model/refresh_token.go`:
```go
package model

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Token     string    `gorm:"type:varchar;not null;uniqueIndex"`
	UserAgent string    `gorm:"type:varchar"`
	IPAddress string    `gorm:"type:varchar"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
```

- [ ] **Step 4: Commit**
```bash
git add backend/migrations/ backend/internal/model/refresh_token.go
git commit -m "feat: add refresh token migration and model"
```

---

### Task 2: Repository & JWT Utilities

**Files:**
- Create: `backend/internal/repository/refresh_token.go`
- Modify: `backend/pkg/jwt.go`
- Modify: `backend/cmd/server/main.go`

**Interfaces:**
- Produces: `repository.RefreshTokenRepository`, `pkg.ValidateToken`

- [ ] **Step 1: Create RefreshTokenRepository**
Create `backend/internal/repository/refresh_token.go`:
```go
package repository

import (
	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(token *model.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *RefreshTokenRepository) FindByToken(tokenStr string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	if err := r.db.Where("token = ?", tokenStr).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *RefreshTokenRepository) Delete(tokenStr string) error {
	return r.db.Where("token = ?", tokenStr).Delete(&model.RefreshToken{}).Error
}
```

- [ ] **Step 2: Add ValidateToken**
Update `backend/pkg/jwt.go` (append to file):
```go
func ValidateToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
```

- [ ] **Step 3: Update main.go**
Modify `backend/cmd/server/main.go`:
Around line 51, add `refreshTokenRepo`:
```go
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	authService := service.NewAuthService(userRepo, refreshTokenRepo, cfg.JWTSecret)
```

- [ ] **Step 4: Commit**
```bash
git add backend/internal/repository/refresh_token.go backend/pkg/jwt.go backend/cmd/server/main.go
git commit -m "feat: add refresh token repository and jwt validation"
```

---

### Task 3: Service Layer Updates

**Files:**
- Modify: `backend/internal/service/auth.go`

**Interfaces:**
- Consumes: `repository.RefreshTokenRepository`, `pkg.ValidateToken`
- Produces: `AuthService.Login`, `AuthService.Refresh`, `AuthService.Logout`

- [ ] **Step 1: Update AuthService struct**
In `backend/internal/service/auth.go`:
Update the struct and constructor:
```go
type AuthService struct {
	userRepo         *repository.UserRepository
	refreshTokenRepo *repository.RefreshTokenRepository
	jwtSecret        string
}

func NewAuthService(userRepo *repository.UserRepository, rtRepo *repository.RefreshTokenRepository, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, refreshTokenRepo: rtRepo, jwtSecret: jwtSecret}
}
```

- [ ] **Step 2: Add Login DTOs**
In `backend/internal/service/auth.go` (near top):
```go
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User         model.User `json:"user"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"-"` // Not serialized to JSON
}
```

Also, update `RegisterResponse` to ignore `RefreshToken` in JSON:
```go
type RegisterResponse struct {
	User         model.User `json:"user"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"-"`
}
```

- [ ] **Step 3: Implement Login**
In `backend/internal/service/auth.go`:
```go
func (s *AuthService) Login(req LoginRequest, ip, userAgent string) (*LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
	}
	if user == nil {
		return nil, &ServiceError{Code: "INVALID_CREDENTIALS", Message: "invalid email or password"}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, &ServiceError{Code: "INVALID_CREDENTIALS", Message: "invalid email or password"}
	}

	accessToken, _ := pkg.GenerateAccessToken(user.ID.String(), s.jwtSecret)
	refreshTokenStr, _ := pkg.GenerateRefreshToken(user.ID.String(), s.jwtSecret)

	rt := &model.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenStr,
		IPAddress: ip,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.refreshTokenRepo.Create(rt); err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "could not save session"}
	}

	return &LoginResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
	}, nil
}
```

- [ ] **Step 4: Update Register to save token**
In `backend/internal/service/auth.go`'s `Register` method, after generating the tokens, add the save logic:
```go
	rt := &model.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	_ = s.refreshTokenRepo.Create(rt) // ignore error to not block registration
```

- [ ] **Step 5: Implement Refresh & Logout**
In `backend/internal/service/auth.go`:
```go
import "time"

func (s *AuthService) Refresh(tokenStr string) (string, error) {
	if tokenStr == "" {
		return "", &ServiceError{Code: "UNAUTHORIZED", Message: "missing refresh token"}
	}
	rt, err := s.refreshTokenRepo.FindByToken(tokenStr)
	if err != nil || rt == nil {
		return "", &ServiceError{Code: "UNAUTHORIZED", Message: "invalid refresh token"}
	}
	if time.Now().After(rt.ExpiresAt) {
		_ = s.refreshTokenRepo.Delete(tokenStr)
		return "", &ServiceError{Code: "UNAUTHORIZED", Message: "refresh token expired"}
	}

	accessToken, _ := pkg.GenerateAccessToken(rt.UserID.String(), s.jwtSecret)
	return accessToken, nil
}

func (s *AuthService) Logout(tokenStr string) error {
	if tokenStr == "" {
		return nil
	}
	return s.refreshTokenRepo.Delete(tokenStr)
}
```

- [ ] **Step 6: Commit**
```bash
git add backend/internal/service/auth.go
git commit -m "feat: implement auth service login, refresh, logout logic"
```

---

### Task 4: Handler Layer Updates

**Files:**
- Modify: `backend/internal/handler/auth.go`
- Modify: `backend/cmd/server/main.go`

**Interfaces:**
- Consumes: `AuthService.Login`, `AuthService.Refresh`, `AuthService.Logout`

- [ ] **Step 1: Helper for setting cookie**
Add this function to `backend/internal/handler/auth.go`:
```go
import "time"

func setRefreshTokenCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	})
}

func clearRefreshTokenCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	})
}
```

- [ ] **Step 2: Update Register Handler**
In `backend/internal/handler/auth.go`, update `Register` handler:
```go
		resp, err := authService.Register(req)
		if err != nil {
			return handleServiceError(c, err)
		}
		setRefreshTokenCookie(c, resp.RefreshToken)
		return success(c, fiber.StatusCreated, resp)
```

- [ ] **Step 3: Add Login, Refresh, Logout Handlers**
In `backend/internal/handler/auth.go`:
```go
func Login(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req service.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return errorResponse(c, fiber.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		}

		resp, err := authService.Login(req, c.IP(), string(c.Request().Header.UserAgent()))
		if err != nil {
			return handleServiceError(c, err)
		}

		setRefreshTokenCookie(c, resp.RefreshToken)
		return success(c, fiber.StatusOK, resp)
	}
}

func Refresh(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("refresh_token")
		newAccessToken, err := authService.Refresh(token)
		if err != nil {
			return handleServiceError(c, err)
		}
		return success(c, fiber.StatusOK, fiber.Map{"access_token": newAccessToken})
	}
}

func Logout(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("refresh_token")
		_ = authService.Logout(token)
		clearRefreshTokenCookie(c)
		return success(c, fiber.StatusOK, fiber.Map{"message": "logged out"})
	}
}
```

- [ ] **Step 4: Register Routes in main.go**
In `backend/cmd/server/main.go`, add the routes under `auth.Post("/register", ...)`:
```go
	auth.Post("/login", handler.Login(authService))
	auth.Post("/refresh", handler.Refresh(authService))
	auth.Post("/logout", handler.Logout(authService))
```

- [ ] **Step 5: Commit**
```bash
git add backend/internal/handler/auth.go backend/cmd/server/main.go
git commit -m "feat: add login, refresh, logout http handlers"
```

---

### Task 5: Auth Middleware

**Files:**
- Create: `backend/internal/middleware/auth.go`

**Interfaces:**
- Produces: `middleware.RequireAuth`

- [ ] **Step 1: Write Middleware**
Create `backend/internal/middleware/auth.go`:
```go
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/ArdyJunata/RitualX/backend/pkg"
)

func RequireAuth(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "UNAUTHORIZED",
					"message": "missing or invalid authorization header",
				},
			})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := pkg.ValidateToken(tokenStr, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "UNAUTHORIZED",
					"message": "invalid or expired token",
				},
			})
		}

		c.Locals("user_id", claims.UserID)
		return c.Next()
	}
}
```

- [ ] **Step 2: Commit**
```bash
git add backend/internal/middleware/auth.go
git commit -m "feat: add auth middleware for protected routes"
```
