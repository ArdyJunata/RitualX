package service

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

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

// ── Request types ─────────────────────────────────────────────────────────────

type CreateRoutineRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	PeriodType  string `json:"period_type"`
	TargetCount int    `json:"target_count"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
}

// UpdateRoutineRequest uses pointer fields so nil = "not provided" (partial update).
type UpdateRoutineRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	PeriodType  *string `json:"period_type"`
	TargetCount *int    `json:"target_count"`
	Icon        *string `json:"icon"`
	Color       *string `json:"color"`
}

type ReorderRoutineItem struct {
	ID        uuid.UUID `json:"id"`
	SortOrder int       `json:"sort_order"`
}

type ReorderRoutineRequest struct {
	Order []ReorderRoutineItem `json:"order"`
}

// ── Service ───────────────────────────────────────────────────────────────────

type RoutineService struct {
	routineRepo *repository.RoutineRepository
}

func NewRoutineService(routineRepo *repository.RoutineRepository) *RoutineService {
	return &RoutineService{routineRepo: routineRepo}
}

// Create — create a new routine for the user.
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

// List — return all active routines for the user ordered by sort_order.
func (s *RoutineService) List(userID uuid.UUID) ([]model.Routine, *ServiceError) {
	log := logger.Get()

	routines, err := s.routineRepo.FindAllByUserID(userID)
	if err != nil {
		log.Error("failed to list routines", "error", err, "user_id", userID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to list routines"}
	}

	return routines, nil
}

// GetByID — return a single routine owned by the user.
func (s *RoutineService) GetByID(userID, routineID uuid.UUID) (*model.Routine, *ServiceError) {
	log := logger.Get()

	routine, err := s.routineRepo.FindByIDAndUserID(routineID, userID)
	if err != nil {
		log.Error("failed to get routine", "error", err, "routine_id", routineID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to get routine"}
	}
	if routine == nil {
		return nil, &ServiceError{Code: "NOT_FOUND", Message: "routine not found"}
	}

	return routine, nil
}

// Update — apply only the non-nil fields from req to the routine.
func (s *RoutineService) Update(userID, routineID uuid.UUID, req UpdateRoutineRequest) (*model.Routine, *ServiceError) {
	log := logger.Get()

	if errs := validateUpdateRoutine(req); len(errs) > 0 {
		return nil, &ServiceError{
			Code:    "VALIDATION_ERROR",
			Message: "validation failed",
			Details: errs,
		}
	}

	routine, err := s.routineRepo.FindByIDAndUserID(routineID, userID)
	if err != nil {
		log.Error("failed to find routine for update", "error", err, "routine_id", routineID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to update routine"}
	}
	if routine == nil {
		return nil, &ServiceError{Code: "NOT_FOUND", Message: "routine not found"}
	}

	if req.Title != nil {
		routine.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		routine.Description = *req.Description
	}
	if req.PeriodType != nil {
		routine.PeriodType = *req.PeriodType
	}
	if req.TargetCount != nil {
		routine.TargetCount = *req.TargetCount
	}
	if req.Icon != nil {
		routine.Icon = *req.Icon
	}
	if req.Color != nil {
		routine.Color = *req.Color
	}

	if err := s.routineRepo.Update(routine); err != nil {
		log.Error("failed to update routine", "error", err, "routine_id", routineID)
		return nil, &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to update routine"}
	}

	return routine, nil
}

// Delete — soft-delete the routine (is_active = false).
func (s *RoutineService) Delete(userID, routineID uuid.UUID) *ServiceError {
	log := logger.Get()

	routine, err := s.routineRepo.FindByIDAndUserID(routineID, userID)
	if err != nil {
		log.Error("failed to find routine for delete", "error", err, "routine_id", routineID)
		return &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to delete routine"}
	}
	if routine == nil {
		return &ServiceError{Code: "NOT_FOUND", Message: "routine not found"}
	}

	if err := s.routineRepo.SoftDelete(routineID); err != nil {
		log.Error("failed to soft-delete routine", "error", err, "routine_id", routineID)
		return &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to delete routine"}
	}

	return nil
}

// Reorder — update sort_order for multiple routines in one transaction.
func (s *RoutineService) Reorder(userID uuid.UUID, req ReorderRoutineRequest) *ServiceError {
	log := logger.Get()

	if errs := validateReorderRoutine(req); len(errs) > 0 {
		return &ServiceError{
			Code:    "VALIDATION_ERROR",
			Message: "validation failed",
			Details: errs,
		}
	}

	items := make([]repository.ReorderItem, len(req.Order))
	for i, o := range req.Order {
		items[i] = repository.ReorderItem{ID: o.ID, SortOrder: o.SortOrder}
	}

	if err := s.routineRepo.Reorder(userID, items); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &ServiceError{Code: "NOT_FOUND", Message: "one or more routines not found"}
		}
		log.Error("failed to reorder routines", "error", err, "user_id", userID)
		return &ServiceError{Code: "INTERNAL_ERROR", Message: "failed to reorder routines"}
	}

	return nil
}

// ── Validators ────────────────────────────────────────────────────────────────

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

func validateUpdateRoutine(req UpdateRoutineRequest) []FieldError {
	var errs []FieldError

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			errs = append(errs, FieldError{Field: "title", Message: "title cannot be empty"})
		} else if len(title) > 100 {
			errs = append(errs, FieldError{Field: "title", Message: "title must be at most 100 characters"})
		}
	}

	if req.PeriodType != nil && !validPeriodTypes[*req.PeriodType] {
		errs = append(errs, FieldError{Field: "period_type", Message: "period_type must be daily, weekly, or monthly"})
	}

	if req.TargetCount != nil && *req.TargetCount < 1 {
		errs = append(errs, FieldError{Field: "target_count", Message: "target_count must be at least 1"})
	}

	return errs
}

func validateReorderRoutine(req ReorderRoutineRequest) []FieldError {
	var errs []FieldError

	if len(req.Order) == 0 {
		errs = append(errs, FieldError{Field: "order", Message: "order must not be empty"})
		return errs
	}

	for i, item := range req.Order {
		if item.SortOrder < 0 {
			errs = append(errs, FieldError{
				Field:   "order",
				Message: fmt.Sprintf("sort_order must be >= 0 at index %d", i),
			})
		}
	}

	return errs
}
