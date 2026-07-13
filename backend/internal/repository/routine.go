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
