package repository

import (
	"go-gin-api-server/internal/database"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/pkg/apperrors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *model.User) (*model.User, error)
	FindByID(id string) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Update(id string, user *model.User) (*model.User, error)
	Delete(id string) error
}

type userRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepositoryImpl{
		db: database.GetDB(),
	}
}

// NewUserRepositoryWithDB 創建使用指定資料庫連接的 Repository
func NewUserRepositoryWithDB(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{
		db: db,
	}
}

func (r *userRepositoryImpl) Create(user *model.User) (*model.User, error) {
	// check if username or email already exists
	var existingUser model.User
	if err := r.db.Where("username = ? OR email = ?", user.Username, user.Email).First(&existingUser).Error; err == nil {
		return nil, apperrors.ErrUserExists
	}

	// generate UUID as UserID
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = user.CreatedAt
	}

	// create user in database
	if err := r.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepositoryImpl) FindByID(id string) (*model.User, error) {
	var user model.User

	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, apperrors.ErrNotFound
	}

	return &user, nil
}

func (r *userRepositoryImpl) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, apperrors.ErrNotFound
	}

	return &user, nil
}

func (r *userRepositoryImpl) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, apperrors.ErrNotFound
	}

	return &user, nil
}

func (r *userRepositoryImpl) Update(id string, updated *model.User) (*model.User, error) {
	result := r.db.Model(&model.User{}).Where("id = ?", id).Updates(updated)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.ErrNotFound
	}

	var user model.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, apperrors.ErrNotFound
	}

	return &user, nil
}

func (r *userRepositoryImpl) Delete(id string) error {
	result := r.db.Where("id = ?", id).Delete(&model.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
