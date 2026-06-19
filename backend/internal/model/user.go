package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Username     string    `gorm:"type:varchar(20);uniqueIndex;not null" json:"username"`
	DisplayName  string    `gorm:"type:varchar(100)" json:"display_name"`
	AvatarURL    string    `gorm:"type:varchar(500)" json:"avatar_url"`
	XP           int       `gorm:"not null;default:0" json:"xp"`
	Level        int       `gorm:"not null;default:1" json:"level"`
	Coins        int       `gorm:"not null;default:0" json:"coins"`
	Title        string    `gorm:"type:varchar(50);not null;default:'Novice'" json:"title"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
