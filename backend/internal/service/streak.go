package service

import (
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/ArdyJunata/RitualX/backend/internal/logger"
	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/ArdyJunata/RitualX/backend/internal/repository"
)

// ── Interfaces (for testability) ──────────────────────────────────────────────

type StreakRepoIface interface {
	Upsert(s *model.Streak) error
	FindByRoutineID(routineID uuid.UUID) (*model.Streak, error)
	FindByUserID(userID uuid.UUID) ([]model.Streak, error)
}

type StreakLogRepoIface interface {
	ListDatesByRoutine(routineID uuid.UUID) ([]time.Time, error)
}

type StreakRoutineRepoIface interface {
	FindByIDAndUserID(id, userID uuid.UUID) (*model.Routine, error)
}

// ── Service ───────────────────────────────────────────────────────────────────

type StreakService struct {
	streakRepo  StreakRepoIface
	logRepo     StreakLogRepoIface
	routineRepo StreakRoutineRepoIface
}

// NewStreakService wires concrete repository types (used in main.go).
func NewStreakService(streakRepo *repository.StreakRepository, logRepo *repository.RoutineLogRepository, routineRepo *repository.RoutineRepository) *StreakService {
	return &StreakService{streakRepo: streakRepo, logRepo: logRepo, routineRepo: routineRepo}
}

// NewStreakServiceIface wires interface types (used in tests).
func NewStreakServiceIface(streakRepo StreakRepoIface, logRepo StreakLogRepoIface, routineRepo StreakRoutineRepoIface) *StreakService {
	return &StreakService{streakRepo: streakRepo, logRepo: logRepo, routineRepo: routineRepo}
}

// Recalculate recomputes and persists the streak for a routine from its logs.
// Runs on each log/undo event (server-side). Ownership-checked.
func (s *StreakService) Recalculate(userID, routineID uuid.UUID) (*model.Streak, *ServiceError) {
	log := logger.Get()

	routine, err := s.routineRepo.FindByIDAndUserID(routineID, userID)
	if err != nil {
		log.Error("failed to find routine for streak recalc", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to update streak"}
	}
	if routine == nil {
		return nil, &ServiceError{Code: "NOT_FOUND", Message: "routine not found"}
	}

	dates, err := s.logRepo.ListDatesByRoutine(routineID)
	if err != nil {
		log.Error("failed to list log dates for streak recalc", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to update streak"}
	}

	current, longest, lastCompleted := CalculateStreak(routine.PeriodType, dates)

	// High-water mark: the persisted longest_streak must never decrease.
	existing, err := s.streakRepo.FindByRoutineID(routineID)
	if err != nil {
		log.Error("failed to load existing streak", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to update streak"}
	}
	if existing != nil && existing.LongestStreak > longest {
		longest = existing.LongestStreak
	}

	streak := &model.Streak{
		UserID:        userID,
		RoutineID:     routineID,
		CurrentStreak: current,
		LongestStreak: longest,
		LastCompleted: lastCompleted,
	}
	if err := s.streakRepo.Upsert(streak); err != nil {
		log.Error("failed to upsert streak", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to update streak"}
	}

	return streak, nil
}

// GetByRoutine returns the persisted streak for a routine, or a zero-value
// streak if none has been computed yet. Ownership-checked.
func (s *StreakService) GetByRoutine(userID, routineID uuid.UUID) (*model.Streak, *ServiceError) {
	log := logger.Get()

	routine, err := s.routineRepo.FindByIDAndUserID(routineID, userID)
	if err != nil {
		log.Error("failed to find routine for streak get", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch streak"}
	}
	if routine == nil {
		return nil, &ServiceError{Code: "NOT_FOUND", Message: "routine not found"}
	}

	streak, err := s.streakRepo.FindByRoutineID(routineID)
	if err != nil {
		log.Error("failed to fetch streak", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch streak"}
	}
	if streak == nil {
		// No logs yet — return a zero-value streak so the client always gets a shape.
		return &model.Streak{UserID: userID, RoutineID: routineID}, nil
	}
	return streak, nil
}

// ListByUser returns all persisted streaks for the user.
func (s *StreakService) ListByUser(userID uuid.UUID) ([]model.Streak, *ServiceError) {
	log := logger.Get()

	streaks, err := s.streakRepo.FindByUserID(userID)
	if err != nil {
		log.Error("failed to list streaks", "error", err, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch streaks"}
	}
	return streaks, nil
}

// ── Pure calculation ──────────────────────────────────────────────────────────

// CalculateStreak folds log dates into completed periods (by period type) and
// returns the current run (ending at the latest period), the longest run over
// all history, and the latest log date. Input need not be sorted or deduped.
// A period counts as completed if it contains at least one log.
func CalculateStreak(periodType string, logDates []time.Time) (current int, longest int, lastCompleted *time.Time) {
	if len(logDates) == 0 {
		return 0, 0, nil
	}

	// Latest actual log date → last_completed.
	last := logDates[0]
	for _, d := range logDates[1:] {
		if d.After(last) {
			last = d
		}
	}

	// Fold into a set of unique period keys, then sort ascending.
	keySet := make(map[time.Time]struct{}, len(logDates))
	for _, d := range logDates {
		keySet[periodKey(periodType, d)] = struct{}{}
	}
	keys := make([]time.Time, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].Before(keys[j]) })

	// Walk keys, tracking consecutive runs.
	longest = 1
	run := 1
	for i := 1; i < len(keys); i++ {
		if isNextPeriod(periodType, keys[i-1], keys[i]) {
			run++
		} else {
			run = 1
		}
		if run > longest {
			longest = run
		}
	}
	current = run // the run ending at the latest period key

	lastDate := time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, time.UTC)
	return current, longest, &lastDate
}

// periodKey normalizes a date to a canonical key for its period.
func periodKey(periodType string, d time.Time) time.Time {
	d = d.UTC()
	switch periodType {
	case "weekly":
		return startOfWeek(d)
	case "monthly":
		return time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.UTC)
	default: // daily
		return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	}
}

// startOfWeek returns the Monday (00:00 UTC) of the week containing d.
func startOfWeek(d time.Time) time.Time {
	d = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	// time.Weekday: Sunday=0..Saturday=6. Convert so Monday=0..Sunday=6.
	offset := (int(d.Weekday()) + 6) % 7
	return d.AddDate(0, 0, -offset)
}

// isNextPeriod reports whether next is exactly one period after prev
// (both are period keys produced by periodKey).
func isNextPeriod(periodType string, prev, next time.Time) bool {
	switch periodType {
	case "weekly":
		return next.Equal(prev.AddDate(0, 0, 7))
	case "monthly":
		return next.Equal(prev.AddDate(0, 1, 0))
	default: // daily
		return next.Equal(prev.AddDate(0, 0, 1))
	}
}
