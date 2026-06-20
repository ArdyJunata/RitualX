# S1-04: User Registration — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `POST /api/v1/auth/register` with bcrypt password hashing, JWT token generation, input validation, and the handler→service→repository pattern.

**Architecture:** Request flows through handler (HTTP parsing + error mapping) → service (validation + business logic) → repository (DB queries). JWT generation lives in `pkg/` as a shared utility. Custom `ServiceError` type enables typed error responses.

**Tech Stack:** Go Fiber v2, GORM, bcrypt, github.com/golang-jwt/jwt/v5

## Global Constraints

- Go module: `github.com/ArdyJunata/RitualX/backend`
- All code under `backend/` directory
- JSON responses: `{"success": true, "data": {...}}` or `{"success": false, "error": {"code": "...", "message": "..."}}`
- Branch: `feat/s1-04-user-registration`
- Uses existing `users` table from S1-03 (no new migrations)
- PostgreSQL must be running: `docker compose -f backend/docker-compose.dev.yml up -d`

---

### Task 1: JWT Token Generation Utility

**Files:**
- Create: `backend/pkg/jwt.go`
- Create: `backend/pkg/jwt_test.go`
- Delete: `backend/pkg/.gitkeep`

**Interfaces:**
- Consumes: `github.com/golang-jwt/jwt/v5` (new dependency)
- Produces:
  - `pkg.Claims` struct
  - `pkg.GenerateAccessToken(userID string, secret string) (string, error)`
  - `pkg.GenerateRefreshToken(userID string, secret string) (string, error)`

- [ ] **Step 1: Install JWT dependency**

```bash
cd backend
go get github.com/golang-jwt/jwt/v5
```

- [ ] **Step 2: Delete .gitkeep placeholder**

```bash
cd backend
rm pkg/.gitkeep
```

- [ ] **Step 3: Write failing tests**

Create file `backend/pkg/jwt_test.go`:

```go
package pkg

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key"

func TestGenerateAccessToken_Valid(t *testing.T) {
	token, err := GenerateAccessToken("user-123", testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}

	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !parsed.Valid {
		t.Error("token is not valid")
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.Type != "access" {
		t.Errorf("Type = %q, want %q", claims.Type, "access")
	}

	exp := claims.ExpiresAt.Time
	expectedExp := time.Now().Add(15 * time.Minute)
	diff := expectedExp.Sub(exp)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("expiry off by %v, want ~15 min from now", diff)
	}
}

func TestGenerateRefreshToken_Valid(t *testing.T) {
	token, err := GenerateRefreshToken("user-456", testSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}

	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !parsed.Valid {
		t.Error("token is not valid")
	}
	if claims.UserID != "user-456" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-456")
	}
	if claims.Type != "refresh" {
		t.Errorf("Type = %q, want %q", claims.Type, "refresh")
	}

	exp := claims.ExpiresAt.Time
	expectedExp := time.Now().Add(7 * 24 * time.Hour)
	diff := expectedExp.Sub(exp)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("expiry off by %v, want ~7 days from now", diff)
	}
}

func TestGenerateAccessToken_DifferentFromRefresh(t *testing.T) {
	access, _ := GenerateAccessToken("user-1", testSecret)
	refresh, _ := GenerateRefreshToken("user-1", testSecret)
	if access == refresh {
		t.Error("access and refresh tokens should be different")
	}
}
```

- [ ] **Step 4: Run tests to verify they fail**

```bash
cd backend
go test ./pkg/ -v
```

Expected: FAIL — `GenerateAccessToken` not defined.

- [ ] **Step 5: Implement JWT utility**

Create file `backend/pkg/jwt.go`:

```go
package pkg

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID string, secret string) (string, error) {
	return generateToken(userID, secret, "access", 15*time.Minute)
}

func GenerateRefreshToken(userID string, secret string) (string, error) {
	return generateToken(userID, secret, "refresh", 7*24*time.Hour)
}

func generateToken(userID, secret, tokenType string, expiry time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
```

- [ ] **Step 6: Run tests to verify they pass**

```bash
cd backend
go test ./pkg/ -v
```

Expected: All 3 tests PASS.

- [ ] **Step 7: Commit**

