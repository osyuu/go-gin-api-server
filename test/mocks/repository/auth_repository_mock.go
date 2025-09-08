package repository

import (
	"go-gin-api-server/internal/model"

	"github.com/stretchr/testify/mock"
)

type AuthRepositoryMock struct {
	mock.Mock
}

func NewAuthRepositoryMock() *AuthRepositoryMock {
	return &AuthRepositoryMock{}
}

func (m *AuthRepositoryMock) CreateCredentials(credentials *model.UserCredentials) (*model.UserCredentials, error) {
	args := m.Called(credentials)
	if user := args.Get(0); user != nil {
		userResult, ok := user.(*model.UserCredentials)
		if !ok {
			return nil, args.Error(1)
		}
		err := args.Error(1)
		return userResult, err
	}
	err := args.Error(1)
	return nil, err
}

func (m *AuthRepositoryMock) FindByUserID(userID string) (*model.UserCredentials, error) {
	args := m.Called(userID)
	if user := args.Get(0); user != nil {
		userResult, ok := user.(*model.UserCredentials)
		if !ok {
			return nil, args.Error(1)
		}
		err := args.Error(1)
		return userResult, err
	}
	err := args.Error(1)
	return nil, err
}

func (m *AuthRepositoryMock) UpdatePassword(userID string, hashedPassword string) error {
	args := m.Called(userID, hashedPassword)
	return args.Error(0)
}

func (m *AuthRepositoryMock) DeleteCredentials(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}
