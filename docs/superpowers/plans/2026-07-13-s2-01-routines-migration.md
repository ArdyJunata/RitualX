# S2-01 Routines & Routine Logs Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create migrations `000003_create_routines` and `000004_create_routine_logs`, plus GORM models `Routine` and `RoutineLog` in `backend/internal/model/`.

**Architecture:** Two sequential SQL migration pairs (up/down). Two GORM model files following existing conventions (`user.go`, `refresh_token.go`). No application logic — pure data layer.

**Tech Stack:** PostgreSQL 16, golang-migrate, GORM v2, `github.com/google/uuid`.

## Global Constraints

- Migration files: `backend/migrations/NNNNNN_<name>.up.sql` and `.down.sql`
- GORM models: `backend/internal/model/<name>.go`, `package model`
- UUID PK: `uuid_generate_v4()` (uuid-ossp already enabled by migration 000001)
- Timestamps: `TIMESTAMP WITH TIME ZONE` in SQL; `time.Time` in Go
- `period_type` ENUM values: `'daily'`, `'weekly'`, `'monthly'` — exact strings
- No commits until explicitly instructed
- Branch: `feat/s2-01-routines-migration`

---

### Task 1: Create feature branch

**Files:** none

- [ ] **Step 1: Create branch**

```powershell
git checkout -b feat/s2-01-routines-migration
```

Expected: `Switched to a new branch 'feat/s2-01-routines-migration'`

---

### Task 2: Migration — create routines table

**Files:**
- Create: `backend/migrations/000003_create_routines.up.sql`
- Create: `backend/migrations/000003_create_routines.down.sql`

- [ ] **Step 1: Create `000003_create_routines.up.sql`**

```sql
CREATE TYPE period_type AS ENUM ('daily', 'weekly', 'monthly');

CREATE TABLE routines (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title        VARCHAR(100) NOT NULL,
    description  TEXT,
    period_type  period_type NOT NULL,
    target_count INTEGER NOT NULL DEFAULT 1,
    icon         VARCHAR(50),
    color        VARCHAR(20),
    is_active    BOOLEAN NOT NULL DEFAULT true,
    sort_order   INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_routines_user_id ON routines(user_id);
CREATE INDEX idx_routines_user_active ON routines(user_id, is_active);
```

- [ ] **Step 2: Create `000003_create_routines.down.sql`**

```sql
DROP TABLE IF EXISTS routines;
DROP TYPE IF EXISTS period_type;
```

- [ ] **Step 3: Verify files exist**

```powershell
Get-ChildItem backend/migrations/ | Select-Object Name
```

Expected: 6 files including `000003_create_routines.up.sql` and `000003_create_routines.down.sql`

---

### Task 3: Migration — create routine_logs table

**Files:**
- Create: `backend/migrations/000004_create_routine_logs.up.sql`
- Create: `backend/migrations/000004_create_routine_logs.down.sql`

- [ ] **Step 1: Create `000004_create_routine_logs.up.sql`**

```sql
CREATE TABLE routine_logs (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    routine_id  UUID NOT NULL REFERENCES routines(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    logged_at   DATE NOT NULL,
    count       INTEGER NOT NULL DEFAULT 1,
    note        TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_routine_logs_unique ON routine_logs(routine_id, logged_at);
CREATE INDEX idx_routine_logs_user_id ON routine_logs(user_id);
CREATE INDEX idx_routine_logs_logged_at ON routine_logs(logged_at);
```

- [ ] **Step 2: Create `000004_create_routine_logs.down.sql`**

```sql
DROP TABLE IF EXISTS routine_logs;
```

---

### Task 4: GORM model — Routine

**Files:**
- Create: `backend/internal/model/routine.go`

- [ ] **Step 1: Create `backend/internal/model/routine.go`**

```go
package model

import (
	"time"

	"github.com/google/uuid"
)

type Routine struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null"                               json:"user_id"`
	Title       string    `gorm:"type:varchar(100);not null"                       json:"title"`
	Description string    `gorm:"type:text"                                        json:"description"`
	PeriodType  string    `gorm:"type:period_type;not null"                        json:"period_type"`
	TargetCount int       `gorm:"not null;default:1"                               json:"target_count"`
	Icon        string    `gorm:"type:varchar(50)"                                 json:"icon"`
	Color       string    `gorm:"type:varchar(20)"                                 json:"color"`
	IsActive    bool      `gorm:"not null;default:true"                            json:"is_active"`
	SortOrder   int       `gorm:"not null;default:0"                               json:"sort_order"`
	CreatedAt   time.Time `                                                        json:"created_at"`
	UpdatedAt   time.Time `                                                        json:"updated_at"`
}
```

- [ ] **Step 2: Verify Go compiles**

```powershell
go build ./...
```

Run from `backend/` directory.
Expected: no errors, no output

---

### Task 5: GORM model — RoutineLog

**Files:**
- Create: `backend/internal/model/routine_log.go`

- [ ] **Step 1: Create `backend/internal/model/routine_log.go`**

```go
package model

