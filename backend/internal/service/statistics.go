package service

import (
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/ArdyJunata/RitualX/backend/internal/logger"
	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/ArdyJunata/RitualX/backend/internal/repository"
)

// ── Interfaces (for testability) ──────────────────────────────────────────────

type StatisticsLogRepoIface interface {
	SumCountByUser(userID uuid.UUID) (int, error)
	SumCountByUserAndDateRange(userID uuid.UUID, from, to time.Time) (int, error)
	SumCountByRoutine(routineID uuid.UUID) (int, error)
	CountByRoutine(routineID uuid.UUID) (int, error)
}

type StatisticsStreakRepoIface interface {
	FindByUserID(userID uuid.UUID) ([]model.Streak, error)
	FindByRoutineID(routineID uuid.UUID) (*model.Streak, error)
}

type StatisticsRoutineRepoIface interface {
	FindByIDAndUserID(id, userID uuid.UUID) (*model.Routine, error)
	CountActiveByUserID(userID uuid.UUID) (int, error)
}

type StatisticsUserRepoIface interface {
	FindByID(id uuid.UUID) (*model.User, error)
}

// ── Response types ────────────────────────────────────────────────────────────

type OverviewStats struct {
	TotalCompletions int `json:"total_completions"`
	ActiveRoutines   int `json:"active_routines"`
	ActiveStreaks    int `json:"active_streaks"`
	LongestStreak    int `json:"longest_streak"`
	Level            int `json:"level"`
	XP               int `json:"xp"`
}

type RoutineStats struct {
	RoutineID        uuid.UUID  `json:"routine_id"`
	TotalCompletions int        `json:"total_completions"`
	DaysLogged       int        `json:"days_logged"`
	CurrentStreak    int        `json:"current_streak"`
	LongestStreak    int        `json:"longest_streak"`
	CompletionRate   float64    `json:"completion_rate"`
	LastCompleted    *time.Time `json:"last_completed"`
}

type PeriodSummary struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Completions int    `json:"completions"`
}

type SummaryResponse struct {
	This  PeriodSummary `json:"this"`
	Last  PeriodSummary `json:"last"`
	Delta int           `json:"delta"`
}

// ── Service ───────────────────────────────────────────────────────────────────

type StatisticsService struct {
	logRepo     StatisticsLogRepoIface
	streakRepo  StatisticsStreakRepoIface
	routineRepo StatisticsRoutineRepoIface
	userRepo    StatisticsUserRepoIface
}

// NewStatisticsService wires concrete repository types (used in main.go).
func NewStatisticsService(logRepo *repository.RoutineLogRepository, streakRepo *repository.StreakRepository, routineRepo *repository.RoutineRepository, userRepo *repository.UserRepository) *StatisticsService {
	return &StatisticsService{logRepo: logRepo, streakRepo: streakRepo, routineRepo: routineRepo, userRepo: userRepo}
}

// NewStatisticsServiceIface wires interface types (used in tests).
func NewStatisticsServiceIface(logRepo StatisticsLogRepoIface, streakRepo StatisticsStreakRepoIface, routineRepo StatisticsRoutineRepoIface, userRepo StatisticsUserRepoIface) *StatisticsService {
	return &StatisticsService{logRepo: logRepo, streakRepo: streakRepo, routineRepo: routineRepo, userRepo: userRepo}
}

