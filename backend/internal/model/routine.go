package model

import (
	"time"

	"github.com/google/uuid"
)

type Routine struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null"                               json:"user_id"`
	Title       string    `gorm:"type:varchar(100);not null"                       json:"title"`
	Description string    `gorm:"type:text"                                        json:"description"`
	PeriodType  string    `gorm:"type:period_type;not null"                        json:"period_type"`
	TargetCount int       `gorm:"not null;default:1"                               json:"target_count"`
	Icon        string    `gorm:"type:varchar(50)"                                 json:"icon"`
	Color       string    `gorm:"type:varchar(20)"                                 json:"color"`
	IsActive    bool      `gorm:"not null;default:true"                            json:"is_active"`
	SortOrder   int       `gorm:"not null;default:0"                               json:"sort_order"`
	CreatedAt   time.Time `                                                        json:"created_at"`
	UpdatedAt   time.Time `                                                        json:"updated_at"`
}
