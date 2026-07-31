# S3-03 Daily Goal Tracking API Implementation Plan

> Steps use checkbox (`- [ ]`) syntax. Windows PowerShell. Log all service-layer errors via `logger.Get()` (rule #3).

**Goal:** Daily goal rollup (total active routines, completed count, achieved) per user/day. Persist on log/undo events; expose today summary + history. Closes #16.

## Global Constraints

- handler → service → repository; thin handlers
- Dual service constructors `NewDailyGoalService` / `NewDailyGoalServiceIface`
- No commits until instructed · Branch: `feat/s3-03-daily-goal-api`

---

### Task 1: DailyGoalRepository

**File:** Create `backend/internal/repository/daily_goal.go`

```go
package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/google/uuid"
)

type DailyGoalRepository struct {
	db *gorm.DB
}

func NewDailyGoalRepository(db *gorm.DB) *DailyGoalRepository {
	return &DailyGoalRepository{db: db}
}

// Upsert inserts or updates the daily goal row for (user_id, date).
func (r *DailyGoalRepository) Upsert(g *model.DailyGoal) error {
	sql := `
		INSERT INTO daily_goals (id, user_id, date, total_routines, completed, is_achieved, created_at)
		VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, NOW())
		ON CONFLICT (user_id, date)
		DO UPDATE SET total_routines = EXCLUDED.total_routines,
		              completed = EXCLUDED.completed,
		              is_achieved = EXCLUDED.is_achieved
		RETURNING id, user_id, date, total_routines, completed, is_achieved, created_at`
	return r.db.Raw(sql, g.UserID, g.Date, g.TotalRoutines, g.Completed, g.IsAchieved).Scan(g).Error
}

// CountActiveRoutines returns the number of active routines for a user.
func (r *DailyGoalRepository) CountActiveRoutines(userID uuid.UUID) (int, error) {
	var count int64
	err := r.db.Model(&model.Routine{}).
		Where("user_id = ? AND is_active = true", userID).
		Count(&count).Error
	return int(count), err
}

// CountCompletedRoutines returns distinct active routines with a log on the date.
func (r *DailyGoalRepository) CountCompletedRoutines(userID uuid.UUID, date time.Time) (int, error) {
	var count int64
	err := r.db.Table("routine_logs AS rl").
		Joins("JOIN routines r ON r.id = rl.routine_id").
		Where("rl.user_id = ? AND rl.logged_at = ? AND r.is_active = true", userID, date).
		Count(&count).Error
	return int(count), err
}

// FindByUserAndDateRange returns persisted daily goals in [from, to], date ASC.
func (r *DailyGoalRepository) FindByUserAndDateRange(userID uuid.UUID, from, to time.Time) ([]model.DailyGoal, error) {
	var goals []model.DailyGoal
	err := r.db.
		Where("user_id = ? AND date >= ? AND date <= ?", userID, from, to).
		Order("date ASC").
		Find(&goals).Error
	if err != nil {
		return nil, err
	}
	return goals, nil
}
```

---

### Task 2: DailyGoalService

**File:** Create `backend/internal/service/daily_goal.go`

Interface `DailyGoalRepoIface` (the 4 repo methods). `DailyGoalService` with `Recalculate` (compute + upsert), `GetToday` (compute live), `GetHistory`. Private `compute(userID, date)` shared by Recalculate/GetToday. `is_achieved = total > 0 && completed >= total`. Private `startOfDayUTC(t)` normalizes dates. All error branches log via `logger.Get()` and return `INTERNAL_ERROR`.

---

### Task 3: Handlers

**File:** Create `backend/internal/handler/daily_goal.go` — `GetDailyGoal` (today) + `GetDailyGoalHistory` (`?from=&to=`, default last 30 days, validate `YYYY-MM-DD`).

**File:** Modify `backend/internal/handler/routine.go` — `LogRoutine` and `DeleteRoutineLog` add `dailyGoalService *service.DailyGoalService`. After success:
- LogRoutine: `_, _ = dailyGoalService.Recalculate(userID, entry.LoggedAt)`
- DeleteRoutineLog: `_, _ = dailyGoalService.Recalculate(userID, time.Now().UTC())`

Add `"time"` import to routine.go.

---

### Task 4: Wire routes

**File:** Modify `backend/cmd/server/main.go`

```go
dailyGoalRepo := repository.NewDailyGoalRepository(db)
dailyGoalService := service.NewDailyGoalService(dailyGoalRepo)

routines.Post("/:id/log", handler.LogRoutine(routineLogService, streakService, dailyGoalService))
routines.Delete("/:id/log/:logId", handler.DeleteRoutineLog(routineLogService, streakService, dailyGoalService))
// ... existing GET /:id/log, /:id/streak ...

// Daily goal routes (auth required)
goals := api.Group("/goals", middleware.RequireAuth(cfg.JWTSecret))
goals.Get("/daily", handler.GetDailyGoal(dailyGoalService))
goals.Get("/daily/history", handler.GetDailyGoalHistory(dailyGoalService))
```

---

### Task 5: Unit tests

**File:** Create `backend/internal/service/daily_goal_test.go` (`package service_test`, mock `DailyGoalRepoIface`)

- Recalculate achieved (total=3, completed=3 → true, Upsert called)
- Recalculate not achieved (3/1 → false)
- Recalculate no routines (0/0 → false)
- GetToday (2/1 → not achieved; no Upsert)
- GetHistory returns rows

---

### Task 6: Verify

- [ ] `go build ./...`, `go vet ./...`, `gofmt -l` clean
- [ ] `go test ./internal/service/ -run "DailyGoal"` all pass

---

### Task 7: Commit / Push / PR

Branch `feat/s3-03-daily-goal-api`; commit (identity via `-c` override), push, PR via REST API, body ends `Closes #16`.