// GetOverview returns overall stats for the user.
func (s *StatisticsService) GetOverview(userID uuid.UUID) (*OverviewStats, *ServiceError) {
	log := logger.Get()

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		log.Error("failed to load user for overview", "error", err, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch overview"}
	}
	if user == nil {
		return nil, &ServiceError{Code: "NOT_FOUND", Message: "user not found"}
	}

	total, err := s.logRepo.SumCountByUser(userID)
	if err != nil {
		log.Error("failed to sum user completions", "error", err, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch overview"}
	}

	activeRoutines, err := s.routineRepo.CountActiveByUserID(userID)
	if err != nil {
		log.Error("failed to count active routines", "error", err, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch overview"}
	}

	streaks, err := s.streakRepo.FindByUserID(userID)
	if err != nil {
		log.Error("failed to load streaks for overview", "error", err, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch overview"}
	}
	activeStreaks := 0
	longest := 0
	for _, st := range streaks {
		if st.CurrentStreak > 0 {
			activeStreaks++
		}
		if st.LongestStreak > longest {
			longest = st.LongestStreak
		}
	}

	return &OverviewStats{
		TotalCompletions: total,
		ActiveRoutines:   activeRoutines,
		ActiveStreaks:    activeStreaks,
		LongestStreak:    longest,
		Level:            user.Level,
		XP:               user.XP,
	}, nil
}

// GetRoutineStats returns per-routine statistics. Ownership-checked.
func (s *StatisticsService) GetRoutineStats(userID, routineID uuid.UUID) (*RoutineStats, *ServiceError) {
	log := logger.Get()

	routine, err := s.routineRepo.FindByIDAndUserID(routineID, userID)
	if err != nil {
		log.Error("failed to find routine for stats", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch routine stats"}
	}
	if routine == nil {
		return nil, &ServiceError{Code: "NOT_FOUND", Message: "routine not found"}
	}

	total, err := s.logRepo.SumCountByRoutine(routineID)
	if err != nil {
		log.Error("failed to sum routine completions", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch routine stats"}
	}

	days, err := s.logRepo.CountByRoutine(routineID)
	if err != nil {
		log.Error("failed to count routine days", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch routine stats"}
	}

	streak, err := s.streakRepo.FindByRoutineID(routineID)
	if err != nil {
		log.Error("failed to load streak for routine stats", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch routine stats"}
	}
	current, longest := 0, 0
	var lastCompleted *time.Time
	if streak != nil {
		current = streak.CurrentStreak
		longest = streak.LongestStreak
		lastCompleted = streak.LastCompleted
	}

	// completion_rate = logged days / active days since creation (capped at 1, 4dp).
	created := startOfDayUTC(routine.CreatedAt)
	today := startOfDayUTC(time.Now().UTC())
	activeDays := int(today.Sub(created).Hours()/24) + 1
	if activeDays < 1 {
		activeDays = 1
	}
	rate := float64(days) / float64(activeDays)
	if rate > 1 {
		rate = 1
	}
	rate = math.Round(rate*10000) / 10000

	return &RoutineStats{
		RoutineID:        routineID,
		TotalCompletions: total,
		DaysLogged:       days,
		CurrentStreak:    current,
		LongestStreak:    longest,
		CompletionRate:   rate,
		LastCompleted:    lastCompleted,
	}, nil
}

// GetWeeklySummary compares this ISO week to last week.
func (s *StatisticsService) GetWeeklySummary(userID uuid.UUID) (*SummaryResponse, *ServiceError) {
	return s.summary(userID, "weekly")
}

// GetMonthlySummary compares this calendar month to last month.
func (s *StatisticsService) GetMonthlySummary(userID uuid.UUID) (*SummaryResponse, *ServiceError) {
	return s.summary(userID, "monthly")
}

// summary computes this-period vs last-period completion totals for the user.
// The current period sum is queried before the previous one.
func (s *StatisticsService) summary(userID uuid.UUID, periodType string) (*SummaryResponse, *ServiceError) {
	log := logger.Get()
	now := time.Now().UTC()

	thisFrom, thisTo := CurrentPeriodRange(periodType, now)
	var lastFrom, lastTo time.Time
	if periodType == "monthly" {
		lastFrom = thisFrom.AddDate(0, -1, 0)
		lastTo = thisFrom.AddDate(0, 0, -1)
	} else { // weekly
		lastFrom = thisFrom.AddDate(0, 0, -7)
		lastTo = thisFrom.AddDate(0, 0, -1)
	}

	thisSum, err := s.logRepo.SumCountByUserAndDateRange(userID, thisFrom, thisTo)
	if err != nil {
		log.Error("failed to sum current period", "error", err, "user_id", userID, "period", periodType)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch summary"}
	}
	lastSum, err := s.logRepo.SumCountByUserAndDateRange(userID, lastFrom, lastTo)
	if err != nil {
		log.Error("failed to sum previous period", "error", err, "user_id", userID, "period", periodType)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch summary"}
	}

	return &SummaryResponse{
		This:  PeriodSummary{From: thisFrom.Format("2006-01-02"), To: thisTo.Format("2006-01-02"), Completions: thisSum},
		Last:  PeriodSummary{From: lastFrom.Format("2006-01-02"), To: lastTo.Format("2006-01-02"), Completions: lastSum},
		Delta: thisSum - lastSum,
	}, nil
}
