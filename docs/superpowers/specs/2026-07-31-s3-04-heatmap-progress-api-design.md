# S3-04 — Heatmap & Progress API Design

> **Date:** 2026-07-31 · **Issue:** #17 · **Points:** 3

## Overview

Two read-only, per-routine endpoints backing the contribution heatmap and the current-period progress ring:
- `GET /api/v1/routines/:id/heatmap?year=YYYY` → date→count map for the year.
- `GET /api/v1/routines/:id/progress` → current-period `{ completed, target, period }`.

Both are scoped to the authenticated user (routine ownership enforced).

## Acceptance Criteria (issue #17 + PRD line 862)

- `GET /routines/:id/heatmap?year=2026` → `{ "2026-01-15": 3, "2026-01-16": 1, ... }` ✔
- `GET /routines/:id/progress` → `{ completed: 4, target: 5, period: "weekly" }` ✔
- Data scoped to authenticated user only ✔

## Non-Goals

- Aggregate/overview statistics across all routines (S3-05)
- Heatmap UI component (S3-06) · routine detail screen (S3-08)

## Semantics

- **Heatmap:** for the given `year` (default = current UTC year), return a map keyed by `YYYY-MM-DD` → that day's `routine_logs.count`. Only days with a log appear (sparse map). `routine_logs` is unique per `(routine_id, logged_at)`, so one entry per date.
- **Progress:** `period` = `routine.period_type`; `target` = `routine.target_count`; `completed` = `SUM(routine_logs.count)` within the **current period**:
  - daily → today
  - weekly → current ISO week (Mon–Sun)
  - monthly → current calendar month

## Data Access — add to `RoutineLogRepository`

- `ListByRoutineAndDateRange(routineID, from, to) ([]RoutineLog, error)` — ordered `logged_at ASC` (heatmap).
- `SumCountByRoutineAndDateRange(routineID, from, to) (int, error)` — `COALESCE(SUM(count),0)` (progress).

## Service — `StatsService`

Interfaces `StatsLogRepoIface` (2 range methods) + `StatsRoutineRepoIface` (`FindByIDAndUserID`); dual constructors. Errors logged via `logger.Get()` (rule #3).
- `GetHeatmap(userID, routineID, year) (map[string]int, *ServiceError)` — ownership → list year range → build map.
- `GetProgress(userID, routineID) (*ProgressResponse, *ServiceError)` — ownership → `CurrentPeriodRange` → sum → `{completed, target, period}`.
- `CurrentPeriodRange(periodType, now) (from, to time.Time)` — **exported pure** helper (reuses `startOfWeek`, `startOfDayUTC`), unit-tested.

`ProgressResponse` JSON: `{ completed, target, period }`.

## Endpoints

| Method | Path | Handler | Response |
|--------|------|---------|----------|
| GET | `/api/v1/routines/:id/heatmap?year=YYYY` | `GetRoutineHeatmap` | 200 `{ data: {"YYYY-MM-DD": count} }` · 404 if not owned · 400 if year invalid |
| GET | `/api/v1/routines/:id/progress` | `GetRoutineProgress` | 200 `{ data: { completed, target, period } }` · 404 if not owned |

`year` defaults to the current UTC year; validated as 2000–2100.

## File Map

| File | Action |
|------|--------|
| `backend/internal/repository/routine_log.go` | Add 2 range query methods |
| `backend/internal/service/stats.go` | New — StatsService + CurrentPeriodRange |
| `backend/internal/service/stats_test.go` | New — unit tests |
| `backend/internal/handler/stats.go` | New — GetRoutineHeatmap, GetRoutineProgress |
| `backend/cmd/server/main.go` | Wire StatsService + 2 routes |

## Dependencies

- S2 routines + routine_logs ✅ · `Routine.PeriodType` / `TargetCount` ✅
- Reuses `startOfWeek` (streak.go) + `startOfDayUTC` (daily_goal.go) — same package

## Open Questions

- `completed` sums `count` (total times done in period). Alternative = distinct days. Chose sum to match weekly targets like "5×/week". Reviewer may refine.
