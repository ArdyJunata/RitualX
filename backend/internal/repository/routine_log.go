package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"github.com/google/uuid"
)

type RoutineLogRepository struct {
	db *gorm.DB
}

func NewRoutineLogRepository(db *gorm.DB) *RoutineLogRepository {
	return &RoutineLogRepository{db: db}
}

// Upsert inserts a new log entry. If a log for the same routine+date already
// exists, it increments the count by the incoming log's Count value.
// After upsert it re-fetches the row via RETURNING so the caller gets the final state.
func (r *RoutineLogRepository) Upsert(log *model.RoutineLog) error {
	sql := `
		INSERT INTO routine_logs (id, routine_id, user_id, logged_at, count, note, created_at)
		VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, NOW())
		ON CONFLICT (routine_id, logged_at)
		DO UPDATE SET count = routine_logs.count + EXCLUDED.count
		RETURNING id, routine_id, user_id, logged_at, count, note, created_at`
	return r.db.Raw(sql,
		log.RoutineID,
		log.UserID,
		log.LoggedAt,
		log.Count,
		log.Note,
	).Scan(log).Error
}

// FindByID returns nil, nil when not found.
func (r *RoutineLogRepository) FindByID(id uuid.UUID) (*model.RoutineLog, error) {
	var l model.RoutineLog
	err := r.db.Where("id = ?", id).First(&l).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// FindByRoutineAndDate returns nil, nil when not found.
func (r *RoutineLogRepository) FindByRoutineAndDate(routineID uuid.UUID, date time.Time) (*model.RoutineLog, error) {
	var l model.RoutineLog
	err := r.db.Where("routine_id = ? AND logged_at = ?", routineID, date).First(&l).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// FindTodayByRoutineAndUser returns today's log for the given routine+user, or nil if none.
func (r *RoutineLogRepository) FindTodayByRoutineAndUser(routineID, userID uuid.UUID) (*model.RoutineLog, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	var l model.RoutineLog
	err := r.db.Where("routine_id = ? AND user_id = ? AND logged_at = ?", routineID, userID, today).First(&l).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}


// Delete removes a log entry by ID.
func (r *RoutineLogRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.RoutineLog{}, "id = ?", id).Error
}
