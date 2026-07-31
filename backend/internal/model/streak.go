package model

import (
	"time"

	"github.com/google/uuid"
)

// Streak tracks the current and longest completion streak for a single routine.
// One row per routine (enforced by the unique index on routine_id in migration
// 000005). LastCompleted is nullable — a freshly created streak has no last
// completion date yet.
type Streak struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null"                               json:"user_id"`
	RoutineID     uuid.UUID  `gorm:"type:uuid;not null"                               json:"routine_id"`
	CurrentStreak int        `gorm:"not null;default:0"                               json:"current_streak"`
	LongestStreak int        `gorm:"not null;default:0"                               json:"longest_streak"`
	LastCompleted *time.Time `gorm:"type:date"                                        json:"last_completed"`
	UpdatedAt     time.Time  `                                                        json:"updated_at"`
}
