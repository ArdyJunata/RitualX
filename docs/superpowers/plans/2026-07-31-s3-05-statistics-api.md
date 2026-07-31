# S3-05 Statistics API Implementation Plan

> Checkbox (`- [ ]`) steps. Windows PowerShell. Log service-layer errors via `logger.Get()` (rule #3). Metric definitions: see the design spec.

**Goal:** `/stats/overview`, `/stats/routines/:id`, `/stats/weekly-summary`, `/stats/monthly-summary`. New `StatisticsService`. Closes #18.

## Global Constraints
- handler → service → repository; thin handlers · dual service constructors
- Reuse `CurrentPeriodRange` + `startOfDayUTC` (service package)
- No commits until instructed · Branch: `feat/s3-05-statistics-api`

---

### Task 1: Repository aggregates

**`backend/internal/repository/routine_log.go`** — append:
```go
func (r *RoutineLogRepository) SumCountByUser(userID uuid.UUID) (int, error) {
	var result struct{ Total int64 }
	err := r.db.Model(&model.RoutineLog{}).Select("COALESCE(SUM(count), 0) AS total").
		Where("user_id = ?", userID).Scan(&result).Error
	return int(result.Total), err
}
func (r *RoutineLogRepository) SumCountByUserAndDateRange(userID uuid.UUID, from, to time.Time) (int, error) {
	var result struct{ Total int64 }
	err := r.db.Model(&model.RoutineLog{}).Select("COALESCE(SUM(count), 0) AS total").
		Where("user_id = ? AND logged_at >= ? AND logged_at <= ?", userID, from, to).Scan(&result).Error
	return int(result.Total), err
}
func (r *RoutineLogRepository) SumCountByRoutine(routineID uuid.UUID) (int, error) {
	var result struct{ Total int64 }
	err := r.db.Model(&model.RoutineLog{}).Select("COALESCE(SUM(count), 0) AS total").
		Where("routine_id = ?", routineID).Scan(&result).Error
	return int(result.Total), err
}
func (r *RoutineLogRepository) CountByRoutine(routineID uuid.UUID) (int, error) {
	var count int64
	err := r.db.Model(&model.RoutineLog{}).Where("routine_id = ?", routineID).Count(&count).Error
	return int(count), err
}
```

**`backend/internal/repository/routine.go`** — append:
```go
func (r *RoutineRepository) CountActiveByUserID(userID uuid.UUID) (int, error) {
	var count int64
	err := r.db.Model(&model.Routine{}).Where("user_id = ? AND is_active = true", userID).Count(&count).Error
	return int(count), err
}
```

**`backend/internal/repository/user.go`** — add `uuid` import + method:
```go
func (r *UserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
```

---

### Task 2: StatisticsService

**File:** Create `backend/internal/service/statistics.go`

- Interfaces: `StatisticsLogRepoIface` (4 methods), `StatisticsStreakRepoIface` (`FindByUserID`, `FindByRoutineID`), `StatisticsRoutineRepoIface` (`FindByIDAndUserID`, `CountActiveByUserID`), `StatisticsUserRepoIface` (`FindByID`).
- Structs: `OverviewStats`, `RoutineStats`, `PeriodSummary{from,to,completions}`, `SummaryResponse{this,last,delta}`.
- Dual constructors `NewStatisticsService` / `NewStatisticsServiceIface`.
- `GetOverview`: user (FindByID; nil→NOT_FOUND) → total (SumCountByUser), active (CountActiveByUserID), streaks (FindByUserID → count current>0, max longest), level/xp.
- `GetRoutineStats`: ownership (nil→NOT_FOUND) → total/days/streak; `completion_rate = days / activeDays` (activeDays from `routine.CreatedAt`→today inclusive, min 1; cap 1; round 4dp via `math`).
- `GetWeeklySummary`/`GetMonthlySummary`: `this` = `CurrentPeriodRange(weekly|monthly, now)`; `last` derived (week: −7d block; month: previous month). **Call this-period sum before last-period sum** (test relies on order). `delta = this − last`. Dates formatted `2006-01-02`.
- All error branches log + `INTERNAL_ERROR`.

---

### Task 3: Handlers + routes

**File:** Create `backend/internal/handler/statistics.go` — `GetStatsOverview`, `GetRoutineStatistics` (parse `:id`), `GetWeeklySummary`, `GetMonthlySummary` (thin, 200).

**File:** `backend/cmd/server/main.go` — after stats(S3-04) wiring:
```go
statisticsService := service.NewStatisticsService(routineLogRepo, streakRepo, routineRepo, userRepo)
stats := api.Group("/stats", middleware.RequireAuth(cfg.JWTSecret))
stats.Get("/overview", handler.GetStatsOverview(statisticsService))
stats.Get("/routines/:id", handler.GetRoutineStatistics(statisticsService))
stats.Get("/weekly-summary", handler.GetWeeklySummary(statisticsService))
stats.Get("/monthly-summary", handler.GetMonthlySummary(statisticsService))
```
(`userRepo`, `streakRepo`, `routineLogRepo`, `routineRepo` are already in scope.)

---

### Task 4: Unit tests

**File:** Create `backend/internal/service/statistics_test.go` (`package service_test`, mocks for the 4 interfaces)
- Overview: total/active/streaks(current>0 count)/longest/level/xp; user-not-found → NOT_FOUND.
- RoutineStats: ownership + fields + completion_rate; not-found → NOT_FOUND; nil streak → zeros.
- Weekly & monthly: mock sum `.Return(this).Once()` then `.Return(last).Once()`; assert `delta`.

---

### Task 5: Verify
- [ ] `go build ./...`, `go vet ./...`, `gofmt -l` clean
- [ ] `go test ./internal/service/ -run "Statistics|Overview|RoutineStats|Summary"` pass
- [ ] Validate aggregate SQL on dev DB

---

### Task 6: Commit / Push / PR
Branch `feat/s3-05-statistics-api`; commit (identity `-c`), push, PR REST API, body ends `Closes #18`.
