# S3-02 — Streak Calculation Engine + API Design

> **Date:** 2026-07-31 · **Issue:** #15 · **Points:** 5

## Overview

Server-side streak engine. On every log/undo event, recompute a routine's current + longest streak from its `routine_logs` and persist to the `streaks` table (S3-01). Expose read-only endpoints. Per PRD: "Streaks are calculated server-side on each log event — no manual CRUD."

## Acceptance Criteria (issue #15)

- Streak engine runs on each log event ✔ (hooked into log + undo)
- Daily / Weekly / Monthly streaks incremented ✔ (period-aware calc)
- `longest_streak` updated ✔ (high-water mark, never decreases)
- Endpoints for streaks ✔ (`GET /routines/:id/streak`, `GET /streaks`)

## Non-Goals

- Daily goal tracking (S3-03) · Heatmap (S3-04) · Statistics (S3-05)
- Target-count-aware completion (a period counts if it has ≥1 log; documented simplification)
- Streak milestone animations (frontend, later)
- Manual streak CRUD (none — engine only)

## Approach

**Recompute-from-source on each event.** Fetch all log dates for the routine, fold into completed *periods* (day / ISO-week / month), find consecutive runs. This is idempotent and self-healing (undo/backfill recompute correctly), simpler and less bug-prone than incremental counters.

## Streak Semantics

- A **period** is completed if it has ≥1 log (routine_logs.count is always ≥1). `target_count` is not factored in (documented simplification; future enhancement).
- Period key by `period_type`:
  - `daily` → the date (00:00 UTC)
  - `weekly` → Monday of that ISO week
  - `monthly` → first of that month
- **current_streak** = length of the consecutive run of completed periods ending at the *latest* completed period.
- **longest_streak** = longest consecutive run over all history, kept as a **high-water mark** (persisted value never decreases, even on undo).
- **last_completed** = latest log date (nullable; NULL when no logs).
- Adjacency: daily = +1 day, weekly = +7 days (week starts), monthly = next calendar month.

> **Freshness caveat:** values are computed at event time and persisted. A `GET` after a long idle gap returns the last-computed value (may be stale). Acceptable for S3-02; a read-time recompute or cron can refine later.

## Core Function (pure, unit-tested)

```go
// CalculateStreak folds log dates into completed periods and returns the
// current run (ending at the latest period), the longest run, and the latest
// log date. Input need not be sorted or deduplicated.
func CalculateStreak(periodType string, logDates []time.Time) (current int, longest int, lastCompleted *time.Time)
```

## Data Access

- **New** `StreakRepository`: `Upsert` (raw SQL `ON CONFLICT (routine_id)`), `FindByRoutineID` (nil,nil when none), `FindByUserID` (slice).
- **Add** to `RoutineLogRepository`: `ListDatesByRoutine(routineID) ([]time.Time, error)` (ordered `logged_at ASC`).

## Service (`StreakService`)

Interfaces `StreakRepoIface`, `StreakLogRepoIface` (`ListDatesByRoutine`), `StreakRoutineRepoIface` (`FindByIDAndUserID`). Dual constructors `NewStreakService` (concrete) + `NewStreakServiceIface` (tests). All error paths log via `logger.Get()` (rule #3).

- `Recalculate(userID, routineID) (*model.Streak, *ServiceError)` — verify ownership → list dates → `CalculateStreak` → high-water longest vs existing → `Upsert`.
- `GetByRoutine(userID, routineID) (*model.Streak, *ServiceError)` — verify ownership; return persisted streak or a zero-value streak (current=longest=0, last_completed=nil) if none.
- `ListByUser(userID) ([]model.Streak, *ServiceError)` — all streaks for the user.

## Integration (engine on each event)

`LogRoutine` and `DeleteRoutineLog` handlers gain a `*service.StreakService` param. After a successful log/delete they call `Recalculate(userID, routineID)`. Best-effort: a recalc error is logged (in-service) but does **not** fail the log/undo response (the log write already succeeded).

## Endpoints

| Method | Path | Handler | Response |
|--------|------|---------|----------|
| GET | `/api/v1/routines/:id/streak` | `GetRoutineStreak` | 200 `{ data: Streak }` · 404 if routine not owned |
| GET | `/api/v1/streaks` | `ListStreaks` | 200 `{ data: [Streak] }` |

`Streak` JSON: `{ id, user_id, routine_id, current_streak, longest_streak, last_completed, updated_at }`.

## File Map

| File | Action |
|------|--------|
| `backend/internal/repository/streak.go` | New — StreakRepository |
| `backend/internal/repository/routine_log.go` | Add `ListDatesByRoutine` |
| `backend/internal/service/streak.go` | New — StreakService + CalculateStreak |
| `backend/internal/service/streak_test.go` | New — calc + service unit tests |
| `backend/internal/handler/streak.go` | New — GetRoutineStreak, ListStreaks |
| `backend/internal/handler/routine.go` | Modify LogRoutine, DeleteRoutineLog signatures |
| `backend/cmd/server/main.go` | Wire StreakRepository/Service + routes |

## Dependencies

- S3-01 `streaks` table + `Streak` model ✅
- S2-04 routine_logs + log/undo endpoints ✅
- `Routine.PeriodType` (`daily`/`weekly`/`monthly`) ✅

## Open Questions

- Should `target_count` gate period completion? Deferred (simplification documented).
- High-water `longest_streak` vs recompute-can-decrease — chose high-water; reviewer may veto.
