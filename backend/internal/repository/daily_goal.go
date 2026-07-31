package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/google/uuid"
)

type DailyGoalRepository struct {
	db *gorm.DB
}

func NewDailyGoalRepository(db *gorm.DB) *DailyGoalRepository {
	return &DailyGoalRepository{db: db}
}

// Upsert inserts or updates the daily goal row for (user_id, date).
// One row per (user_id, date), enforced by the unique index from migration 000006.
func (r *DailyGoalRepository) Upsert(g *model.DailyGoal) error {
	sql := `
		INSERT INTO daily_goals (id, user_id, date, total_routines, completed, is_achieved, created_at)
		VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, NOW())
		ON CONFLICT (user_id, date)
		DO UPDATE SET total_routines = EXCLUDED.total_routines,
		              completed = EXCLUDED.completed,
		              is_achieved = EXCLUDED.is_achieved
		RETURNING id, user_id, date, total_routines, completed, is_achieved, created_at`
	return r.db.Raw(sql,
		g.UserID,
		g.Date,
		g.TotalRoutines,
		g.Completed,
		g.IsAchieved,
	).Scan(g).Error
}

// CountActiveRoutines returns the number of active routines for a user.
func (r *DailyGoalRepository) CountActiveRoutines(userID uuid.UUID) (int, error) {
	var count int64
	err := r.db.Model(&model.Routine{}).
		Where("user_id = ? AND is_active = true", userID).
		Count(&count).Error
	return int(count), err
}

// CountCompletedRoutines returns the number of distinct active routines that
// have a log on the given date for the user. routine_logs has one row per
// (routine_id, logged_at), so a count of matching rows equals distinct routines.
func (r *DailyGoalRepository) CountCompletedRoutines(userID uuid.UUID, date time.Time) (int, error) {
	var count int64
	err := r.db.Table("routine_logs AS rl").
		Joins("JOIN routines r ON r.id = rl.routine_id").
		Where("rl.user_id = ? AND rl.logged_at = ? AND r.is_active = true", userID, date).
		Count(&count).Error
	return int(count), err
}

// FindByUserAndDateRange returns persisted daily goals for a user within
// [from, to] (inclusive), ordered by date ASC.
func (r *DailyGoalRepository) FindByUserAndDateRange(userID uuid.UUID, from, to time.Time) ([]model.DailyGoal, error) {
	var goals []model.DailyGoal
	err := r.db.
		Where("user_id = ? AND date >= ? AND date <= ?", userID, from, to).
		Order("date ASC").
		Find(&goals).Error
	if err != nil {
		return nil, err
	}
	return goals, nil
}
