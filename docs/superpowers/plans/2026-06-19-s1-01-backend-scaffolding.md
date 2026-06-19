# S1-01: Backend Scaffolding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Scaffold a working Go Fiber backend with GORM/PostgreSQL, structured logging, and health check endpoint.

**Architecture:** Single Go binary (`cmd/server/main.go`) loads config from `.env`, initializes slog JSON logger, connects to PostgreSQL via GORM, registers a health check route, and starts Fiber with graceful shutdown. Dev PostgreSQL runs via Docker Compose.

**Tech Stack:** Go 1.23+, Fiber v2, GORM, PostgreSQL 16, slog, godotenv

## Global Constraints

- Go module: `github.com/ArdyJunata/RitualX/backend`
- All code under `backend/` directory
- JSON responses follow: `{"success": true, "data": {...}}` or `{"success": false, "error": {"code": "...", "message": "..."}}`
- No auto-migration, no CORS, no auth middleware in this story
- Branch: `feat/s1-01-backend-scaffolding`

---

### Task 1: Project Init, Config & Docker Compose

**Files:**
- Create: `backend/go.mod`
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/config/config_test.go`
- Create: `backend/.env.example`
- Create: `backend/docker-compose.dev.yml`

**Interfaces:**
- Consumes: nothing
- Produces: `config.Load() (*Config, error)`, `Config` struct with fields: `AppPort`, `DBHost`, `DBPort`, `DBUser`, `DBPassword`, `DBName`, `JWTSecret`, `LogLevel`

- [ ] **Step 1: Initialize Go module and install dependencies**

```bash
cd backend
go mod init github.com/ArdyJunata/RitualX/backend
go get github.com/gofiber/fiber/v2
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/joho/godotenv
go get github.com/google/uuid
```

- [ ] **Step 2: Create `.env.example`**

Create file `backend/.env.example`:

```
APP_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=ritualx
DB_PASSWORD=ritualx_dev
DB_NAME=ritualx
JWT_SECRET=change-me-in-production
LOG_LEVEL=info
```

- [ ] **Step 3: Create `docker-compose.dev.yml`**

Create file `backend/docker-compose.dev.yml`:

```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: ritualx-postgres
    environment:
      POSTGRES_USER: ${DB_USER:-ritualx}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-ritualx_dev}
      POSTGRES_DB: ${DB_NAME:-ritualx}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

- [ ] **Step 4: Write failing test for config.Load()**

Create file `backend/internal/config/config_test.go`:

```go
package config

import (
	"os"
	"testing"
)

func TestLoad_WithAllEnvVars(t *testing.T) {
	os.Setenv("APP_PORT", "9090")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("JWT_SECRET", "testsecret")
	os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("APP_PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.AppPort != "9090" {
		t.Errorf("AppPort = %q, want %q", cfg.AppPort, "9090")
	}
	if cfg.DBHost != "localhost" {
		t.Errorf("DBHost = %q, want %q", cfg.DBHost, "localhost")
	}
	if cfg.DBPort != "5433" {
		t.Errorf("DBPort = %q, want %q", cfg.DBPort, "5433")
	}
	if cfg.DBUser != "testuser" {
		t.Errorf("DBUser = %q, want %q", cfg.DBUser, "testuser")
	}
	if cfg.DBPassword != "testpass" {
		t.Errorf("DBPassword = %q, want %q", cfg.DBPassword, "testpass")
	}
	if cfg.DBName != "testdb" {
		t.Errorf("DBName = %q, want %q", cfg.DBName, "testdb")
	}
	if cfg.JWTSecret != "testsecret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "testsecret")
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
}

func TestLoad_Defaults(t *testing.T) {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("JWT_SECRET", "testsecret")
	os.Unsetenv("APP_PORT")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("LOG_LEVEL")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.AppPort != "8080" {
		t.Errorf("AppPort = %q, want default %q", cfg.AppPort, "8080")
	}
	if cfg.DBPort != "5432" {
		t.Errorf("DBPort = %q, want default %q", cfg.DBPort, "5432")
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want default %q", cfg.LogLevel, "info")
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	os.Clearenv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing required vars, got nil")
	}
}
```

- [ ] **Step 5: Run test to verify it fails**

```bash
cd backend
go test ./internal/config/ -v
```

Expected: FAIL — `Load` not defined.

- [ ] **Step 6: Implement config.Load()**

Create file `backend/internal/config/config.go`:

