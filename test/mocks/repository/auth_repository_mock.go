package repositories

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
	return args.Get(0).(*model.UserCredentials), args.Error(1)
}

func (m *AuthRepositoryMock) FindByUserID(userID string) (*model.UserCredentials, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserCredentials), args.Error(1)
}

func (m *AuthRepositoryMock) UpdatePassword(userID string, hashedPassword string) error {
	args := m.Called(userID, hashedPassword)
	return args.Error(0)
}

func (m *AuthRepositoryMock) DeleteCredentials(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}
