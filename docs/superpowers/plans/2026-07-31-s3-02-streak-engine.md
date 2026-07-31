# S3-02 Streak Calculation Engine + API Implementation Plan

> Steps use checkbox (`- [ ]`) syntax. Execute top to bottom. Commands are Windows PowerShell. Log all service-layer errors via `logger.Get()` (project rule #3).

**Goal:** Server-side streak engine that recomputes a routine's current/longest streak from `routine_logs` on every log/undo event, persists to `streaks`, and exposes read endpoints. Closes #15.

**Tech:** Go, Fiber v2, GORM v2, PostgreSQL, testify.

## Global Constraints

- handler → service → repository; thin handlers (`parseUserID`, `parseParamID`, `success`, `handleServiceError`)
- Dual service constructors: `NewStreakService` (concrete, main.go) + `NewStreakServiceIface` (interfaces, tests)
- Repos return `nil, nil` when not found
- No commits until instructed · Branch: `feat/s3-02-streak-engine`

---

### Task 1: StreakRepository

**File:** Create `backend/internal/repository/streak.go`

- [ ] **Step 1:**
```go
package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/google/uuid"
)

type StreakRepository struct {
	db *gorm.DB
}

func NewStreakRepository(db *gorm.DB) *StreakRepository {
	return &StreakRepository{db: db}
}

// Upsert inserts or updates the single streak row for a routine (unique routine_id).
func (r *StreakRepository) Upsert(s *model.Streak) error {
	sql := `
		INSERT INTO streaks (id, user_id, routine_id, current_streak, longest_streak, last_completed, updated_at)
		VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, NOW())
		ON CONFLICT (routine_id)
		DO UPDATE SET current_streak = EXCLUDED.current_streak,
		              longest_streak = EXCLUDED.longest_streak,
		              last_completed = EXCLUDED.last_completed,
		              updated_at = NOW()
		RETURNING id, user_id, routine_id, current_streak, longest_streak, last_completed, updated_at`
	return r.db.Raw(sql, s.UserID, s.RoutineID, s.CurrentStreak, s.LongestStreak, s.LastCompleted).Scan(s).Error
}

// FindByRoutineID returns nil, nil when not found.
func (r *StreakRepository) FindByRoutineID(routineID uuid.UUID) (*model.Streak, error) {
	var s model.Streak
	err := r.db.Where("routine_id = ?", routineID).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// FindByUserID returns all streaks for a user (possibly empty).
func (r *StreakRepository) FindByUserID(userID uuid.UUID) ([]model.Streak, error) {
	var streaks []model.Streak
	err := r.db.Where("user_id = ?", userID).Find(&streaks).Error
	if err != nil {
		return nil, err
	}
	return streaks, nil
}
```

---

### Task 2: RoutineLogRepository.ListDatesByRoutine

**File:** Modify `backend/internal/repository/routine_log.go` (append method; imports already include `time`, `model`, `uuid`)

- [ ] **Step 1:**
```go
// ListDatesByRoutine returns all logged_at dates for a routine, ordered ascending.
func (r *RoutineLogRepository) ListDatesByRoutine(routineID uuid.UUID) ([]time.Time, error) {
	var dates []time.Time
	err := r.db.Model(&model.RoutineLog{}).
		Where("routine_id = ?", routineID).
		Order("logged_at ASC").
		Pluck("logged_at", &dates).Error
	if err != nil {
		return nil, err
	}
	return dates, nil
}
```

---

### Task 3: StreakService + CalculateStreak

**File:** Create `backend/internal/service/streak.go` (full contents in repo; key parts below)

- `CalculateStreak(periodType string, logDates []time.Time) (current, longest int, lastCompleted *time.Time)` — dedupe→period keys (daily=date, weekly=Monday, monthly=1st)→sort→longest run + run-ending-at-latest.
- `Recalculate` — verify ownership; `ListDatesByRoutine`; `CalculateStreak`; longest = max(computed, existing) (high-water); `Upsert`. Errors logged + `INTERNAL_ERROR`.
- `GetByRoutine` — ownership; persisted streak or zero-value if none.
- `ListByUser` — all streaks.

Adjacency helpers: daily `+1d`, weekly `+7d` (week starts), monthly next calendar month. `startOfWeek` = Monday, UTC.

---

### Task 4: Handlers

**File:** Create `backend/internal/handler/streak.go` — `GetRoutineStreak`, `ListStreaks` (thin, 200 on success).

**File:** Modify `backend/internal/handler/routine.go` — `LogRoutine` and `DeleteRoutineLog` gain a `streakService *service.StreakService` param; after success call `streakService.Recalculate(userID, routineID)` best-effort (`_, _ =`), errors logged in-service.

---

### Task 5: Wire routes

**File:** Modify `backend/cmd/server/main.go`

- [ ] **Step 1:** After `routineLogService := ...`:
```go
streakRepo := repository.NewStreakRepository(db)
streakService := service.NewStreakService(streakRepo, routineLogRepo, routineRepo)
```
- [ ] **Step 2:** Update log routes + add streak routes:
```go
routines.Post("/:id/log", handler.LogRoutine(routineLogService, streakService))
routines.Delete("/:id/log/:logId", handler.DeleteRoutineLog(routineLogService, streakService))
routines.Get("/:id/log", handler.GetRoutineLog(routineLogService))
routines.Get("/:id/streak", handler.GetRoutineStreak(streakService))

streaks := api.Group("/streaks", middleware.RequireAuth(cfg.JWTSecret))
streaks.Get("/", handler.ListStreaks(streakService))
```

---

### Task 6: Unit tests

**File:** Create `backend/internal/service/streak_test.go` (`package service_test`)

- `CalculateStreak`: daily consecutive / daily gap / unsorted+dupes / empty / weekly consecutive / weekly same-week / weekly gap / monthly consecutive / monthly gap.
- Service (testify mocks for the 3 interfaces): Recalculate success, high-water longest preserved, routine-not-found → `NOT_FOUND`; GetByRoutine zero-when-none; ListByUser.

---

### Task 7: Verify

- [ ] `cd backend; go build ./...` → clean
- [ ] `go vet ./...` → clean
- [ ] `go test ./internal/service/ -run "Streak|Calculate|Recalculate|GetByRoutine|ListByUser" -v` → all pass
- [ ] (optional) migrate up on dev DB + curl `GET /routines/:id/streak` after logging

---

### Task 8: Commit / Push / PR (only when instructed)

```powershell
git add backend/internal/repository/streak.go backend/internal/repository/routine_log.go backend/internal/service/streak.go backend/internal/service/streak_test.go backend/internal/handler/streak.go backend/internal/handler/routine.go backend/cmd/server/main.go docs/superpowers/specs/2026-07-31-s3-02-streak-engine-design.md docs/superpowers/plans/2026-07-31-s3-02-streak-engine.md
# commit with -c user.name/email override; message ends with "Closes #15"
```
PR via GitHub REST API (token from `.gemini/settings.json`), base `main`, head `feat/s3-02-streak-engine`, body ends `Closes #15`.
