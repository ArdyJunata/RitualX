# S3-01 Streaks & Daily Goals Migration Implementation Plan

> **For agentic workers:** Steps use checkbox (`- [ ]`) syntax for tracking. Execute task-by-task, top to bottom. Every command is copy-paste ready for **Windows PowerShell**.

**Goal:** Create migrations `000005_create_streaks` and `000006_create_daily_goals`, plus GORM models `Streak` and `DailyGoal` in `backend/internal/model/`.

**Architecture:** Two sequential SQL migration pairs (up/down). Two GORM model files following existing conventions (`routine.go`, `routine_log.go`). No application logic — pure data layer.

**Tech Stack:** PostgreSQL 16, golang-migrate, GORM v2, `github.com/google/uuid`.

## Global Constraints

- Migration files: `backend/migrations/NNNNNN_<name>.up.sql` and `.down.sql`
- GORM models: `backend/internal/model/<name>.go`, `package model`
- UUID PK: `uuid_generate_v4()` (uuid-ossp already enabled by migration 000001)
- Timestamps: `TIMESTAMP WITH TIME ZONE` in SQL; `time.Time` in Go
- Nullable dates: `DATE` (no NOT NULL) in SQL; `*time.Time` in Go
- No commits until explicitly instructed
- Branch: `feat/s3-01-streaks-daily-goals`

> **Note (pre-existing, out of scope):** `go vet ./...` / `go test ./...` currently fail to compile `internal/service` test package because `mockRoutineLogRepo` in `routine_log_test.go` is missing `FindTodayByRoutineAndUser` (introduced by S2-06/PR #81). This is unrelated to S3-01. Verify S3-01 with `go build ./...` and `go vet ./internal/model/...` (both must pass). Do not fix the mock as part of this task unless instructed.

---

### Task 1: Create feature branch

**Files:** none

- [ ] **Step 1: Create branch**

```powershell
git checkout -b feat/s3-01-streaks-daily-goals
```

Expected: `Switched to a new branch 'feat/s3-01-streaks-daily-goals'`

---

### Task 2: Migration — create streaks table

**Files:**
- Create: `backend/migrations/000005_create_streaks.up.sql`
- Create: `backend/migrations/000005_create_streaks.down.sql`

- [ ] **Step 1: Create `000005_create_streaks.up.sql`**

```sql
CREATE TABLE streaks (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    routine_id     UUID NOT NULL REFERENCES routines(id) ON DELETE CASCADE,
    current_streak INTEGER NOT NULL DEFAULT 0,
    longest_streak INTEGER NOT NULL DEFAULT 0,
    last_completed DATE,
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_streaks_routine_id ON streaks(routine_id);
CREATE INDEX idx_streaks_user_id ON streaks(user_id);
```

- [ ] **Step 2: Create `000005_create_streaks.down.sql`**

```sql
DROP TABLE IF EXISTS streaks;
```

- [ ] **Step 3: Verify files exist**

```powershell
Get-ChildItem backend/migrations/ | Select-Object Name
```

Expected: includes `000005_create_streaks.up.sql` and `000005_create_streaks.down.sql`

---

### Task 3: Migration — create daily_goals table

**Files:**
- Create: `backend/migrations/000006_create_daily_goals.up.sql`
- Create: `backend/migrations/000006_create_daily_goals.down.sql`

- [ ] **Step 1: Create `000006_create_daily_goals.up.sql`**

```sql
CREATE TABLE daily_goals (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date           DATE NOT NULL,
    total_routines INTEGER NOT NULL DEFAULT 0,
    completed      INTEGER NOT NULL DEFAULT 0,
    is_achieved    BOOLEAN NOT NULL DEFAULT false,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_daily_goals_user_date ON daily_goals(user_id, date);
```

- [ ] **Step 2: Create `000006_create_daily_goals.down.sql`**

```sql
DROP TABLE IF EXISTS daily_goals;
```

---

### Task 4: GORM model — Streak

**Files:**
- Create: `backend/internal/model/streak.go`

- [ ] **Step 1: Create `backend/internal/model/streak.go`**

```go
package model

import (
	"time"

	"github.com/google/uuid"
)

// Streak tracks the current and longest completion streak for a single routine.
// One row per routine (enforced by the unique index on routine_id in migration
// 000005). LastCompleted is nullable — a freshly created streak has no last
// completion date yet.
type Streak struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null"                               json:"user_id"`
	RoutineID     uuid.UUID  `gorm:"type:uuid;not null"                               json:"routine_id"`
	CurrentStreak int        `gorm:"not null;default:0"                               json:"current_streak"`
	LongestStreak int        `gorm:"not null;default:0"                               json:"longest_streak"`
	LastCompleted *time.Time `gorm:"type:date"                                        json:"last_completed"`
	UpdatedAt     time.Time  `                                                        json:"updated_at"`
}
```

- [ ] **Step 2: Verify Go compiles**

```powershell
cd backend
go build ./...
cd ..
```

Expected: no errors, no output.

---

### Task 5: GORM model — DailyGoal

**Files:**
- Create: `backend/internal/model/daily_goal.go`

- [ ] **Step 1: Create `backend/internal/model/daily_goal.go`**

```go
package model

