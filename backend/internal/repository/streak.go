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

// Upsert inserts or updates the single streak row for a routine
// (one row per routine, enforced by the unique index on routine_id).
// It re-fetches the row via RETURNING so the caller gets the final state.
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
	return r.db.Raw(sql,
		s.UserID,
		s.RoutineID,
		s.CurrentStreak,
		s.LongestStreak,
		s.LastCompleted,
	).Scan(s).Error
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