```go
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort    string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	LogLevel   string
}

func Load() (*Config, error) {
	_ = godotenv.Load() // non-fatal if .env missing

	cfg := &Config{
		AppPort:    getEnvOrDefault("APP_PORT", "8080"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     getEnvOrDefault("DB_PORT", "5432"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		LogLevel:   getEnvOrDefault("LOG_LEVEL", "info"),
	}

	if cfg.DBHost == "" {
		return nil, fmt.Errorf("required env var DB_HOST is not set")
	}
	if cfg.DBUser == "" {
		return nil, fmt.Errorf("required env var DB_USER is not set")
	}
	if cfg.DBPassword == "" {
		return nil, fmt.Errorf("required env var DB_PASSWORD is not set")
	}
	if cfg.DBName == "" {
		return nil, fmt.Errorf("required env var DB_NAME is not set")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("required env var JWT_SECRET is not set")
	}

	return cfg, nil
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 7: Run tests to verify they pass**

```bash
cd backend
go test ./internal/config/ -v
```

Expected: All 3 tests PASS.

- [ ] **Step 8: Commit**

```bash
git add backend/go.mod backend/go.sum backend/.env.example backend/docker-compose.dev.yml backend/internal/config/
git commit -m "feat(backend): add project init, config loading & dev docker-compose"
```

---

### Task 2: Structured Logger (slog)

**Files:**
- Create: `backend/internal/logger/logger.go`
- Create: `backend/internal/logger/logger_test.go`

**Interfaces:**
- Consumes: `config.Config.LogLevel` (string: "debug"|"info"|"warn"|"error")
- Produces:
  - `logger.Init(level string)`
  - `logger.Get() *slog.Logger`
  - `logger.FromContext(ctx context.Context) *slog.Logger`
  - `logger.WithTraceID(ctx context.Context, id string) context.Context`

- [ ] **Step 1: Write failing tests for logger**

Create file `backend/internal/logger/logger_test.go`:

```go
package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestInit_SetsLogLevel(t *testing.T) {
	Init("warn")
	l := Get()
	if l == nil {
		t.Fatal("Get() returned nil after Init()")
	}
	if !l.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("expected WARN level to be enabled")
	}
	if l.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("expected INFO level to be disabled at WARN threshold")
	}
}

func TestInit_InvalidLevelDefaultsToInfo(t *testing.T) {
	Init("invalid")
	l := Get()
	if !l.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("expected INFO level enabled for invalid input")
	}
	if l.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("expected DEBUG disabled when defaulting to INFO")
	}
}

func TestGet_BeforeInit_ReturnsFallback(t *testing.T) {
	defaultLogger = nil
	l := Get()
	if l == nil {
		t.Fatal("Get() returned nil before Init()")
	}
}

func TestFromContext_WithTraceID(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	defaultLogger = slog.New(handler)

	ctx := WithTraceID(context.Background(), "test-trace-123")
	l := FromContext(ctx)
	l.Info("test message")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log JSON: %v", err)
	}
	if entry["trace_id"] != "test-trace-123" {
		t.Errorf("trace_id = %v, want %q", entry["trace_id"], "test-trace-123")
	}
}

func TestFromContext_WithoutTraceID(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	defaultLogger = slog.New(handler)

	l := FromContext(context.Background())
	l.Info("no trace")

	var entry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log JSON: %v", err)
	}
	if _, exists := entry["trace_id"]; exists {
		t.Error("expected no trace_id in log entry")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend
go test ./internal/logger/ -v
```

Expected: FAIL — package has no Go files / functions not defined.

- [ ] **Step 3: Implement logger package**

Create file `backend/internal/logger/logger.go`:

```go
package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type ctxKey string

const traceIDKey ctxKey = "trace_id"

var defaultLogger *slog.Logger

func Init(level string) {
	parsedLevel := parseLevel(level)
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parsedLevel,
	})
	defaultLogger = slog.New(handler)
}

func Get() *slog.Logger {
	if defaultLogger == nil {
		return slog.Default()
	}
	return defaultLogger
}

func FromContext(ctx context.Context) *slog.Logger {
	l := Get()
	if traceID, ok := ctx.Value(traceIDKey).(string); ok && traceID != "" {
		return l.With("trace_id", traceID)
	}
	return l
}

func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey, id)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend
go test ./internal/logger/ -v
```

Expected: All 5 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/logger/
git commit -m "feat(backend): add structured JSON logger with trace_id context"
```

---

### Task 3: Health Check Handler

**Files:**
- Create: `backend/internal/handler/health.go`
- Create: `backend/internal/handler/health_test.go`

**Interfaces:**
- Consumes: `*gorm.DB` (passed as dependency)
- Produces: `handler.HealthCheck(db *gorm.DB) fiber.Handler`

- [ ] **Step 1: Write failing test for health check handler**

Create file `backend/internal/handler/health_test.go`:

```go
package handler

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

func TestHealthCheck_Healthy(t *testing.T) {
	db := setupTestDB(t)

	app := fiber.New()
	app.Get("/api/v1/health", HealthCheck(db))

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["success"] != true {
		t.Errorf("success = %v, want true", result["success"])
	}
	data := result["data"].(map[string]interface{})
	if data["status"] != "healthy" {
		t.Errorf("status = %v, want healthy", data["status"])
	}
	if data["version"] != "0.1.0" {
		t.Errorf("version = %v, want 0.1.0", data["version"])
	}
}

func TestHealthCheck_Unhealthy(t *testing.T) {
	// Use invalid DSN to simulate DB failure
	dsn := "host=localhost port=9999 user=invalid password=invalid dbname=invalid sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("Cannot create gorm instance for unhealthy test")
	}

	app := fiber.New()
	app.Get("/api/v1/health", HealthCheck(db))

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != 503 {
		t.Errorf("status = %d, want 503", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["success"] != false {
		t.Errorf("success = %v, want false", result["success"])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend
go test ./internal/handler/ -v
```

