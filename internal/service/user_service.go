package service

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/pkg/apperrors"
	"time"
)

type UserService interface {
	CreateUser(name string, username, email *string, birthDate *time.Time) (*model.User, error)
	GetUserByID(id string) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	GetUserByEmail(email string) (*model.User, error)
	GetUserProfile(username string) (*model.UserProfile, error)
	UpdateUserProfile(userID string, req model.UpdateUserProfileRequest) (*model.User, error)

	// Admin operations
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

func (s *userServiceImpl) GetUserProfile(username string) (*model.UserProfile, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	return &model.UserProfile{
		Name:      user.Name,
		Username:  user.Username,
		BirthDate: user.BirthDate,
	}, nil
}

func (s *userServiceImpl) CreateUser(name string, username, email *string, birthDate *time.Time) (*model.User, error) {
	user := model.CreateUser(name, username, email, birthDate)

	// business logic validation: check if the user is under 13
	if birthDate != nil {
		if s.isUnder13(*birthDate) {
			return nil, apperrors.ErrUserUnderAge
		}
	}

	// business logic validation: check if the username is reserved
	if username != nil && s.isReservedUsername(*username) {
		return nil, apperrors.ErrValidation // 可以定義更具體的錯誤
	}

	return s.repo.Create(user)
}

func (s *userServiceImpl) UpdateUserProfile(userID string, req model.UpdateUserProfileRequest) (*model.User, error) {
	// business logic validation: if updating birth date, check if the user is under 13
	if req.BirthDate != nil {
		if s.isUnder13(*req.BirthDate) {
			return nil, apperrors.ErrUserUnderAge
		}
	}

	update := &model.User{
		Name:      req.Name,
		BirthDate: req.BirthDate,
	}

	return s.repo.Update(userID, update)
}

func (s *userServiceImpl) DeleteUser(userID string) error {
	return s.repo.Delete(userID)
}

// business logic validation helper methods

// check if the user is under 13
func (s *userServiceImpl) isUnder13(birthDate time.Time) bool {
	now := time.Now().UTC().Truncate(time.Microsecond)
	age := now.Year() - birthDate.Year()

	// if the user's birthday hasn't passed yet, subtract 1 from the age
	if now.YearDay() < birthDate.YearDay() {
		age--
	}

	return age < 13
}

// check if the username is reserved
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
