package service

import (
	"blog_server/internal/model"
	"blog_server/internal/service"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserRepository struct {
	mock.Mock
}

func NewMockUserRepository() *mockUserRepository {
	return &mockUserRepository{}
}

func (m *mockUserRepository) GetUserById(id string) (*model.User, error) {
	args := m.Called(id)
	if user := args.Get(0); user != nil {
		return user.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepository) CreateUser(user *model.User) error {
	args := m.Called(user)

	if args.Error(0) == nil && user.ID == "" {
		user.ID = "1"
	}

	return args.Error(0)
}

func TestGetUserById_Success(t *testing.T) {
	repo := NewMockUserRepository()
	expected := &model.User{
		ID: "1", Name: "Mock User", Age: 20,
	}
	repo.On("GetUserById", mock.Anything).Return(expected, nil)

	userService := service.NewUserService(repo)

	user, err := userService.GetUserById("1")

	assert.NoError(t, err)
	assert.Equal(t, "Mock User", user.Name)
	assert.Equal(t, 20, user.Age)
}

func TestGetUserById_NotFound(t *testing.T) {
	repo := NewMockUserRepository()
	repo.On("GetUserById", mock.Anything).Return(nil, nil)

	userService := service.NewUserService(repo)

	user, err := userService.GetUserById("999")

	assert.Nil(t, user)
	assert.EqualError(t, err, "user not found")
}

func TestCreateUser_Success(t *testing.T) {
	repo := NewMockUserRepository()
	repo.On("CreateUser", mock.Anything).Return(nil)

	userService := service.NewUserService(repo)

	user := &model.User{ID: "1", Name: "Mock User", Age: 20}
	err := userService.CreateUser(user)
	assert.NoError(t, err)
}

func TestCreateUser_Error(t *testing.T) {
	repo := NewMockUserRepository()
	repo.On("CreateUser", mock.Anything).Return(errors.New("error"))

	userService := service.NewUserService(repo)

	user := &model.User{ID: "1", Name: "Mock User", Age: 20}
	err := userService.CreateUser(user)
	assert.Error(t, err)
}
