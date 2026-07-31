package service

import (
	"time"

	"github.com/google/uuid"

	"github.com/ArdyJunata/RitualX/backend/internal/logger"
	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/ArdyJunata/RitualX/backend/internal/repository"
)

// ── Interfaces (for testability) ──────────────────────────────────────────────

type StatsLogRepoIface interface {
	ListByRoutineAndDateRange(routineID uuid.UUID, from, to time.Time) ([]model.RoutineLog, error)
	SumCountByRoutineAndDateRange(routineID uuid.UUID, from, to time.Time) (int, error)
}

type StatsRoutineRepoIface interface {
	FindByIDAndUserID(id, userID uuid.UUID) (*model.Routine, error)
}

// ProgressResponse is the payload for GET /routines/:id/progress.
type ProgressResponse struct {
	Completed int    `json:"completed"`
	Target    int    `json:"target"`
	Period    string `json:"period"`
}

// ── Service ───────────────────────────────────────────────────────────────────

type StatsService struct {
	logRepo     StatsLogRepoIface
	routineRepo StatsRoutineRepoIface
}

// NewStatsService wires concrete repository types (used in main.go).
func NewStatsService(logRepo *repository.RoutineLogRepository, routineRepo *repository.RoutineRepository) *StatsService {
	return &StatsService{logRepo: logRepo, routineRepo: routineRepo}
}

// NewStatsServiceIface wires interface types (used in tests).
func NewStatsServiceIface(logRepo StatsLogRepoIface, routineRepo StatsRoutineRepoIface) *StatsService {
	return &StatsService{logRepo: logRepo, routineRepo: routineRepo}
}

// GetHeatmap returns a date(YYYY-MM-DD)->count map for the routine within the given year.
// Only days with a log appear (routine_logs is unique per routine+date).
func (s *StatsService) GetHeatmap(userID, routineID uuid.UUID, year int) (map[string]int, *ServiceError) {
	log := logger.Get()

	routine, err := s.routineRepo.FindByIDAndUserID(routineID, userID)
	if err != nil {
		log.Error("failed to find routine for heatmap", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch heatmap"}
	}
	if routine == nil {
		return nil, &ServiceError{Code: "NOT_FOUND", Message: "routine not found"}
	}

	from := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)

	logs, err := s.logRepo.ListByRoutineAndDateRange(routineID, from, to)
	if err != nil {
		log.Error("failed to list logs for heatmap", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch heatmap"}
	}

	m := make(map[string]int, len(logs))
	for _, l := range logs {
		m[l.LoggedAt.UTC().Format("2006-01-02")] = l.Count
	}
	return m, nil
}

// GetProgress returns current-period progress for a routine (completed vs target).
func (s *StatsService) GetProgress(userID, routineID uuid.UUID) (*ProgressResponse, *ServiceError) {
	log := logger.Get()

	routine, err := s.routineRepo.FindByIDAndUserID(routineID, userID)
	if err != nil {
		log.Error("failed to find routine for progress", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch progress"}
	}
	if routine == nil {
		return nil, &ServiceError{Code: "NOT_FOUND", Message: "routine not found"}
	}

	from, to := CurrentPeriodRange(routine.PeriodType, time.Now().UTC())
	completed, err := s.logRepo.SumCountByRoutineAndDateRange(routineID, from, to)
	if err != nil {
		log.Error("failed to sum progress", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch progress"}
	}

	return &ProgressResponse{
		Completed: completed,
		Target:    routine.TargetCount,
		Period:    routine.PeriodType,
	}, nil
}

// CurrentPeriodRange returns the inclusive [from, to] date range of the current
// period (by period type) containing now. Reuses startOfWeek/startOfDayUTC.
func CurrentPeriodRange(periodType string, now time.Time) (time.Time, time.Time) {
	day := startOfDayUTC(now)
	switch periodType {
	case "weekly":
		start := startOfWeek(day)
		return start, start.AddDate(0, 0, 6)
	case "monthly":
		start := time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, time.UTC)
		return start, start.AddDate(0, 1, -1)
	default: // daily
		return day, day
	}
}
