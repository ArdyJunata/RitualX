# S3-01 — Streaks & Daily Goals DB Migration Design

> **Date:** 2026-07-31

## Overview

Create database migrations and GORM models for the `streaks` and `daily_goals` tables. This is the foundational data layer for Sprint 3 (Heatmap, Streaks & Statistics). The streak engine (S3-02) and daily-goal tracking API (S3-03) depend on it.

## Goals

- `streaks` table: per-routine current/longest streak tracking
- `daily_goals` table: per-user, per-day completion rollup
- GORM models for both tables, matching existing conventions (`Routine`, `RoutineLog`)
- Rollback (down) migrations
- No application logic — pure data layer

## Non-Goals

- Streak calculation engine + API (S3-02)
- Daily-goal tracking API (S3-03)
- Heatmap / statistics endpoints (S3-04, S3-05)
- Repositories or services (added in S3-02 / S3-03)
- Seeding data

## Approach

Follow the existing migration pattern (`000001`–`000004`): sequential numbered SQL files, `up` + `down` pairs, run via `golang-migrate`. GORM models follow the `Routine` / `RoutineLog` conventions: `uuid.UUID` PK with `uuid_generate_v4()`, explicit GORM tags, JSON tags, GORM-managed timestamps. Table names rely on GORM default pluralization (`Streak` → `streaks`, `DailyGoal` → `daily_goals`) — no `TableName()` override, consistent with existing models.

## Design Details

### Migration: `000005_create_streaks.up.sql`

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

### Migration: `000005_create_streaks.down.sql`

```sql
DROP TABLE IF EXISTS streaks;
```

### Migration: `000006_create_daily_goals.up.sql`

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

### Migration: `000006_create_daily_goals.down.sql`

```sql
DROP TABLE IF EXISTS daily_goals;
```

### GORM Model: `backend/internal/model/streak.go`

```go
type Streak struct {
    ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    UserID        uuid.UUID  `gorm:"type:uuid;not null"                               json:"user_id"`
    RoutineID     uuid.UUID  `gorm:"type:uuid;not null"                               json:"routine_id"`
    CurrentStreak int        `gorm:"not null;default:0"                               json:"current_streak"`
    LongestStreak int        `gorm:"not null;default:0"                               json:"longest_streak"`
    LastCompleted *time.Time `gorm:"type:date"                                        json:"last_completed"`
    UpdatedAt     time.Time                                                            `json:"updated_at"`
}
```

### GORM Model: `backend/internal/model/daily_goal.go`

```go
type DailyGoal struct {
    ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    UserID        uuid.UUID `gorm:"type:uuid;not null"                               json:"user_id"`
    Date          time.Time `gorm:"type:date;not null"                               json:"date"`
    TotalRoutines int       `gorm:"not null;default:0"                               json:"total_routines"`
    Completed     int       `gorm:"not null;default:0"                               json:"completed"`
    IsAchieved    bool      `gorm:"not null;default:false"                           json:"is_achieved"`
    CreatedAt     time.Time                                                           `json:"created_at"`
}
```

## Constraints & Design Decisions

- **Schema source:** PRD "Streak & Goals" section (`docs/superpowers/specs/2026-06-19-ritualx-design.md`).
- **`UNIQUE(routine_id)` on streaks** — *not* in the PRD; added deliberately. One routine has exactly one streak row; this enables upsert by the streak engine (S3-02) and prevents duplicates. `routine_id` already implies `user_id` (via `routines.user_id`), so this is equivalent to `UNIQUE(user_id, routine_id)`. Reviewer may veto.
- **`UNIQUE(user_id, date)` on daily_goals** — *not* explicit in the PRD; added deliberately. A daily goal is inherently one-per-user-per-day; mirrors `routine_logs`' `UNIQUE(routine_id, logged_at)`. Enables upsert by the daily-goal API (S3-03). The composite index's leftmost prefix also serves `WHERE user_id = ?` queries, so no separate `user_id` index is added.
- **`NOT NULL DEFAULT` on integer/boolean columns** — follows codebase convention (`users.xp`, `routines.target_count`, `routine_logs.count`). PRD left `total_routines`' default blank; set to `DEFAULT 0` for consistency.
- **`last_completed DATE` nullable** — a fresh streak has no completion yet; modeled as `*time.Time` in Go.
- **streaks has `updated_at` only; daily_goals has `created_at` only** — matches the PRD column lists exactly.
- **`ON DELETE CASCADE`** on all FKs — deleting a user or routine removes dependent rows, consistent with existing tables.

## Dependencies

- Migration `000001` (users, uuid-ossp) ✅
- Migration `000003` (routines) ✅ — `streaks.routine_id` FK
- No other dependencies

## Open Questions

- Should `streaks` also carry `created_at`? PRD omits it; deferred to reviewer.
