package repository

import (
	"github.com/ArdyJunata/RitualX/backend/internal/model"
	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(token *model.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *RefreshTokenRepository) FindByToken(tokenStr string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	if err := r.db.Where("token = ?", tokenStr).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *RefreshTokenRepository) Delete(tokenStr string) error {
	return r.db.Where("token = ?", tokenStr).Delete(&model.RefreshToken{}).Error
}
