# S2-01 — Routines & Routine Logs DB Migration Design

> **Date:** 2026-07-13

## Overview

Create database migrations and GORM models for `routines` and `routine_logs` tables. This is the foundational data layer for Sprint 2. All routine CRUD and log APIs depend on it.

## Goals

- `routines` table: stores user-owned routine definitions
- `routine_logs` table: stores per-day completion records
- GORM models for both tables, matching existing conventions
- Rollback (down) migrations
- No application logic — pure data layer

## Non-Goals

- API endpoints (S2-02, S2-03, S2-04)
- Streak / daily_goals tables (S3-01)
- Seeding data

## Approach

Follow existing migration pattern (`000001`, `000002`): sequential numbered SQL files, `up` + `down` pairs. GORM models follow `User` / `RefreshToken` conventions: `uuid.UUID` PK, explicit GORM tags, JSON tags.

## Design Details

### Migration: `000003_create_routines.up.sql`

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

### Migration: `000003_create_routines.down.sql`

```sql
DROP TABLE IF EXISTS routines;
DROP TYPE IF EXISTS period_type;
```

### Migration: `000004_create_routine_logs.up.sql`

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

### Migration: `000004_create_routine_logs.down.sql`

```sql
DROP TABLE IF EXISTS routine_logs;
```

### GORM Model: `backend/internal/model/routine.go`

```go
type Routine struct {
    ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    UserID      uuid.UUID  `gorm:"type:uuid;not null"                               json:"user_id"`
    Title       string     `gorm:"type:varchar(100);not null"                       json:"title"`
    Description string     `gorm:"type:text"                                        json:"description"`
    PeriodType  string     `gorm:"type:period_type;not null"                        json:"period_type"`
    TargetCount int        `gorm:"not null;default:1"                               json:"target_count"`
    Icon        string     `gorm:"type:varchar(50)"                                 json:"icon"`
    Color       string     `gorm:"type:varchar(20)"                                 json:"color"`
    IsActive    bool       `gorm:"not null;default:true"                            json:"is_active"`
    SortOrder   int        `gorm:"not null;default:0"                               json:"sort_order"`
    CreatedAt   time.Time                                                           `json:"created_at"`
    UpdatedAt   time.Time                                                           `json:"updated_at"`
}
```

### GORM Model: `backend/internal/model/routine_log.go`

```go
type RoutineLog struct {
    ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    RoutineID  uuid.UUID  `gorm:"type:uuid;not null"                               json:"routine_id"`
    UserID     uuid.UUID  `gorm:"type:uuid;not null"                               json:"user_id"`
    LoggedAt   time.Time  `gorm:"type:date;not null"                               json:"logged_at"`
    Count      int        `gorm:"not null;default:1"                               json:"count"`
    Note       string     `gorm:"type:text"                                        json:"note"`
    CreatedAt  time.Time                                                           `json:"created_at"`
}
```

## Constraints

- `period_type` ENUM: `'daily'`, `'weekly'`, `'monthly'` — validated at DB level
- `UNIQUE (routine_id, logged_at)` — prevents duplicate logs per routine per day
- `ON DELETE CASCADE` — deleting user removes all their routines + logs
- `target_count` min 1 (enforced in service layer, not DB)

## Dependencies

- Migration `000001` (users table, uuid-ossp) ✅
- No other dependencies

## Open Questions

- None