import (
	"time"

	"github.com/google/uuid"
)

type RoutineLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	RoutineID uuid.UUID `gorm:"type:uuid;not null"                               json:"routine_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"                               json:"user_id"`
	LoggedAt  time.Time `gorm:"type:date;not null"                               json:"logged_at"`
	Count     int       `gorm:"not null;default:1"                               json:"count"`
	Note      string    `gorm:"type:text"                                        json:"note"`
	CreatedAt time.Time `                                                        json:"created_at"`
}
```

- [ ] **Step 2: Verify Go compiles**

```powershell
go build ./...
```

Run from `backend/` directory.
Expected: no errors, no output

---

### Task 6: Run migrations against dev DB

- [ ] **Step 1: Start dev DB**

```powershell
docker compose -f backend/docker-compose.dev.yml up -d
```

Expected: `ritualx-postgres` container running

- [ ] **Step 2: Wait for postgres to be ready**

```powershell
Start-Sleep -Seconds 3
docker exec ritualx-postgres pg_isready -U ritualx -d ritualx
```

Expected: `ritualx:5432 - accepting connections`

- [ ] **Step 3: Run all pending migrations**

```powershell
docker run --rm --network host `
  -v "${PWD}/backend/migrations:/migrations" `
  migrate/migrate `
  -path=/migrations `
  -database "postgres://ritualx:ritualx_dev@localhost:5432/ritualx?sslmode=disable" `
  up
```

Expected:
```
000003/u create_routines (Xms)
000004/u create_routine_logs (Xms)
```

- [ ] **Step 4: Verify routines table**

```powershell
docker exec ritualx-postgres psql -U ritualx -d ritualx -c "\d routines"
```

Expected: table description with all columns (id, user_id, title, description, period_type, target_count, icon, color, is_active, sort_order, created_at, updated_at)

- [ ] **Step 5: Verify routine_logs table**

```powershell
docker exec ritualx-postgres psql -U ritualx -d ritualx -c "\d routine_logs"
```

Expected: table description with all columns (id, routine_id, user_id, logged_at, count, note, created_at)

- [ ] **Step 6: Verify UNIQUE constraint**

```powershell
docker exec ritualx-postgres psql -U ritualx -d ritualx -c "\d+ routine_logs"
```

Expected: `idx_routine_logs_unique` unique index on `(routine_id, logged_at)` visible

- [ ] **Step 7: Verify period_type ENUM**

```powershell
docker exec ritualx-postgres psql -U ritualx -d ritualx -c "SELECT enum_range(NULL::period_type);"
```

Expected: `{daily,weekly,monthly}`

---

### Task 7: Test rollback

- [ ] **Step 1: Roll back migration 000004**

```powershell
docker run --rm --network host `
  -v "${PWD}/backend/migrations:/migrations" `
  migrate/migrate `
  -path=/migrations `
  -database "postgres://ritualx:ritualx_dev@localhost:5432/ritualx?sslmode=disable" `
  down 1
```

Expected: `000004/d create_routine_logs`

- [ ] **Step 2: Verify routine_logs dropped**

```powershell
docker exec ritualx-postgres psql -U ritualx -d ritualx -c "\dt"
```

Expected: `routine_logs` NOT in the list; `routines` still present

- [ ] **Step 3: Roll back migration 000003**

```powershell
docker run --rm --network host `
  -v "${PWD}/backend/migrations:/migrations" `
  migrate/migrate `
  -path=/migrations `
  -database "postgres://ritualx:ritualx_dev@localhost:5432/ritualx?sslmode=disable" `
  down 1
```

Expected: `000003/d create_routines`

- [ ] **Step 4: Verify routines dropped and ENUM gone**

```powershell
docker exec ritualx-postgres psql -U ritualx -d ritualx -c "\dt"
docker exec ritualx-postgres psql -U ritualx -d ritualx -c "SELECT enum_range(NULL::period_type);"
```

Expected: `routines` NOT in list; ENUM query returns error (type does not exist)

- [ ] **Step 5: Re-apply migrations**

```powershell
docker run --rm --network host `
  -v "${PWD}/backend/migrations:/migrations" `
  migrate/migrate `
  -path=/migrations `
  -database "postgres://ritualx:ritualx_dev@localhost:5432/ritualx?sslmode=disable" `
  up
```

Expected: both 000003 and 000004 applied cleanly

---

### Task 8: Update CHECKPOINT.md

**Files:**
- Modify: `.gemini/CHECKPOINT.md`

- [ ] **Step 1: Mark S2-01 as done**

Change `⬜ Pending` → `✅ Done` for S2-01 row.
