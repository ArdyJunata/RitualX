package service

import (
	"strings"

	"github.com/google/uuid"

	"github.com/ArdyJunata/RitualX/backend/internal/logger"
	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/ArdyJunata/RitualX/backend/internal/repository"
)

var validPeriodTypes = map[string]bool{
	"daily":   true,
	"weekly":  true,
	"monthly": true,
}

type CreateRoutineRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	PeriodType  string `json:"period_type"`
	TargetCount int    `json:"target_count"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
}

type RoutineService struct {
	routineRepo *repository.RoutineRepository
}

func NewRoutineService(routineRepo *repository.RoutineRepository) *RoutineService {
	return &RoutineService{routineRepo: routineRepo}
}

func (s *RoutineService) Create(userID uuid.UUID, req CreateRoutineRequest) (*model.Routine, error) {
	log := logger.Get()

	if errs := validateCreateRoutine(req); len(errs) > 0 {
		return nil, &ServiceError{
			Code:    "VALIDATION_ERROR",
			Message: "validation failed",
			Details: errs,
		}
	}

	routine := &model.Routine{
		UserID:      userID,
		Title:       strings.TrimSpace(req.Title),
		Description: req.Description,
		PeriodType:  req.PeriodType,
		TargetCount: req.TargetCount,
		Icon:        req.Icon,
		Color:       req.Color,
	}

	if err := s.routineRepo.Create(routine); err != nil {
		log.Error("failed to create routine", "error", err, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to create routine"}
	}

	return routine, nil
}

func validateCreateRoutine(req CreateRoutineRequest) []FieldError {
	var errs []FieldError

	title := strings.TrimSpace(req.Title)
	if title == "" {
		errs = append(errs, FieldError{Field: "title", Message: "title is required"})
	} else if len(title) > 100 {
		errs = append(errs, FieldError{Field: "title", Message: "title must be at most 100 characters"})
	}

	if req.PeriodType == "" {
		errs = append(errs, FieldError{Field: "period_type", Message: "period_type is required"})
	} else if !validPeriodTypes[req.PeriodType] {
		errs = append(errs, FieldError{Field: "period_type", Message: "period_type must be daily, weekly, or monthly"})
	}

	if req.TargetCount < 1 {
		errs = append(errs, FieldError{Field: "target_count", Message: "target_count must be at least 1"})
	}

	return errs
}
