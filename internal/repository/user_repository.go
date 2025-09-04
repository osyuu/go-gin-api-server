package repository

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/pkg/apperrors"
	"sync"
	"time"

	"github.com/google/uuid"
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
	mutex sync.RWMutex
	users map[string]*model.User
}

func NewUserRepository() UserRepository {
	return &userRepositoryImpl{
		users: make(map[string]*model.User),
	}
}

func (r *userRepositoryImpl) Create(user *model.User) (*model.User, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// check if username or email already exists
	for _, existingUser := range r.users {
		if existingUser.Username == user.Username {
			return nil, apperrors.ErrUserExists
		}
		if existingUser.Email == user.Email {
			return nil, apperrors.ErrUserExists
		}
	}

	// generate UUID as UserID
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = user.CreatedAt
	}

	r.users[user.ID] = user
	return user, nil
}

func (r *userRepositoryImpl) FindByID(id string) (*model.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, apperrors.ErrNotFound
	}

	return user, nil
}

func (r *userRepositoryImpl) FindByUsername(username string) (*model.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, apperrors.ErrNotFound
}

func (r *userRepositoryImpl) FindByEmail(email string) (*model.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, apperrors.ErrNotFound
}

func (r *userRepositoryImpl) Update(id string, updated *model.User) (*model.User, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	existingUser, exists := r.users[id]
	if !exists {
		return nil, apperrors.ErrNotFound
	}

	if updated.Name != "" {
		existingUser.Name = updated.Name
	}
	if updated.Email != "" {
		existingUser.Email = updated.Email
	}
	if updated.BirthDate != nil {
		existingUser.BirthDate = updated.BirthDate
	}
	if updated.Username != "" {
		existingUser.Username = updated.Username
	}
	if updated.IsActive != existingUser.IsActive {
		existingUser.IsActive = updated.IsActive
	}

	// 更新時間戳
	existingUser.UpdatedAt = time.Now()

	// 保存更新後的用戶
	r.users[id] = existingUser

	return existingUser, nil
}

func (r *userRepositoryImpl) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	_, exists := r.users[id]
	if !exists {
		return apperrors.ErrNotFound
	}

	delete(r.users, id)
	return nil
}
