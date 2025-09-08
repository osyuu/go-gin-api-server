package repository

import (
	"go-gin-api-server/internal/model"

	"github.com/stretchr/testify/mock"
)

type UserRepositoryMock struct {
	mock.Mock
}

func NewUserRepositoryMock() *UserRepositoryMock {
	return &UserRepositoryMock{}
}

// Mock methods

func (m *UserRepositoryMock) Create(user *model.User) (*model.User, error) {
	args := m.Called(user)
	if userResult := args.Get(0); userResult != nil {
		user, ok := userResult.(*model.User)
		if !ok {
			return nil, args.Error(1)
		}
		err := args.Error(1)
		return user, err
	}
	err := args.Error(1)
	return nil, err
}

func (m *UserRepositoryMock) FindByID(id string) (*model.User, error) {
	args := m.Called(id)
	if user := args.Get(0); user != nil {
		userResult, ok := user.(*model.User)
		if !ok {
			return nil, args.Error(1)
		}
		err := args.Error(1)
		return userResult, err
	}
	err := args.Error(1)
	return nil, err
}

func (m *UserRepositoryMock) FindByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if user := args.Get(0); user != nil {
		userResult, ok := user.(*model.User)
		if !ok {
			return nil, args.Error(1)
		}
		err := args.Error(1)
		return userResult, err
	}
	err := args.Error(1)
	return nil, err
}

func (m *UserRepositoryMock) FindByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if user := args.Get(0); user != nil {
		userResult, ok := user.(*model.User)
		if !ok {
			return nil, args.Error(1)
		}
		err := args.Error(1)
		return userResult, err
	}
	err := args.Error(1)
	return nil, err
}

func (m *UserRepositoryMock) Update(id string, user *model.User) (*model.User, error) {
	args := m.Called(id, user)
	if u := args.Get(0); u != nil {
		userResult, ok := u.(*model.User)
		if !ok {
			return nil, args.Error(1)
		}
		err := args.Error(1)
		return userResult, err
	}
	err := args.Error(1)
	return nil, err
}

func (m *UserRepositoryMock) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
