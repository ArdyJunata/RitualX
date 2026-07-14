package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/google/uuid"
)

type RoutineRepository struct {
	db *gorm.DB
}

func NewRoutineRepository(db *gorm.DB) *RoutineRepository {
	return &RoutineRepository{db: db}
}

// ReorderItem is used by the Reorder method to map ID → new sort_order.
type ReorderItem struct {
	ID        uuid.UUID
	SortOrder int
}

func (r *RoutineRepository) Create(routine *model.Routine) error {
	return r.db.Create(routine).Error
}

func (r *RoutineRepository) FindByID(id uuid.UUID) (*model.Routine, error) {
	var routine model.Routine
	err := r.db.Where("id = ?", id).First(&routine).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &routine, nil
}

// FindAllByUserID returns all active routines for a user, ordered by sort_order ASC.
func (r *RoutineRepository) FindAllByUserID(userID uuid.UUID) ([]model.Routine, error) {
	var routines []model.Routine
	err := r.db.
		Where("user_id = ? AND is_active = true", userID).
		Order("sort_order ASC").
		Find(&routines).Error
	if err != nil {
		return nil, err
	}
	return routines, nil
}

// FindByIDAndUserID returns a routine only if it belongs to the given user and is active.
// Returns nil, nil when not found (caller maps nil → NOT_FOUND).
func (r *RoutineRepository) FindByIDAndUserID(id, userID uuid.UUID) (*model.Routine, error) {
	var routine model.Routine
	err := r.db.Where("id = ? AND user_id = ? AND is_active = true", id, userID).First(&routine).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &routine, nil
}

// Update persists all fields of the given routine (full save).
func (r *RoutineRepository) Update(routine *model.Routine) error {
	return r.db.Save(routine).Error
}

// SoftDelete sets is_active = false for the given routine ID.
func (r *RoutineRepository) SoftDelete(id uuid.UUID) error {
	return r.db.Model(&model.Routine{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"is_active": false}).Error
}

// Reorder updates sort_order for multiple routines in a single transaction.
// Returns gorm.ErrRecordNotFound if any item is not found/not owned by userID.
func (r *RoutineRepository) Reorder(userID uuid.UUID, items []ReorderItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			result := tx.Model(&model.Routine{}).
				Where("id = ? AND user_id = ? AND is_active = true", item.ID, userID).
				Update("sort_order", item.SortOrder)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}
		return nil
	})
}