```bash
git add backend/pkg/jwt.go backend/pkg/jwt_test.go
git rm backend/pkg/.gitkeep
git commit -m "feat(backend): add JWT token generation utility"
```

---


### Task 2: Service Error Types & User Repository

**Files:**
- Create: `backend/internal/service/errors.go`
- Create: `backend/internal/repository/user.go`
- Create: `backend/internal/repository/user_test.go`
- Delete: `backend/internal/service/.gitkeep`
- Delete: `backend/internal/repository/.gitkeep`

**Interfaces:**
- Consumes: `model.User` from `internal/model`, `*gorm.DB`
- Produces:
  - `service.ServiceError{Code, Message, Details}`, `service.FieldError{Field, Message}`
  - `repository.NewUserRepository(db *gorm.DB) *UserRepository`
  - `(*UserRepository).Create(user *model.User) error`
  - `(*UserRepository).FindByEmail(email string) (*model.User, error)`
  - `(*UserRepository).FindByUsername(username string) (*model.User, error)`

- [ ] **Step 1: Delete .gitkeep placeholders**

```bash
cd backend
rm internal/service/.gitkeep internal/repository/.gitkeep
```

- [ ] **Step 2: Create service error types**

Create file `backend/internal/service/errors.go`:

```go
package service

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ServiceError struct {
	Code    string
	Message string
	Details []FieldError
}

func (e *ServiceError) Error() string {
	return e.Message
}
```

- [ ] **Step 3: Write repository tests**

Create file `backend/internal/repository/user_test.go`:

```go
package repository

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "host=localhost port=5432 user=ritualx password=ritualx_dev dbname=ritualx sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration test")
	}
	return db
}

func cleanupUser(t *testing.T, db *gorm.DB, email string) {
	t.Helper()
	db.Where("email = ?", email).Delete(&model.User{})
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	email := "test-create-" + uuid.New().String()[:8] + "@test.com"
	defer cleanupUser(t, db, email)

	user := &model.User{
		Email:        email,
		PasswordHash: "hashed-password",
		Username:     "testuser" + uuid.New().String()[:6],
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if user.ID == uuid.Nil {
		t.Error("user ID not set after create")
	}
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	email := "test-dup-" + uuid.New().String()[:8] + "@test.com"
	defer cleanupUser(t, db, email)

	user1 := &model.User{
		Email:        email,
		PasswordHash: "hash1",
		Username:     "user1" + uuid.New().String()[:6],
	}
	_ = repo.Create(user1)

	user2 := &model.User{
		Email:        email,
		PasswordHash: "hash2",
		Username:     "user2" + uuid.New().String()[:6],
	}
	err := repo.Create(user2)
	if err == nil {
		t.Error("expected error for duplicate email, got nil")
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	email := "test-find-" + uuid.New().String()[:8] + "@test.com"
	defer cleanupUser(t, db, email)

	user := &model.User{
		Email:        email,
		PasswordHash: "hash",
		Username:     "finduser" + uuid.New().String()[:6],
	}
	_ = repo.Create(user)

	found, err := repo.FindByEmail(email)
	if err != nil {
		t.Fatalf("FindByEmail failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected user, got nil")
	}
	if found.Email != email {
		t.Errorf("email = %q, want %q", found.Email, email)
	}
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	found, err := repo.FindByEmail("nonexistent@test.com")
	if err != nil {
		t.Fatalf("FindByEmail failed: %v", err)
	}
	if found != nil {
		t.Error("expected nil for nonexistent email")
	}
}

func TestUserRepository_FindByUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	email := "test-uname-" + uuid.New().String()[:8] + "@test.com"
	username := "uname" + uuid.New().String()[:6]
	defer cleanupUser(t, db, email)

	user := &model.User{
		Email:        email,
		PasswordHash: "hash",
		Username:     username,
	}
	_ = repo.Create(user)

	found, err := repo.FindByUsername(username)
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}
	if found == nil {
		t.Fatal("expected user, got nil")
	}
	if found.Username != username {
		t.Errorf("username = %q, want %q", found.Username, username)
	}
}

func TestUserRepository_FindByUsername_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	found, err := repo.FindByUsername("nonexistent_user_xyz")
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}
	if found != nil {
		t.Error("expected nil for nonexistent username")
	}
}
```