Expected: FAIL — `HealthCheck` not defined.

- [ ] **Step 3: Implement health check handler**

Create file `backend/internal/handler/health.go`:

```go
package handler

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func HealthCheck(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var result int
		if err := db.Raw("SELECT 1").Scan(&result).Error; err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "DB_UNHEALTHY",
					"message": "database connection failed",
				},
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"status":  "healthy",
				"version": "0.1.0",
			},
		})
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd backend
docker compose -f docker-compose.dev.yml up -d
go test ./internal/handler/ -v
```

Expected: `TestHealthCheck_Healthy` PASS (requires running postgres). `TestHealthCheck_Unhealthy` PASS or SKIP.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handler/
git commit -m "feat(backend): add health check handler with DB ping"
```

---

### Task 4: Main Entry Point & Directory Structure

**Files:**
- Create: `backend/cmd/server/main.go`
- Create: `backend/internal/middleware/.gitkeep`
- Create: `backend/internal/model/.gitkeep`
- Create: `backend/internal/service/.gitkeep`
- Create: `backend/internal/repository/.gitkeep`
- Create: `backend/internal/engine/.gitkeep`
- Create: `backend/migrations/.gitkeep`
- Create: `backend/pkg/.gitkeep`

**Interfaces:**
- Consumes: `config.Load()`, `logger.Init()`, `logger.Get()`, `handler.HealthCheck()`
- Produces: Running Fiber server on configured port with graceful shutdown

- [ ] **Step 1: Create placeholder directories**

Create `.gitkeep` files in each empty directory:

```bash
cd backend
mkdir -p internal/middleware internal/model internal/service internal/repository internal/engine migrations pkg
touch internal/middleware/.gitkeep internal/model/.gitkeep internal/service/.gitkeep internal/repository/.gitkeep internal/engine/.gitkeep migrations/.gitkeep pkg/.gitkeep
```

- [ ] **Step 2: Implement main.go**

Create file `backend/cmd/server/main.go`:

```go
package main

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
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.LogLevel)
	log := logger.Get()

	log.Info("starting server", "port", cfg.AppPort)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	log.Info("database connected")

	app := fiber.New(fiber.Config{})

	api := app.Group("/api/v1")
	api.Get("/health", handler.HealthCheck(db))

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info("shutting down server")
		_ = app.Shutdown()
	}()

	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 3: Verify the server starts and health check works**

```bash
cd backend
cp .env.example .env
docker compose -f docker-compose.dev.yml up -d
go run cmd/server/main.go
```

In a separate terminal:

```bash
curl http://localhost:8080/api/v1/health
```

Expected response:

```json
{"success":true,"data":{"status":"healthy","version":"0.1.0"}}
```

Stop the server with Ctrl+C — should see "shutting down server" log and exit cleanly.

- [ ] **Step 4: Verify structured JSON logging output**

When server starts, stdout should show JSON lines like:

```json
{"time":"...","level":"INFO","msg":"starting server","port":"8080"}
{"time":"...","level":"INFO","msg":"database connected"}
```

- [ ] **Step 5: Run all tests**

```bash
cd backend
go test ./... -v
```

Expected: All tests in `config`, `logger`, `handler` packages pass.

- [ ] **Step 6: Commit**

```bash
git add backend/cmd/ backend/internal/middleware/ backend/internal/model/ backend/internal/service/ backend/internal/repository/ backend/internal/engine/ backend/migrations/ backend/pkg/
git commit -m "feat(backend): add main entry point with Fiber server & graceful shutdown"
```

---

## Verification Checklist

After all tasks complete, verify each acceptance criterion:

| # | Criterion | How to Verify |
|---|-----------|---------------|
| 1 | Fiber starts on configured port | `APP_PORT=9090 go run cmd/server/main.go` → listens on 9090 |
| 2 | GORM connects to PostgreSQL | Server starts without "database connection failed" error |
| 3 | Health check returns 200/503 | `curl localhost:8080/api/v1/health` with DB up/down |
| 4 | Folder structure correct | `ls` — cmd/, internal/, pkg/, migrations/ all present |
| 5 | Config loads from .env | Modify .env values, restart, see them reflected |
| 6 | Missing vars produce error | Unset `DB_PASSWORD`, run → "required env var DB_PASSWORD is not set" |
| 7 | slog initialized | JSON logs appear on stdout |
| 8 | LOG_LEVEL works | `LOG_LEVEL=debug` shows debug logs, `LOG_LEVEL=error` hides info |
| 9 | `logger.Get()` works | Used in main.go successfully |
| 10 | `logger.FromContext()` works | Unit test passes with trace_id extraction |
| 11 | Docker postgres starts | `docker compose -f docker-compose.dev.yml up -d` → `docker ps` shows container |
| 12 | Graceful shutdown | Ctrl+C → "shutting down server" log, clean exit |
