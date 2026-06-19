# S1-01: Backend Scaffolding — Go Fiber, GORM & PostgreSQL

> **Date:** 2026-06-19
> **Status:** Approved
> **Story Points:** 3
> **Sprint:** Sprint 1 — Foundation & Auth

---

## Overview

Scaffold the backend project with Go Fiber as the HTTP framework, GORM as the ORM, and PostgreSQL as the database. Includes structured JSON logging via `slog`, environment-based configuration, a health check endpoint, and a dev-only Docker Compose for PostgreSQL.

## Goals

- Working Go Fiber server that starts and connects to PostgreSQL
- Clean project structure following `cmd/`, `internal/`, `pkg/` convention
- Structured JSON logger with configurable log level and trace_id context propagation
- Health check endpoint that verifies DB connectivity
- Developer can `docker compose up` postgres and `go run` the server immediately

## Non-Goals

- Database migrations (S1-03)
- Auth middleware (S1-04/S1-05)
- Trace middleware (S1-08)
- Full Docker Compose with all services (S1-09)
- CORS configuration (will be added when frontend connects)

---

## Design Details

### 1. Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Env loading, Config struct
│   ├── logger/
│   │   └── logger.go            # slog JSON setup
│   ├── middleware/              # (empty — placeholder for S1-08)
│   ├── model/                   # (empty — placeholder for S1-03)
│   ├── handler/
│   │   └── health.go            # GET /api/v1/health
│   ├── service/                 # (empty — placeholder)
│   ├── repository/              # (empty — placeholder)
│   └── engine/                  # (empty — placeholder)
├── migrations/                  # (empty — placeholder for S1-03)
├── pkg/                         # (empty — placeholder)
├── go.mod
├── go.sum
├── .env.example
└── docker-compose.dev.yml       # Dev-only: PostgreSQL
```

**Go module:** `github.com/ArdyJunata/RitualX/backend`
**Go version:** 1.23+

### 2. Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/gofiber/fiber/v2` | latest | HTTP framework |
| `gorm.io/gorm` | latest | ORM |
| `gorm.io/driver/postgres` | latest | PostgreSQL driver for GORM |
| `github.com/joho/godotenv` | latest | .env file loading |
| `github.com/google/uuid` | latest | UUID generation |

### 3. Config (`internal/config/config.go`)

```go
type Config struct {
    AppPort    string // env: APP_PORT, default: "8080"
    DBHost     string // env: DB_HOST, required
    DBPort     string // env: DB_PORT, default: "5432"
    DBUser     string // env: DB_USER, required
    DBPassword string // env: DB_PASSWORD, required
    DBName     string // env: DB_NAME, required
    JWTSecret  string // env: JWT_SECRET, required
    LogLevel   string // env: LOG_LEVEL, default: "info"
}
```

**Exported function:**

```go
func Load() (*Config, error)
```

**Behavior:**
- Calls `godotenv.Load()` — non-fatal if `.env` missing (supports pure env vars in containers)
- Reads values from `os.Getenv()` with defaults for `AppPort`, `DBPort`, `LogLevel`
- Returns error if any required field is empty: `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`

### 4. Logger (`internal/logger/logger.go`)

**Exported API:**

```go
func Init(level string)                                       // Initialize the global logger
func Get() *slog.Logger                                       // Return the global logger
func FromContext(ctx context.Context) *slog.Logger            // Logger with trace_id from context
func WithTraceID(ctx context.Context, id string) context.Context // Inject trace_id into context
```

**Internals:**
- Unexported context key: `type ctxKey string`, const `traceIDKey ctxKey = "trace_id"`
- Package-level `var defaultLogger *slog.Logger`

**`Init(level string)` behavior:**
- Parses level string to `slog.Level`: "debug" → -4, "info" → 0, "warn" → 4, "error" → 8
- Invalid/unknown level defaults to INFO
- Creates handler: `slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parsedLevel})`
- Assigns to `defaultLogger`

**`Get()` behavior:**
- Returns `defaultLogger`
- If called before `Init()`, returns `slog.Default()` (safety fallback)

**`FromContext(ctx)` behavior:**
- Extracts `trace_id` from context via `ctx.Value(traceIDKey)`
- If found: returns `defaultLogger.With("trace_id", traceID)`
- If not found: returns `defaultLogger`

**`WithTraceID(ctx, id)` behavior:**
- Returns `context.WithValue(ctx, traceIDKey, id)`

### 5. Database Connection

Located in `cmd/server/main.go` (inline, not extracted — keep it simple for now).

**DSN format:**
```
host=%s port=%s user=%s password=%s dbname=%s sslmode=disable
```

**Connection:**
```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
```

**On success:** Log `"database connected"` at INFO level
**On failure:** Log error at ERROR level, `os.Exit(1)`

No auto-migration. No connection pooling config (GORM defaults are fine for dev).

### 6. Health Check (`internal/handler/health.go`)

**Endpoint:** `GET /api/v1/health`
**Auth:** None required

**Handler signature:**
```go
func HealthCheck(db *gorm.DB) fiber.Handler
```

**Logic:**
1. Ping DB: `db.Raw("SELECT 1").Scan(&result)`
2. If success → 200:
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "version": "0.1.0"
  }
}
```
3. If DB error → 503:
```json
{
  "success": false,
  "error": {
    "code": "DB_UNHEALTHY",
    "message": "database connection failed"
  }
}
```

### 7. Boot Sequence (`cmd/server/main.go`)

```
1. config.Load() → fatal if error
2. logger.Init(cfg.LogLevel)
3. Log "starting server" with port
4. Connect to PostgreSQL via GORM → fatal if error
5. fiber.New(fiber.Config{})
6. Register route: GET /api/v1/health
7. Graceful shutdown goroutine: listen os.Interrupt/SIGTERM → app.Shutdown()
8. app.Listen(":" + cfg.AppPort)
```

### 8. PostgreSQL Docker Compose (`docker-compose.dev.yml`)

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

### 9. `.env.example`

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

---

## Acceptance Criteria

- [ ] Go Fiber server starts on configured port
- [ ] GORM connects to PostgreSQL
- [ ] Health check endpoint `/api/v1/health` returns 200 (healthy) or 503 (unhealthy)
- [ ] Project follows the defined folder structure (`cmd/`, `internal/`, `pkg/`)
- [ ] Environment config loads from `.env`
- [ ] Required env vars validated — missing vars produce clear error message
- [ ] Structured JSON logger (`slog`) initialized in `internal/logger/`
- [ ] `LOG_LEVEL` env var controls log verbosity (debug/info/warn/error)
- [ ] `logger.Get()` returns configured `*slog.Logger`
- [ ] `logger.FromContext(ctx)` extracts logger with `trace_id`
- [ ] `docker compose -f docker-compose.dev.yml up -d` starts PostgreSQL
- [ ] Graceful shutdown on SIGINT/SIGTERM

---

## Open Questions

None.
