package service

import (
	"time"

	"github.com/google/uuid"

	"github.com/ArdyJunata/RitualX/backend/internal/logger"
	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/ArdyJunata/RitualX/backend/internal/repository"
)

// ── Interface (for testability) ───────────────────────────────────────────────

type DailyGoalRepoIface interface {
	Upsert(g *model.DailyGoal) error
	CountActiveRoutines(userID uuid.UUID) (int, error)
	CountCompletedRoutines(userID uuid.UUID, date time.Time) (int, error)
	FindByUserAndDateRange(userID uuid.UUID, from, to time.Time) ([]model.DailyGoal, error)
}

// ── Service ───────────────────────────────────────────────────────────────────

type DailyGoalService struct {
	repo DailyGoalRepoIface
}

// NewDailyGoalService wires the concrete repository (used in main.go).
func NewDailyGoalService(repo *repository.DailyGoalRepository) *DailyGoalService {
	return &DailyGoalService{repo: repo}
}

// NewDailyGoalServiceIface wires an interface (used in tests).
func NewDailyGoalServiceIface(repo DailyGoalRepoIface) *DailyGoalService {
	return &DailyGoalService{repo: repo}
}

// compute builds a DailyGoal for the given date from current counts (no persistence).
func (s *DailyGoalService) compute(userID uuid.UUID, date time.Time) (*model.DailyGoal, error) {
	total, err := s.repo.CountActiveRoutines(userID)
	if err != nil {
		return nil, err
	}
	completed, err := s.repo.CountCompletedRoutines(userID, date)
	if err != nil {
		return nil, err
	}
	return &model.DailyGoal{
		UserID:        userID,
		Date:          date,
		TotalRoutines: total,
		Completed:     completed,
		IsAchieved:    total > 0 && completed >= total,
	}, nil
}

// Recalculate computes and persists the daily goal for (user, date).
// Runs on each log/undo event.
func (s *DailyGoalService) Recalculate(userID uuid.UUID, date time.Time) (*model.DailyGoal, *ServiceError) {
	log := logger.Get()
	date = startOfDayUTC(date)

	goal, err := s.compute(userID, date)
	if err != nil {
		log.Error("failed to compute daily goal", "error", err, "user_id", userID, "date", date)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to update daily goal"}
	}
	if err := s.repo.Upsert(goal); err != nil {
		log.Error("failed to upsert daily goal", "error", err, "user_id", userID, "date", date)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to update daily goal"}
	}
	return goal, nil
}

// GetToday returns today's goal summary, computed live (not persisted) so it
// reflects the current active-routine count.
func (s *DailyGoalService) GetToday(userID uuid.UUID) (*model.DailyGoal, *ServiceError) {
	log := logger.Get()
	today := startOfDayUTC(time.Now().UTC())

	goal, err := s.compute(userID, today)
	if err != nil {
		log.Error("failed to compute today daily goal", "error", err, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch daily goal"}
	}
	return goal, nil
}

// GetHistory returns persisted daily goals within [from, to] (inclusive).
func (s *DailyGoalService) GetHistory(userID uuid.UUID, from, to time.Time) ([]model.DailyGoal, *ServiceError) {
	log := logger.Get()
	from = startOfDayUTC(from)
	to = startOfDayUTC(to)

	goals, err := s.repo.FindByUserAndDateRange(userID, from, to)
	if err != nil {
		log.Error("failed to fetch daily goal history", "error", err, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch daily goal history"}
	}
	return goals, nil
}

// startOfDayUTC truncates a timestamp to 00:00 UTC of that calendar day.
func startOfDayUTC(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
