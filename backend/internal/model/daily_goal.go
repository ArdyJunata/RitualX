package model

import (
	"time"

	"github.com/google/uuid"
)

// DailyGoal is a per-user, per-day rollup of routine completion progress.
// One row per (user_id, date) (enforced by the unique index in migration
// 000006). IsAchieved is set when Completed >= TotalRoutines for that day.
type DailyGoal struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null"                               json:"user_id"`
	Date          time.Time `gorm:"type:date;not null"                               json:"date"`
	TotalRoutines int       `gorm:"not null;default:0"                               json:"total_routines"`
	Completed     int       `gorm:"not null;default:0"                               json:"completed"`
	IsAchieved    bool      `gorm:"not null;default:false"                           json:"is_achieved"`
	CreatedAt     time.Time `                                                        json:"created_at"`
}