- [ ] **Step 4: Run tests to verify they fail**

```bash
cd backend
go test ./internal/repository/ -v
```

Expected: FAIL — `NewUserRepository` not defined.

- [ ] **Step 5: Implement user repository**

Create file `backend/internal/repository/user.go`:

```go
package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
```

- [ ] **Step 6: Run tests to verify they pass**

```bash
cd backend
go test ./internal/repository/ -v
```

Expected: All 6 tests PASS (requires running PostgreSQL).

- [ ] **Step 7: Commit**

```bash
git add backend/internal/service/errors.go backend/internal/repository/user.go backend/internal/repository/user_test.go
git rm backend/internal/service/.gitkeep backend/internal/repository/.gitkeep
git commit -m "feat(backend): add ServiceError types and UserRepository"
```

---


### Task 3: Auth Service (Validation + Business Logic)

**Files:**
- Create: `backend/internal/service/auth.go`
- Create: `backend/internal/service/auth_test.go`

**Interfaces:**
- Consumes:
  - `repository.UserRepository` from Task 2
  - `pkg.GenerateAccessToken`, `pkg.GenerateRefreshToken` from Task 1
  - `golang.org/x/crypto/bcrypt`
- Produces:
  - `service.NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService`
  - `(*AuthService).Register(req RegisterRequest) (*RegisterResponse, error)`
  - `service.RegisterRequest{Email, Username, Password}`
  - `service.RegisterResponse{User, AccessToken, RefreshToken}`

- [ ] **Step 1: Install bcrypt dependency**

```bash
cd backend
go get golang.org/x/crypto/bcrypt
```

- [ ] **Step 2: Write failing tests for auth service**

Create file `backend/internal/service/auth_test.go`:

```go
package service

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/ArdyJunata/RitualX/backend/internal/repository"
	"github.com/ArdyJunata/RitualX/backend/pkg"
)

const testJWTSecret = "test-jwt-secret"

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "host=localhost port=5432 user=ritualx password=ritualx_dev dbname=ritualx sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration test")
	}
	return db
}

func cleanupUser(t *testing.T, db *gorm.DB, email string) {
	t.Helper()
	db.Where("email = ?", email).Delete(&model.User{})
}

func newAuthService(t *testing.T) (*AuthService, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	svc := NewAuthService(repo, testJWTSecret)
	return svc, db
}

func TestRegister_Success(t *testing.T) {
	svc, db := newAuthService(t)
	defer cleanupUser(t, db, "register-ok@test.com")

	resp, err := svc.Register(RegisterRequest{
		Email:    "register-ok@test.com",
		Username: "registerok",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if resp.User.Email != "register-ok@test.com" {
		t.Errorf("email = %q, want %q", resp.User.Email, "register-ok@test.com")
	}
	if resp.User.Username != "registerok" {
		t.Errorf("username = %q, want %q", resp.User.Username, "registerok")
	}
	if resp.AccessToken == "" {
		t.Error("access token is empty")
	}
	if resp.RefreshToken == "" {
		t.Error("refresh token is empty")
	}

	// Verify access token claims
	claims := &pkg.Claims{}
	jwt.ParseWithClaims(resp.AccessToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(testJWTSecret), nil
	})
	if claims.UserID != resp.User.ID.String() {
		t.Errorf("token user_id = %q, want %q", claims.UserID, resp.User.ID.String())
	}
	if claims.Type != "access" {
		t.Errorf("token type = %q, want %q", claims.Type, "access")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, db := newAuthService(t)
	defer cleanupUser(t, db, "dup-email@test.com")

	req := RegisterRequest{
		Email:    "dup-email@test.com",
		Username: "dupuser1",
		Password: "password123",
	}
	_, _ = svc.Register(req)

	req.Username = "dupuser2"
	_, err := svc.Register(req)
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
	svcErr, ok := err.(*ServiceError)
	if !ok {
		t.Fatalf("expected *ServiceError, got %T", err)
	}
	if svcErr.Code != "EMAIL_TAKEN" {
		t.Errorf("code = %q, want %q", svcErr.Code, "EMAIL_TAKEN")
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	svc, db := newAuthService(t)
	defer cleanupUser(t, db, "dup-uname1@test.com")
	defer cleanupUser(t, db, "dup-uname2@test.com")

	req := RegisterRequest{
		Email:    "dup-uname1@test.com",
		Username: "sameuser",
		Password: "password123",
	}
	_, _ = svc.Register(req)

	req.Email = "dup-uname2@test.com"
	_, err := svc.Register(req)
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
	svcErr, ok := err.(*ServiceError)
	if !ok {
		t.Fatalf("expected *ServiceError, got %T", err)
	}
	if svcErr.Code != "USERNAME_TAKEN" {
		t.Errorf("code = %q, want %q", svcErr.Code, "USERNAME_TAKEN")
	}
}

func TestRegister_ValidationErrors(t *testing.T) {
	svc, _ := newAuthService(t)

	tests := []struct {
		name  string
		req   RegisterRequest
		field string
	}{
		{"empty email", RegisterRequest{Email: "", Username: "valid", Password: "12345678"}, "email"},
		{"invalid email", RegisterRequest{Email: "notanemail", Username: "valid", Password: "12345678"}, "email"},
		{"empty username", RegisterRequest{Email: "a@b.com", Username: "", Password: "12345678"}, "username"},
		{"short username", RegisterRequest{Email: "a@b.com", Username: "ab", Password: "12345678"}, "username"},
		{"long username", RegisterRequest{Email: "a@b.com", Username: "abcdefghijklmnopqrstu", Password: "12345678"}, "username"},
		{"invalid chars username", RegisterRequest{Email: "a@b.com", Username: "bad-user!", Password: "12345678"}, "username"},
		{"empty password", RegisterRequest{Email: "a@b.com", Username: "valid", Password: ""}, "password"},
		{"short password", RegisterRequest{Email: "a@b.com", Username: "valid", Password: "1234567"}, "password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Register(tt.req)
			if err == nil {
				t.Fatal("expected validation error")
			}
			svcErr, ok := err.(*ServiceError)
			if !ok {
				t.Fatalf("expected *ServiceError, got %T", err)
			}
			if svcErr.Code != "VALIDATION_ERROR" {
				t.Errorf("code = %q, want %q", svcErr.Code, "VALIDATION_ERROR")
			}
			found := false
			for _, d := range svcErr.Details {
				if d.Field == tt.field {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected field error for %q, got details: %v", tt.field, svcErr.Details)
			}
		})
	}
}

func TestRegister_PasswordHashed(t *testing.T) {
	svc, db := newAuthService(t)
	defer cleanupUser(t, db, "hash-test@test.com")

	_, err := svc.Register(RegisterRequest{
		Email:    "hash-test@test.com",
		Username: "hashtest",
		Password: "mypassword",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Check DB directly — password should NOT be plaintext
	var user model.User
	db.Where("email = ?", "hash-test@test.com").First(&user)
	if user.PasswordHash == "mypassword" {
		t.Error("password stored as plaintext")
	}
	if user.PasswordHash == "" {
		t.Error("password_hash is empty")
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

```bash
cd backend
go test ./internal/service/ -v
```

Expected: FAIL — `NewAuthService` not defined.

- [ ] **Step 4: Implement auth service**

Create file `backend/internal/service/auth.go`:

```go
package service

