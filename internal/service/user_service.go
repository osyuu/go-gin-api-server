package service

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/pkg/apperrors"
	"time"
)

type UserService interface {
	CreateUser(name, username, email string, birthDate *time.Time) (*model.User, error)
	GetUserByID(id string) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	GetUserByEmail(email string) (*model.User, error)
	UpdateUserProfile(userID string, name string, birthDate *time.Time) (*model.User, error)
	ActivateUser(userID string) error
	DeactivateUser(userID string) error
	DeleteUser(userID string) error
}

type userServiceImpl struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userServiceImpl{
		repo: repo,
	}
}

func (s *userServiceImpl) GetUserByID(id string) (*model.User, error) {
	return s.repo.FindByID(id)
}

func (s *userServiceImpl) GetUserByUsername(username string) (*model.User, error) {
	return s.repo.FindByUsername(username)
}

func (s *userServiceImpl) GetUserByEmail(email string) (*model.User, error) {
	return s.repo.FindByEmail(email)
}

func (s *userServiceImpl) CreateUser(name, username, email string, birthDate *time.Time) (*model.User, error) {
	user := model.CreateUser(name, username, email, birthDate)

	// 業務邏輯驗證：年齡限制（13歲以上）
	if birthDate != nil {
		if s.isUnder13(*birthDate) {
			return nil, apperrors.ErrUserUnderAge
		}
	}

	// 業務邏輯驗證：保留用戶名檢查
	if s.isReservedUsername(username) {
		return nil, apperrors.ErrValidation // 可以定義更具體的錯誤
	}

	return s.repo.Create(user)
}

func (s *userServiceImpl) UpdateUserProfile(userID string, name string, birthDate *time.Time) (*model.User, error) {
	update := &model.User{
		Name:      name,
		BirthDate: birthDate,
	}

	// 業務邏輯驗證：如果更新生日，檢查年齡限制
	if birthDate != nil {
		if s.isUnder13(*birthDate) {
			return nil, apperrors.ErrUserUnderAge
		}
	}

	return s.repo.Update(userID, update)
}

func (s *userServiceImpl) ActivateUser(userID string) error {
	update := &model.User{IsActive: true}
	_, err := s.repo.Update(userID, update)
	return err
}

func (s *userServiceImpl) DeactivateUser(userID string) error {
	update := &model.User{IsActive: false}
	_, err := s.repo.Update(userID, update)
	return err
}

func (s *userServiceImpl) DeleteUser(userID string) error {
	return s.repo.Delete(userID)
}

// 業務邏輯驗證輔助方法

// isUnder13 檢查用戶是否未滿13歲
func (s *userServiceImpl) isUnder13(birthDate time.Time) bool {
	now := time.Now()
	age := now.Year() - birthDate.Year()

	// 如果還沒到生日，年齡減1
	if now.YearDay() < birthDate.YearDay() {
		age--
	}

	return age < 13
}

// isReservedUsername 檢查用戶名是否為保留字
func (s *userServiceImpl) isReservedUsername(username string) bool {
	reservedUsernames := []string{
		"admin", "administrator", "root", "system", "api",
		"www", "mail", "ftp", "support", "help",
		"test", "demo", "guest", "user", "default",
	}

	for _, reserved := range reservedUsernames {
		if username == reserved {
			return true
		}
	}

	return false
}
