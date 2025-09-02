package repository

import (
	"blog_server/internal/model"
	"fmt"
	"sync"
)

type UserRepository interface {
	CreateUser(user *model.User) error
	GetUserById(id string) (*model.User, error)
	// UpdateUser(user *model.User) error
	// DeleteUser(id string) error
}

type userRepositoryImpl struct {
	mutex sync.Mutex
	users []model.User
}

func NewUserRepository() UserRepository {
	return &userRepositoryImpl{}
}

func (r *userRepositoryImpl) GetUserById(id string) (*model.User, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, user := range r.users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, nil
}

func (r *userRepositoryImpl) CreateUser(user *model.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Generate ID if not provided
	if user.ID == "" {
		user.ID = fmt.Sprintf("%d", len(r.users)+1)
	}

	r.users = append(r.users, *user)
	return nil
}
