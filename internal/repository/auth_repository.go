package repository

import (
	"go-gin-api-server/internal/database"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/pkg/apperrors"

	"gorm.io/gorm"
)

type AuthRepository interface {
	CreateCredentials(credentials *model.UserCredentials) (*model.UserCredentials, error)
	FindByUserID(userID string) (*model.UserCredentials, error)
	UpdatePassword(userID string, hashedPassword string) error
	DeleteCredentials(userID string) error
}

type authRepositoryImpl struct {
	db *gorm.DB
}

func NewAuthRepository() AuthRepository {
	return &authRepositoryImpl{
		db: database.GetDB(),
	}
}

func NewAuthRepositoryWithDB(db *gorm.DB) AuthRepository {
	return &authRepositoryImpl{
		db: db,
	}
}

func (r *authRepositoryImpl) CreateCredentials(credentials *model.UserCredentials) (*model.UserCredentials, error) {
	// Validate input
	if credentials.UserID == "" {
		return nil, apperrors.ErrValidation
	}

	var existingCredentials model.UserCredentials
	if err := r.db.Where("user_id = ?", credentials.UserID).First(&existingCredentials).Error; err == nil {
		return nil, apperrors.ErrUserExists
	}

	if err := r.db.Create(credentials).Error; err != nil {
		return nil, err
	}

	return credentials, nil
}

func (r *authRepositoryImpl) FindByUserID(userID string) (*model.UserCredentials, error) {
	var credentials model.UserCredentials
	if err := r.db.Where("user_id = ?", userID).First(&credentials).Error; err != nil {
		return nil, apperrors.ErrNotFound
	}

	return &credentials, nil
}

func (r *authRepositoryImpl) UpdatePassword(userID string, hashedPassword string) error {
	result := r.db.Model(&model.UserCredentials{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{"password": hashedPassword})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *authRepositoryImpl) DeleteCredentials(userID string) error {
	result := r.db.Where("user_id = ?", userID).Delete(&model.UserCredentials{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