import (
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/ArdyJunata/RitualX/backend/internal/repository"
	"github.com/ArdyJunata/RitualX/backend/pkg"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	User         model.User `json:"user"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
}

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(req RegisterRequest) (*RegisterResponse, error) {
	if errs := validateRegister(req); len(errs) > 0 {
		return nil, &ServiceError{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: errs,
		}
	}

	existing, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
	}
	if existing != nil {
		return nil, &ServiceError{Code: "EMAIL_TAKEN", Message: "email already registered"}
	}

	existing, err = s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
	}
	if existing != nil {
		return nil, &ServiceError{Code: "USERNAME_TAKEN", Message: "username already taken"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
	}

	user := &model.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hash),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
	}

	accessToken, err := pkg.GenerateAccessToken(user.ID.String(), s.jwtSecret)
	if err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
	}

	refreshToken, err := pkg.GenerateRefreshToken(user.ID.String(), s.jwtSecret)
	if err != nil {
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "internal error"}
	}

	return &RegisterResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func validateRegister(req RegisterRequest) []FieldError {
	var errs []FieldError

	if !isValidEmail(req.Email) {
		errs = append(errs, FieldError{Field: "email", Message: "invalid email format"})
	}
	if !usernameRegex.MatchString(req.Username) {
		errs = append(errs, FieldError{Field: "username", Message: "must be 3-20 characters, alphanumeric and underscore only"})
	}
	if len(req.Password) < 8 {
		errs = append(errs, FieldError{Field: "password", Message: "must be at least 8 characters"})
	}

	return errs
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	return strings.Contains(parts[1], ".")
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd backend
go test ./internal/service/ -v
```

Expected: All 5 test functions PASS (requires running PostgreSQL).

- [ ] **Step 6: Commit**

```bash
git add backend/internal/service/auth.go backend/internal/service/auth_test.go
git commit -m "feat(backend): add AuthService with registration logic"
```

---


### Task 4: Auth Handler & Route Registration

**Files:**
- Create: `backend/internal/handler/auth.go`
- Create: `backend/internal/handler/auth_test.go`
- Modify: `backend/cmd/server/main.go`

**Interfaces:**
- Consumes:
  - `service.AuthService` from Task 3
  - `service.ServiceError` from Task 2
  - `service.RegisterRequest` from Task 3
- Produces:
  - `handler.Register(authService *service.AuthService) fiber.Handler`
  - Route `POST /api/v1/auth/register` active in server

- [ ] **Step 1: Write failing handler tests**

Create file `backend/internal/handler/auth_test.go`:

```go
package handler

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/ArdyJunata/RitualX/backend/internal/repository"
	"github.com/ArdyJunata/RitualX/backend/internal/service"
)

func setupAuthTest(t *testing.T) (*fiber.App, *gorm.DB) {
	t.Helper()
	dsn := "host=localhost port=5432 user=ritualx password=ritualx_dev dbname=ritualx sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration test")
	}

	app := fiber.New()
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, "test-secret")
	app.Post("/api/v1/auth/register", Register(authService))

	return app, db
}

func cleanupUser(t *testing.T, db *gorm.DB, email string) {
	t.Helper()
	db.Where("email = ?", email).Delete(&model.User{})
}

func TestRegisterHandler_Success(t *testing.T) {
	app, db := setupAuthTest(t)
	defer cleanupUser(t, db, "handler-ok@test.com")

	body := `{"email":"handler-ok@test.com","username":"handlerok","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 201 {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 201, body: %s", resp.StatusCode, respBody)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if result["success"] != true {
		t.Errorf("success = %v, want true", result["success"])
	}

	data := result["data"].(map[string]interface{})
	if data["access_token"] == nil || data["access_token"] == "" {
		t.Error("access_token missing")
	}
	if data["refresh_token"] == nil || data["refresh_token"] == "" {
		t.Error("refresh_token missing")
	}

	user := data["user"].(map[string]interface{})
	if user["email"] != "handler-ok@test.com" {
		t.Errorf("email = %v, want handler-ok@test.com", user["email"])
	}
	if user["username"] != "handlerok" {
		t.Errorf("username = %v, want handlerok", user["username"])
	}
	// password_hash should NOT be in response (json:"-")
	if _, exists := user["password_hash"]; exists {
		t.Error("password_hash leaked in response")
	}
}