import (
	"time"

	"github.com/google/uuid"
)

// DailyGoal is a per-user, per-day rollup of routine completion progress.
// One row per (user_id, date) (enforced by the unique index in migration
// 000006). IsAchieved is set when Completed >= TotalRoutines for that day.
type DailyGoal struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null"                               json:"user_id"`
	Date          time.Time `gorm:"type:date;not null"                               json:"date"`
	TotalRoutines int       `gorm:"not null;default:0"                               json:"total_routines"`
	Completed     int       `gorm:"not null;default:0"                               json:"completed"`
	IsAchieved    bool      `gorm:"not null;default:false"                           json:"is_achieved"`
	CreatedAt     time.Time `                                                        json:"created_at"`
}
```

- [ ] **Step 2: Verify Go compiles + gofmt + model vet**

```powershell
cd backend
go build ./...
gofmt -l internal/model/streak.go internal/model/daily_goal.go
go vet ./internal/model/...
cd ..
```

Expected: `go build` no output; `gofmt -l` prints nothing (files already formatted); `go vet ./internal/model/...` no output.

---

### Task 6: Run migrations against dev DB

- [ ] **Step 1: Start dev DB**

```powershell
docker compose -f backend/docker-compose.dev.yml up -d
```

Expected: `ritualx-postgres` container running.

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

Expected (only the new ones apply if 000001–000004 already ran):
```
000005/u create_streaks (Xms)
000006/u create_daily_goals (Xms)
```

- [ ] **Step 4: Verify streaks table**

```powershell
docker exec ritualx-postgres psql -U ritualx -d ritualx -c "\d streaks"
```

Expected: columns `id, user_id, routine_id, current_streak, longest_streak, last_completed, updated_at`.

- [ ] **Step 5: Verify daily_goals table**

```powershell
docker exec ritualx-postgres psql -U ritualx -d ritualx -c "\d daily_goals"
```

Expected: columns `id, user_id, date, total_routines, completed, is_achieved, created_at`.

- [ ] **Step 6: Verify unique indexes**

```powershell
docker exec ritualx-postgres psql -U ritualx -d ritualx -c "\d+ streaks"
docker exec ritualx-postgres psql -U ritualx -d ritualx -c "\d+ daily_goals"
```

Expected: `idx_streaks_routine_id` UNIQUE on `(routine_id)`; `idx_daily_goals_user_date` UNIQUE on `(user_id, date)`.

---

### Task 7: Test rollback

- [ ] **Step 1: Roll back migration 000006**

```powershell
docker run --rm --network host `
  -v "${PWD}/backend/migrations:/migrations" `
  migrate/migrate `
  -path=/migrations `
  -database "postgres://ritualx:ritualx_dev@localhost:5432/ritualx?sslmode=disable" `
  down 1
```

Expected: `000006/d create_daily_goals`

- [ ] **Step 2: Verify daily_goals dropped**

```powershell
docker exec ritualx-postgres psql -U ritualx -d ritualx -c "\dt"
```

Expected: `daily_goals` NOT listed; `streaks` still present.

- [ ] **Step 3: Roll back migration 000005**

```powershell
docker run --rm --network host `
  -v "${PWD}/backend/migrations:/migrations" `
  migrate/migrate `
  -path=/migrations `
  -database "postgres://ritualx:ritualx_dev@localhost:5432/ritualx?sslmode=disable" `
  down 1
```

Expected: `000005/d create_streaks`

- [ ] **Step 4: Verify streaks dropped**

```powershell
docker exec ritualx-postgres psql -U ritualx -d ritualx -c "\dt"
```

Expected: `streaks` NOT listed.

- [ ] **Step 5: Re-apply migrations**

```powershell
docker run --rm --network host `
  -v "${PWD}/backend/migrations:/migrations" `
  migrate/migrate `
  -path=/migrations `
  -database "postgres://ritualx:ritualx_dev@localhost:5432/ritualx?sslmode=disable" `
  up
```

Expected: both 000005 and 000006 applied cleanly.

---

### Task 8: Update CHECKPOINT.md

**Files:**
- Modify: `.gemini/CHECKPOINT.md`

- [ ] **Step 1: Mark S3-01 as done**

In the Sprint 3 table, mark S3-01 `✅ Done`. Update "Next Session — Resume Here" to point to **S3-02** (Streak engine + API).

---

### Task 9: Commit (ONLY when instructed)

- [ ] **Step 1: Stage + commit**

```powershell
git add backend/migrations/000005_create_streaks.up.sql `
        backend/migrations/000005_create_streaks.down.sql `
        backend/migrations/000006_create_daily_goals.up.sql `
        backend/migrations/000006_create_daily_goals.down.sql `
        backend/internal/model/streak.go `
        backend/internal/model/daily_goal.go `
        docs/superpowers/specs/2026-07-31-s3-01-streaks-daily-goals-migration.md `
        docs/superpowers/plans/2026-07-31-s3-01-streaks-daily-goals.md
git commit -m "feat(backend): add streaks + daily_goals migrations and GORM models (S3-01)"
```
