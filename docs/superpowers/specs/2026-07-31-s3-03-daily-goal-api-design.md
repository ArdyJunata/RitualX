# S3-03 — Daily Goal Tracking API Design

> **Date:** 2026-07-31 · **Issue:** #16 · **Points:** 3

## Overview

Per-user, per-day completion rollup. Compute how many of a user's active routines were completed on a given day and whether the daily goal was achieved. Persist to the `daily_goals` table (S3-01) on each log/undo event; expose a today-summary endpoint and a history endpoint.

## Acceptance Criteria (issue #16 + PRD)

- Calculate total active routines and completed count per day ✔
- `GET /api/v1/goals/daily` — today's summary (total, completed, achieved) ✔
- `GET /api/v1/goals/daily/history?from=&to=` — history ✔

## Non-Goals

- Heatmap endpoint (S3-04) · Statistics (S3-05) · progress-ring UI (S3-10)
- Per-period-type goal weighting (all active routines count; documented simplification)

## Semantics

For a user + date:
- `total_routines` = count of the user's **active** routines (`is_active = true`).
- `completed` = count of **distinct active routines** that have a log on that date (`routine_logs.logged_at = date`, joined to active routines).
- `is_achieved` = `total_routines > 0 && completed >= total_routines`.
- `date` normalized to 00:00 UTC.

## Approach

- **Today summary** (`GET /goals/daily`) — computed **live** (no write) so it is always accurate even if routines were added/removed since the last log event.
- **Event persistence** — on each log/undo event, `Recalculate(userID, date)` computes and **upserts** the row so `daily_goals` accumulates historical snapshots (used by history + future heatmap/stats).
- **History** (`GET /goals/daily/history`) — returns persisted rows in `[from, to]` (defaults to the last 30 days). Past-date totals come from the snapshot captured at event time.

> Note: `total_routines` for a *past* date reflects the snapshot persisted when logs occurred that day. `GetToday` always reflects the current active-routine count.

## Data Access — `DailyGoalRepository`

- `Upsert(g)` — raw SQL `ON CONFLICT (user_id, date)`.
- `CountActiveRoutines(userID) (int, error)` — `routines WHERE user_id=? AND is_active=true`.
- `CountCompletedRoutines(userID, date) (int, error)` — `routine_logs rl JOIN routines r ON r.id=rl.routine_id WHERE rl.user_id=? AND rl.logged_at=? AND r.is_active=true`.
- `FindByUserAndDateRange(userID, from, to) ([]DailyGoal, error)` — ordered by `date ASC`.

## Service — `DailyGoalService`

Interface `DailyGoalRepoIface`; dual constructors. All error paths log via `logger.Get()` (rule #3).
- `Recalculate(userID, date) (*DailyGoal, *ServiceError)` — compute + upsert (event-driven).
- `GetToday(userID) (*DailyGoal, *ServiceError)` — compute live for today (no write).
- `GetHistory(userID, from, to) ([]DailyGoal, *ServiceError)` — persisted rows.

## Integration (log flow)

`LogRoutine` recalculates for the logged entry's date (`entry.LoggedAt`); `DeleteRoutineLog` recalculates for **today** (UI undo targets today). Both best-effort (errors logged, do not fail the log/undo response). Handlers gain a `*service.DailyGoalService` param.

## Endpoints

| Method | Path | Handler | Response |
|--------|------|---------|----------|
| GET | `/api/v1/goals/daily` | `GetDailyGoal` | 200 `{ data: DailyGoal }` |
| GET | `/api/v1/goals/daily/history?from=&to=` | `GetDailyGoalHistory` | 200 `{ data: [DailyGoal] }` (default last 30d) |

`DailyGoal` JSON: `{ id, user_id, date, total_routines, completed, is_achieved, created_at }`.

## File Map

| File | Action |
|------|--------|
| `backend/internal/repository/daily_goal.go` | New — DailyGoalRepository |
| `backend/internal/service/daily_goal.go` | New — DailyGoalService |
| `backend/internal/service/daily_goal_test.go` | New — unit tests |
| `backend/internal/handler/daily_goal.go` | New — GetDailyGoal, GetDailyGoalHistory |
| `backend/internal/handler/routine.go` | Modify LogRoutine, DeleteRoutineLog (add dailyGoalService) |
| `backend/cmd/server/main.go` | Wire DailyGoalRepository/Service + `/goals` routes |

## Dependencies

- S3-01 `daily_goals` table + `DailyGoal` model ✅ · `UNIQUE(user_id, date)` backs the upsert ✅
- S2 routines + routine_logs ✅

## Open Questions

- Should only `daily` period-type routines count toward the daily goal? Currently all active routines count (documented). Reviewer may refine.
