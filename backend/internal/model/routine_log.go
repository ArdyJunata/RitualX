package model

import (
	"time"

	"github.com/google/uuid"
)

type RoutineLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	RoutineID uuid.UUID `gorm:"type:uuid;not null"                               json:"routine_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"                               json:"user_id"`
	LoggedAt  time.Time `gorm:"type:date;not null"                               json:"logged_at"`
	Count     int       `gorm:"not null;default:1"                               json:"count"`
	Note      string    `gorm:"type:text"                                        json:"note"`
	CreatedAt time.Time `                                                        json:"created_at"`
}
