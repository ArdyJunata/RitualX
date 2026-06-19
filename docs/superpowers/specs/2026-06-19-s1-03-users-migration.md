# S1-03: Database Migration — Users Table

> **Date:** 2026-06-19
> **Status:** Approved
> **Story Points:** 2
> **Sprint:** Sprint 1 — Foundation & Auth

---

## Overview

Create the first versioned database migration for the `users` table using `golang-migrate`. Also define the GORM model for use in subsequent auth stories (S1-04, S1-05).

## Goals

- Versioned SQL migration that creates the `users` table with all columns per PRD data model
- Migration runs forward (`up`) and rolls back (`down`) cleanly
- UUID generation via `uuid-ossp` PostgreSQL extension
- GORM model struct matching the table schema for ORM queries
- Migration pattern established for all future tables

## Non-Goals

- Auto-migration on server boot (migrations are explicit CLI steps)
- Auth logic (S1-04/S1-05)
- Other tables: routines, streaks, etc. (S2-01, S3-01)
- Seeding test data

---

## Design Details

### 1. Migration Tool

**Package:** `github.com/golang-migrate/migrate/v4`

**Installation (CLI):**
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

**Usage:**
```bash
# Run all pending migrations
migrate -path backend/migrations/ -database "postgres://ritualx:ritualx_dev@localhost:5432/ritualx?sslmode=disable" up

# Rollback last migration
migrate -path backend/migrations/ -database "..." down 1

# Check current version
migrate -path backend/migrations/ -database "..." version
```

**File naming convention:** `{sequence}_{description}.{direction}.sql`
- Sequence: 6 digits, zero-padded (`000001`)
- Direction: `up` or `down`

### 2. Migration SQL

**File:** `backend/migrations/000001_create_users.up.sql`

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

**File:** `backend/migrations/000001_create_users.down.sql`

```sql
DROP TABLE IF EXISTS users;
```

**Notes:**
- `uuid-ossp` extension enables `uuid_generate_v4()` for default PK values
- UNIQUE constraints on `email` and `username` already create implicit indexes, but explicit named indexes improve query plan readability
- `TIMESTAMP WITH TIME ZONE` stores UTC, converts on retrieval per connection timezone

### 3. GORM Model

**File:** `backend/internal/model/user.go`

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

**Key decisions:**
- `PasswordHash` has `json:"-"` to never serialize to API responses
- `uuid.UUID` type from `github.com/google/uuid` (already in go.mod)
- JSON tags follow snake_case to match API convention
- GORM tags document the schema but won't be used for auto-migration

### 4. Makefile Target (convenience)

**File:** `backend/Makefile` (create new)

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

**Usage:**
```bash
make migrate-up
make migrate-down
make migrate-create name=create_routines
```

---

## Acceptance Criteria

- [ ] `backend/migrations/000001_create_users.up.sql` creates `users` table with all columns per data model
- [ ] `backend/migrations/000001_create_users.down.sql` drops the table cleanly
- [ ] `migrate ... up` runs without error on a clean database
- [ ] `migrate ... down 1` removes the table without error
- [ ] Running `up` again after `down` succeeds (idempotent extension creation)
- [ ] UUID generation works — inserting a row without specifying `id` auto-generates a UUID
- [ ] GORM model in `internal/model/user.go` matches the table schema
- [ ] `PasswordHash` field is excluded from JSON serialization
- [ ] Makefile provides `migrate-up` and `migrate-down` targets

---

## Open Questions

None.
