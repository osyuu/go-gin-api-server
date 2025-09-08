package service

import (
	"go-gin-api-server/internal/model"
	"time"

	"github.com/stretchr/testify/mock"
)

type UserServiceMock struct {
	mock.Mock
}

func NewUserServiceMock() *UserServiceMock {
	return &UserServiceMock{}
}

func (m *UserServiceMock) CreateUser(name, username, email string, birthDate *time.Time) (*model.User, error) {
	args := m.Called(name, username, email, birthDate)
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

func (m *UserServiceMock) GetUserByID(id string) (*model.User, error) {
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

func (m *UserServiceMock) GetUserByUsername(username string) (*model.User, error) {
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

func (m *UserServiceMock) GetUserByEmail(email string) (*model.User, error) {
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

func (m *UserServiceMock) UpdateUserProfile(userID string, name string, birthDate *time.Time) (*model.User, error) {
	args := m.Called(userID, name, birthDate)
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

func (m *UserServiceMock) ActivateUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *UserServiceMock) DeactivateUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *UserServiceMock) DeleteUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}
