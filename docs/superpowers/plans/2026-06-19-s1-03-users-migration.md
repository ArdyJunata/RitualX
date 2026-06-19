# S1-03: Users Table Migration — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create versioned SQL migration for the `users` table and a matching GORM model struct.

**Architecture:** Raw SQL migration files managed by `golang-migrate` CLI. GORM model defined for ORM queries in later stories. Makefile provides convenience targets.

**Tech Stack:** golang-migrate v4, PostgreSQL 16, GORM, github.com/google/uuid

## Global Constraints

- Go module: `github.com/ArdyJunata/RitualX/backend`
- All code under `backend/` directory
- Migrations live in `backend/migrations/` as versioned SQL files
- No auto-migration on server boot
- Branch: `feat/s1-03-users-migration`
- PostgreSQL must be running: `docker compose -f backend/docker-compose.dev.yml up -d`

---

### Task 1: Migration SQL Files

**Files:**
- Create: `backend/migrations/000001_create_users.up.sql`
- Create: `backend/migrations/000001_create_users.down.sql`
- Delete: `backend/migrations/.gitkeep`

**Interfaces:**
- Consumes: PostgreSQL database (running via docker-compose.dev.yml)
- Produces: `users` table with all columns, indexes, and UUID generation

- [ ] **Step 1: Create the UP migration**

Create file `backend/migrations/000001_create_users.up.sql`:

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    username VARCHAR(20) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    avatar_url VARCHAR(500),
    xp INTEGER NOT NULL DEFAULT 0,
    level INTEGER NOT NULL DEFAULT 1,
    coins INTEGER NOT NULL DEFAULT 0,
    title VARCHAR(50) NOT NULL DEFAULT 'Novice',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
```

- [ ] **Step 2: Create the DOWN migration**

Create file `backend/migrations/000001_create_users.down.sql`:

```sql
DROP TABLE IF EXISTS users;
```

- [ ] **Step 3: Delete the placeholder .gitkeep**

```bash
cd backend
rm migrations/.gitkeep
```

- [ ] **Step 4: Install golang-migrate CLI (if not already installed)**

```bash
go install -tags "postgres" github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Verify:

```bash
migrate -version
```

Expected: prints version number (e.g., `4.17.0`).

- [ ] **Step 5: Run migration UP**

```bash
cd backend
migrate -path migrations/ -database "postgres://ritualx:ritualx_dev@localhost:5432/ritualx?sslmode=disable" up
```

Expected output: `1/u create_users (Xms)`

- [ ] **Step 6: Verify table was created**

```bash
docker exec -it ritualx-postgres psql -U ritualx -d ritualx -c "\d users"
```

Expected: table columns listed matching the schema (id, email, password_hash, username, display_name, avatar_url, xp, level, coins, title, created_at, updated_at).

- [ ] **Step 7: Verify UUID generation works**

```bash
docker exec -it ritualx-postgres psql -U ritualx -d ritualx -c "INSERT INTO users (email, password_hash, username) VALUES ('test@example.com', 'hash123', 'testuser') RETURNING id;"
```

Expected: returns a UUID like `a1b2c3d4-e5f6-...`.

Clean up:

```bash
docker exec -it ritualx-postgres psql -U ritualx -d ritualx -c "DELETE FROM users WHERE email='test@example.com';"
```

- [ ] **Step 8: Run migration DOWN**

```bash
cd backend
migrate -path migrations/ -database "postgres://ritualx:ritualx_dev@localhost:5432/ritualx?sslmode=disable" down 1
```

Expected output: `1/d create_users (Xms)`

- [ ] **Step 9: Verify table is gone**

```bash
docker exec -it ritualx-postgres psql -U ritualx -d ritualx -c "\d users"
```

Expected: `Did not find any relation named "users".`

- [ ] **Step 10: Run UP again to leave DB in ready state**

```bash
cd backend
migrate -path migrations/ -database "postgres://ritualx:ritualx_dev@localhost:5432/ritualx?sslmode=disable" up
```

Expected: succeeds (idempotent `CREATE EXTENSION IF NOT EXISTS`).

- [ ] **Step 11: Commit**

```bash
git add backend/migrations/
git commit -m "feat(backend): add users table migration (000001)"
```

---

### Task 2: GORM Model & Makefile

**Files:**
- Create: `backend/internal/model/user.go`
- Create: `backend/Makefile`
- Delete: `backend/internal/model/.gitkeep`

**Interfaces:**
- Consumes: `github.com/google/uuid` (already in go.mod)
- Produces: `model.User` struct used by repository/service layers in S1-04, S1-05

- [ ] **Step 1: Create the User model**

Create file `backend/internal/model/user.go`:

```go
package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Username     string    `gorm:"type:varchar(20);uniqueIndex;not null" json:"username"`
	DisplayName  string    `gorm:"type:varchar(100)" json:"display_name"`
	AvatarURL    string    `gorm:"type:varchar(500)" json:"avatar_url"`
	XP           int       `gorm:"not null;default:0" json:"xp"`
	Level        int       `gorm:"not null;default:1" json:"level"`
	Coins        int       `gorm:"not null;default:0" json:"coins"`
	Title        string    `gorm:"type:varchar(50);not null;default:'Novice'" json:"title"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
```

- [ ] **Step 2: Delete the placeholder .gitkeep**

```bash
cd backend
rm internal/model/.gitkeep
```

- [ ] **Step 3: Verify model compiles**

```bash
cd backend
go build ./internal/model/
```

Expected: no errors.

- [ ] **Step 4: Create Makefile**

Create file `backend/Makefile`:

```makefile
DB_URL ?= postgres://ritualx:ritualx_dev@localhost:5432/ritualx?sslmode=disable

.PHONY: migrate-up migrate-down migrate-create

migrate-up:
	migrate -path migrations/ -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations/ -database "$(DB_URL)" down 1

migrate-create:
	migrate create -ext sql -dir migrations/ -seq $(name)
```

- [ ] **Step 5: Test Makefile targets**

```bash
cd backend
make migrate-down
make migrate-up
```

Expected: migration runs down then up successfully.

- [ ] **Step 6: Run all backend tests to ensure nothing broke**

```bash
cd backend
go test ./... -v
```

Expected: all existing tests (config: 3, logger: 5, handler: 2 skip) still pass.

- [ ] **Step 7: Commit**

```bash
git add backend/internal/model/user.go backend/Makefile
git commit -m "feat(backend): add User GORM model and Makefile migration targets"
```

---

## Verification Checklist

| # | Criterion | How to Verify |
|---|-----------|---------------|
| 1 | UP migration creates users table | `make migrate-up` + `\d users` in psql |
| 2 | DOWN migration drops table | `make migrate-down` + `\d users` → not found |
| 3 | UP after DOWN works | `make migrate-down && make migrate-up` succeeds |
| 4 | UUID auto-generates | INSERT without id → returns UUID |
| 5 | All columns present | `\d users` shows 12 columns with correct types |
| 6 | Indexes created | `\di` shows idx_users_email, idx_users_username |
| 7 | GORM model compiles | `go build ./internal/model/` no errors |
| 8 | PasswordHash hidden from JSON | `json:"-"` tag on field |
| 9 | Makefile works | `make migrate-up`, `make migrate-down` succeed |
| 10 | Existing tests still pass | `go test ./...` all pass |
