# S3-04 Heatmap & Progress API Implementation Plan

> Checkbox (`- [ ]`) steps. Windows PowerShell. Log service-layer errors via `logger.Get()` (rule #3).

**Goal:** `GET /routines/:id/heatmap?year=` (date→count map) + `GET /routines/:id/progress` (current-period completed/target/period). Closes #17.

## Global Constraints

- handler → service → repository; thin handlers · dual service constructors
- No commits until instructed · Branch: `feat/s3-04-heatmap-progress`

---

### Task 1: Repository range queries

**File:** Modify `backend/internal/repository/routine_log.go` (append; imports already have `time`, `model`, `uuid`)

```go
// ListByRoutineAndDateRange returns logs for a routine within [from, to], date ASC.
func (r *RoutineLogRepository) ListByRoutineAndDateRange(routineID uuid.UUID, from, to time.Time) ([]model.RoutineLog, error) {
	var logs []model.RoutineLog
	err := r.db.
		Where("routine_id = ? AND logged_at >= ? AND logged_at <= ?", routineID, from, to).
		Order("logged_at ASC").
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// SumCountByRoutineAndDateRange returns COALESCE(SUM(count),0) within [from, to].
func (r *RoutineLogRepository) SumCountByRoutineAndDateRange(routineID uuid.UUID, from, to time.Time) (int, error) {
	var result struct{ Total int64 }
	err := r.db.Model(&model.RoutineLog{}).
		Select("COALESCE(SUM(count), 0) AS total").
		Where("routine_id = ? AND logged_at >= ? AND logged_at <= ?", routineID, from, to).
		Scan(&result).Error
	return int(result.Total), err
}
```

---

### Task 2: StatsService

**File:** Create `backend/internal/service/stats.go`

- Interfaces `StatsLogRepoIface` (the 2 range methods) + `StatsRoutineRepoIface` (`FindByIDAndUserID`).
- `ProgressResponse{ Completed, Target int; Period string }` (json: completed/target/period).
- `GetHeatmap(userID, routineID, year)` — ownership check (nil→NOT_FOUND); `ListByRoutineAndDateRange(Jan1..Dec31)`; build `map[string]int` keyed `LoggedAt.Format("2006-01-02")`.
- `GetProgress(userID, routineID)` — ownership; `from,to := CurrentPeriodRange(routine.PeriodType, time.Now().UTC())`; `SumCountByRoutineAndDateRange`; return `{completed, routine.TargetCount, routine.PeriodType}`.
- Exported `CurrentPeriodRange(periodType string, now time.Time) (time.Time, time.Time)`: daily→(day,day); weekly→(Monday, Monday+6) via `startOfWeek`; monthly→(1st, last). Uses existing `startOfDayUTC` + `startOfWeek`.
- All error branches log + `INTERNAL_ERROR`.

---

### Task 3: Handlers

**File:** Create `backend/internal/handler/stats.go`

- `GetRoutineHeatmap` — `?year` optional (`strconv.Atoi`, validate 2000–2100, default `time.Now().UTC().Year()`); 400 on invalid; `success(200, map)`.
- `GetRoutineProgress` — `success(200, ProgressResponse)`.

---

### Task 4: Wire routes

**File:** Modify `backend/cmd/server/main.go` (after daily-goal wiring)

```go
statsService := service.NewStatsService(routineLogRepo, routineRepo)
routines.Get("/:id/heatmap", handler.GetRoutineHeatmap(statsService))
routines.Get("/:id/progress", handler.GetRoutineProgress(statsService))
```

---

### Task 5: Unit tests

**File:** Create `backend/internal/service/stats_test.go` (`package service_test`)

- `CurrentPeriodRange` daily / weekly (Wed 2026-07-15 → Mon 07-13..Sun 07-19) / monthly (07-01..07-31).
- `GetHeatmap` builds map from mocked logs; routine-not-found → NOT_FOUND.
- `GetProgress` weekly (target 5, sum 4 → {4,5,"weekly"}); routine-not-found → NOT_FOUND.
- Mocks: `mockStatsLogRepo`, `mockStatsRoutineRepo`.

---

### Task 6: Verify

- [ ] `go build ./...`, `go vet ./...`, `gofmt -l` clean
- [ ] `go test ./internal/service/ -run "Stats|Heatmap|Progress|CurrentPeriodRange"` pass
- [ ] Validate range/sum SQL on dev DB

---

### Task 7: Commit / Push / PR

Branch `feat/s3-04-heatmap-progress`; commit (identity via `-c`), push, PR via REST API, body ends `Closes #17`.
