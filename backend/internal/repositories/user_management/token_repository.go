package user_management

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/josy-coder/adminsuite/internal/models"

)

type TokenRepository interface {
	Create(token *models.Token) error
	FindByToken(token string) (*models.Token, error)
	FindByUserID(userID uuid.UUID) ([]*models.Token, error)
	Delete(id uuid.UUID) error
	DeleteExpired() error
}

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Create(token *models.Token) error {
	return r.db.Create(token).Error
}

func (r *tokenRepository) FindByToken(token string) (*models.Token, error) {
	var t models.Token
	err := r.db.Where("token = ?", token).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *tokenRepository) FindByUserID(userID uuid.UUID) ([]*models.Token, error) {
	var tokens []*models.Token
	err := r.db.Where("user_id = ?", userID).Find(&tokens).Error
	return tokens, err
}

func (r *tokenRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Token{}, "id = ?", id).Error
}

func (r *tokenRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.Token{}).Error
}