func TestRegisterHandler_ValidationError(t *testing.T) {
	app, _ := setupAuthTest(t)

	body := `{"email":"bad","username":"ab","password":"short"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if result["success"] != false {
		t.Errorf("success = %v, want false", result["success"])
	}

	errObj := result["error"].(map[string]interface{})
	if errObj["code"] != "VALIDATION_ERROR" {
		t.Errorf("code = %v, want VALIDATION_ERROR", errObj["code"])
	}
	if errObj["details"] == nil {
		t.Error("details missing from validation error")
	}
}

func TestRegisterHandler_DuplicateEmail(t *testing.T) {
	app, db := setupAuthTest(t)
	defer cleanupUser(t, db, "dup-handler@test.com")

	body := `{"email":"dup-handler@test.com","username":"duphand1","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	app.Test(req)

	// Second request with same email
	body = `{"email":"dup-handler@test.com","username":"duphand2","password":"password123"}`
	req = httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 409 {
		t.Errorf("status = %d, want 409", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	errObj := result["error"].(map[string]interface{})
	if errObj["code"] != "EMAIL_TAKEN" {
		t.Errorf("code = %v, want EMAIL_TAKEN", errObj["code"])
	}
}

func TestRegisterHandler_InvalidJSON(t *testing.T) {
	app, _ := setupAuthTest(t)

	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend
go test ./internal/handler/ -v -run "TestRegisterHandler"
```

Expected: FAIL — `Register` not defined in handler package.

- [ ] **Step 3: Implement auth handler**

Create file `backend/internal/handler/auth.go`:

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
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INVALID_REQUEST",
					"message": "invalid request body",
				},
			})
		}

		resp, err := authService.Register(req)
		if err != nil {
			return handleServiceError(c, err)
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"data":    resp,
		})
	}
}

func handleServiceError(c *fiber.Ctx, err error) error {
	svcErr, ok := err.(*service.ServiceError)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "unexpected error",
			},
		})
	}

	status := mapErrorCodeToStatus(svcErr.Code)

	errResponse := fiber.Map{
		"code":    svcErr.Code,
		"message": svcErr.Message,
	}
	if len(svcErr.Details) > 0 {
		errResponse["details"] = svcErr.Details
	}

	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error":   errResponse,
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

- [ ] **Step 4: Run handler tests**

```bash
cd backend
go test ./internal/handler/ -v
```

Expected: All handler tests PASS (requires running PostgreSQL).

- [ ] **Step 5: Update main.go to register the route**

Modify `backend/cmd/server/main.go` — add import and route registration.

Add to imports:

```go
	"github.com/ArdyJunata/RitualX/backend/internal/repository"
	"github.com/ArdyJunata/RitualX/backend/internal/service"
```

After `api.Get("/health", handler.HealthCheck(db))`, add:

```go
	// Auth routes
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)

	auth := api.Group("/auth")
	auth.Post("/register", handler.Register(authService))
```

The full imports block should be:

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
	"github.com/ArdyJunata/RitualX/backend/internal/repository"
	"github.com/ArdyJunata/RitualX/backend/internal/service"
)
```

- [ ] **Step 6: Verify compilation**

```bash
cd backend
go build ./cmd/server/
```

Expected: compiles without errors.

- [ ] **Step 7: Run all tests**

```bash
cd backend
go test ./... -count=1
```

Expected: All packages pass (config, logger, middleware, handler, service, repository, pkg).

- [ ] **Step 8: Commit**

```bash
git add backend/internal/handler/auth.go backend/internal/handler/auth_test.go backend/cmd/server/main.go
git commit -m "feat(backend): add Register handler and wire route in main.go"
```

---

## Verification Checklist

After all tasks complete, verify each acceptance criterion:

| # | Criterion | How to Verify |
|---|-----------|---------------|
| 1 | POST /auth/register returns 201 | `TestRegisterHandler_Success` |
| 2 | Password bcrypt hashed | `TestRegister_PasswordHashed` |
| 3 | Duplicate email → 409 | `TestRegister_DuplicateEmail` + `TestRegisterHandler_DuplicateEmail` |
| 4 | Duplicate username → 409 | `TestRegister_DuplicateUsername` |
| 5 | Invalid email → 400 | `TestRegister_ValidationErrors` |
| 6 | Short password → 400 | `TestRegister_ValidationErrors` |
| 7 | Invalid username → 400 | `TestRegister_ValidationErrors` |
| 8 | Missing fields → 400 | `TestRegister_ValidationErrors` |
| 9 | Access token 15 min expiry | `TestGenerateAccessToken_Valid` |
| 10 | Refresh token 7 day expiry | `TestGenerateRefreshToken_Valid` |
| 11 | Tokens contain user_id | `TestRegister_Success` verifies claims |
| 12 | No password_hash in response | `TestRegisterHandler_Success` checks json:"-" |
| 13 | Handler→service→repository pattern | Code structure matches |
| 14 | No new migrations | Uses existing users table |
