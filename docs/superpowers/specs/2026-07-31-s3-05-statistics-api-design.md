# S3-05 — Statistics API Design

> **Date:** 2026-07-31 · **Issue:** #18 · **Points:** 5

## Overview

Aggregate, read-only statistics under `/api/v1/stats`. A new `StatisticsService` (separate from S3-04's per-routine `StatsService` for heatmap/progress, to avoid changing merged code). All user-scoped.

## Acceptance Criteria (issue #18 + PRD)

- `GET /stats/overview` — total completions, active streaks, level, XP ✔
- `GET /stats/routines/:id` — per-routine stats (completion rate, best streak) ✔
- `GET /stats/weekly-summary` — this week vs last week ✔
- `GET /stats/monthly-summary` — this month vs last month ✔

## Non-Goals

- Statistics screen UI (S3-09) · richer trend analytics/charts
- XP/level *mutation* (Sprint 4) — overview only *reads* `users.xp`/`users.level`

## Metric Definitions (all concrete, computable from current data)

**Overview** (`OverviewStats`):
- `total_completions` = `SUM(routine_logs.count)` for the user (all-time)
- `active_routines` = count of active routines
- `active_streaks` = count of the user's streaks with `current_streak > 0`
- `longest_streak` = max `longest_streak` across the user's streaks (0 if none)
- `level`, `xp` = from `users`

**Per-routine** (`RoutineStats`, ownership-checked):
- `total_completions` = `SUM(count)` for the routine
- `days_logged` = number of logged days (row count; unique per day)
- `current_streak`, `longest_streak`, `last_completed` = from `streaks` (0/0/null if none)
- `completion_rate` = `days_logged / active_days` (capped at 1, rounded 4dp), where `active_days` = whole days from `routine.created_at` to today, inclusive

**Weekly / Monthly summary** (`SummaryResponse`):
- `this` / `last` = `{ from, to, completions }` where `completions` = `SUM(count)` across all the user's logs in that period
- `delta` = `this.completions - last.completions`
- Weekly period = ISO week (Mon–Sun) via `CurrentPeriodRange("weekly", now)`; last week = previous 7-day block. Monthly = calendar month; last month = previous month.

## Data Access (new repo methods)

`RoutineLogRepository`:
- `SumCountByUser(userID) (int, error)`
- `SumCountByUserAndDateRange(userID, from, to) (int, error)`
- `SumCountByRoutine(routineID) (int, error)`
- `CountByRoutine(routineID) (int, error)`

`RoutineRepository`: `CountActiveByUserID(userID) (int, error)`
`UserRepository`: `FindByID(id) (*User, error)` (new; nil,nil when not found)
`StreakRepository`: reuse `FindByUserID`, `FindByRoutineID`.

## Service — `StatisticsService`

Interfaces `StatisticsLogRepoIface`, `StatisticsStreakRepoIface`, `StatisticsRoutineRepoIface`, `StatisticsUserRepoIface`; dual constructors. Errors logged via `logger.Get()` (rule #3). Reuses `CurrentPeriodRange` + `startOfDayUTC` (same package). Methods: `GetOverview`, `GetRoutineStats`, `GetWeeklySummary`, `GetMonthlySummary`.

## Endpoints

| Method | Path | Handler |
|--------|------|---------|
| GET | `/api/v1/stats/overview` | `GetStatsOverview` |
| GET | `/api/v1/stats/routines/:id` | `GetRoutineStatistics` (404 if not owned) |
| GET | `/api/v1/stats/weekly-summary` | `GetWeeklySummary` |
| GET | `/api/v1/stats/monthly-summary` | `GetMonthlySummary` |

## File Map

| File | Action |
|------|--------|
| `backend/internal/repository/routine_log.go` | +4 aggregate methods |
| `backend/internal/repository/routine.go` | +CountActiveByUserID |
| `backend/internal/repository/user.go` | +FindByID (uuid import) |
| `backend/internal/service/statistics.go` | New — StatisticsService + response structs |
| `backend/internal/service/statistics_test.go` | New — unit tests |
| `backend/internal/handler/statistics.go` | New — 4 handlers |
| `backend/cmd/server/main.go` | Wire StatisticsService + `/stats` group |

## Dependencies

- streaks (S3-02), routine_logs (S2), users (S1) ✅ · reuses `CurrentPeriodRange` (S3-04)

## Open Questions

- Per-routine "trend": represented via `current_streak` vs `longest_streak` + `completion_rate` (no separate time-series). Reviewer may expand later.
- `completion_rate` is day-based even for weekly/monthly routines (documented). 
