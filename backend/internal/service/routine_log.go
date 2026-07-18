package service

import (
	"time"

	"github.com/google/uuid"

	"github.com/ArdyJunata/RitualX/backend/internal/logger"
	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/ArdyJunata/RitualX/backend/internal/repository"
)

// ── Interfaces (for testability) ──────────────────────────────────────────────

type RoutineLogRepoIface interface {
	Upsert(log *model.RoutineLog) error
	FindByID(id uuid.UUID) (*model.RoutineLog, error)
	FindTodayByRoutineAndUser(routineID, userID uuid.UUID) (*model.RoutineLog, error)
	Delete(id uuid.UUID) error
}

type RoutineRepoIface interface {
	FindByIDAndUserID(id, userID uuid.UUID) (*model.Routine, error)
}

// ── Request type ──────────────────────────────────────────────────────────────

// LogRoutineRequest is the JSON body for POST /routines/:id/log.
// LoggedAt defaults to today (UTC) if empty. Count defaults to 1 if nil or 0.
type LogRoutineRequest struct {
	LoggedAt string  `json:"logged_at"`
	Count    *int    `json:"count"`
	Note     *string `json:"note"`
}

// ── Service ───────────────────────────────────────────────────────────────────

type RoutineLogService struct {
	logRepo     RoutineLogRepoIface
	routineRepo RoutineRepoIface
}

// NewRoutineLogService wires concrete repository types (used in main.go).
func NewRoutineLogService(logRepo *repository.RoutineLogRepository, routineRepo *repository.RoutineRepository) *RoutineLogService {
	return &RoutineLogService{logRepo: logRepo, routineRepo: routineRepo}
}

// NewRoutineLogServiceIface wires interface types (used in tests).
func NewRoutineLogServiceIface(logRepo RoutineLogRepoIface, routineRepo RoutineRepoIface) *RoutineLogService {
	return &RoutineLogService{logRepo: logRepo, routineRepo: routineRepo}
}

// Log creates or increments a routine completion log for the given date.
func (s *RoutineLogService) Log(userID, routineID uuid.UUID, req LogRoutineRequest) (*model.RoutineLog, *ServiceError) {
	log := logger.Get()

	// Validate + resolve date
	var loggedAt time.Time
	if req.LoggedAt == "" {
		loggedAt = time.Now().UTC().Truncate(24 * time.Hour)
	} else {
		var err error
		loggedAt, err = time.Parse("2006-01-02", req.LoggedAt)
		if err != nil {
			return nil, &ServiceError{
				Code:    "VALIDATION_ERROR",
				Message: "validation failed",
				Details: []FieldError{{Field: "logged_at", Message: "logged_at must be in YYYY-MM-DD format"}},
			}
		}
	}

	// Verify routine ownership
	routine, err := s.routineRepo.FindByIDAndUserID(routineID, userID)
	if err != nil {
		log.Error("failed to find routine for log", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to create log"}
	}
	if routine == nil {
		return nil, &ServiceError{Code: "NOT_FOUND", Message: "routine not found"}
	}

	// Resolve count
	count := 1
	if req.Count != nil && *req.Count > 0 {
		count = *req.Count
	}

	note := ""
	if req.Note != nil {
		note = *req.Note
	}

	entry := &model.RoutineLog{
		RoutineID: routineID,
		UserID:    userID,
		LoggedAt:  loggedAt,
		Count:     count,
		Note:      note,
	}

	if err := s.logRepo.Upsert(entry); err != nil {
		log.Error("failed to upsert routine log", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to create log"}
	}

	return entry, nil
}

// Delete removes a log entry. Verifies the log belongs to the user and routine.
func (s *RoutineLogService) Delete(userID, routineID, logID uuid.UUID) *ServiceError {
	log := logger.Get()

	entry, err := s.logRepo.FindByID(logID)
	if err != nil {
		log.Error("failed to find routine log", "error", err, "log_id", logID, "user_id", userID)
		return &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to delete log"}
	}
	if entry == nil {
		return &ServiceError{Code: "NOT_FOUND", Message: "log not found"}
	}
	if entry.UserID != userID || entry.RoutineID != routineID {
		return &ServiceError{Code: "FORBIDDEN", Message: "access denied"}
	}

	if err := s.logRepo.Delete(logID); err != nil {
		log.Error("failed to delete routine log", "error", err, "log_id", logID, "user_id", userID)
		return &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to delete log"}
	}

	return nil
}

// GetToday returns today's log for the given routine, or nil if none exists.
func (s *RoutineLogService) GetToday(userID, routineID uuid.UUID) (*model.RoutineLog, *ServiceError) {
	log := logger.Get()

	// Verify routine ownership
	routine, err := s.routineRepo.FindByIDAndUserID(routineID, userID)
	if err != nil {
		log.Error("failed to find routine for get-today log", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch log"}
	}
	if routine == nil {
		return nil, &ServiceError{Code: "NOT_FOUND", Message: "routine not found"}
	}

	entry, err := s.logRepo.FindTodayByRoutineAndUser(routineID, userID)
	if err != nil {
		log.Error("failed to fetch today log", "error", err, "routine_id", routineID, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to fetch log"}
	}
	return entry, nil
}
