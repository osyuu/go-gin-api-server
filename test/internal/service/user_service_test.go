package service

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/apperrors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserRepository struct {
	mock.Mock
}

func NewMockUserRepository() *mockUserRepository {
	return &mockUserRepository{}
}

// Helper functions

func setupTestUserService() (*mockUserRepository, service.UserService) {
	mockRepo := NewMockUserRepository()
	mockService := service.NewUserService(mockRepo)
	return mockRepo, mockService
}

func createTestUser() *model.User {
	birthDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	return model.CreateUser(
		"Mock User",
		"mock_user",
		"mock_user@test.com",
		&birthDate,
	)
}

// Mock methods

func (m *mockUserRepository) Create(user *model.User) (*model.User, error) {
	args := m.Called(user)
	if user := args.Get(0); user != nil {
		return user.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepository) FindByID(id string) (*model.User, error) {
	args := m.Called(id)
	if user := args.Get(0); user != nil {
		return user.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepository) FindByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if user := args.Get(0); user != nil {
		return user.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepository) FindByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if user := args.Get(0); user != nil {
		return user.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepository) Update(id string, user *model.User) (*model.User, error) {
	args := m.Called(id, user)
	if u := args.Get(0); u != nil {
		return u.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// Testcases

func TestCreateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, service := setupTestUserService()
		expected := createTestUser()
		repo.On("Create", mock.Anything).Return(expected, nil)

		// run
		created, err := service.CreateUser(
			expected.Name,
			expected.Username,
			expected.Email,
			expected.BirthDate,
		)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, expected, created)
		repo.AssertExpectations(t)
	})

	t.Run("ErrorUnderAge", func(t *testing.T) {
		repo, service := setupTestUserService()
		expected := createTestUser()
		underAgeBirthDate := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC) // under age
		expected.BirthDate = &underAgeBirthDate

		// run
		created, err := service.CreateUser(
			expected.Name,
			expected.Username,
			expected.Email,
			expected.BirthDate,
		)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrUserUnderAge)
		assert.Nil(t, created)
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("ErrorReservedUsername", func(t *testing.T) {
		repo, service := setupTestUserService()
		expected := createTestUser()
		expected.Username = "administrator" // reserved username

		// run
		created, err := service.CreateUser(
			expected.Name,
			expected.Username,
			expected.Email,
			expected.BirthDate,
		)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrValidation)
		assert.Nil(t, created)
		repo.AssertNotCalled(t, "Create")
	})
}

func TestUpdateUserProfile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mockService := setupTestUserService()
		created := createTestUser()
		birthDate := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
		expected := &model.User{
			Name:      "updated",
			Username:  created.Username,
			Email:     created.Email,
			BirthDate: &birthDate,
			IsActive:  created.IsActive,
			CreatedAt: created.CreatedAt,
			UpdatedAt: time.Now(),
		}

		repo.On("Update", mock.Anything, mock.Anything).Return(expected, nil)

		// run
		updated, err := mockService.UpdateUserProfile(created.ID, "updated", &birthDate)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, updated.ID)
		assert.Equal(t, expected.Name, updated.Name)
		assert.Equal(t, expected.BirthDate, updated.BirthDate)
		assert.Equal(t, expected.Username, updated.Username)
		assert.Equal(t, expected.Email, updated.Email)
		assert.Equal(t, expected.CreatedAt, updated.CreatedAt)
		assert.True(t, updated.UpdatedAt.After(updated.CreatedAt))
		repo.AssertExpectations(t)
	})

	t.Run("ErrorUnderAge", func(t *testing.T) {
		repo, mockService := setupTestUserService()
		underAgeBirthDate := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC) // under age

		// run
		created, err := mockService.UpdateUserProfile(
			"1",
			"", // 不更新 name
			&underAgeBirthDate,
		)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrUserUnderAge)
		assert.Nil(t, created)
		repo.AssertNotCalled(t, "Update")
	})
}

func TestGetUserByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mockService := setupTestUserService()
		expected := createTestUser()
		repo.On("FindByID", expected.ID).Return(expected, nil)

		user, err := mockService.GetUserByID(expected.ID)

		assert.NoError(t, err)
		assert.Equal(t, expected, user)
		repo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		repo, mockService := setupTestUserService()
		repo.On("FindByID", "nonexistent").Return(nil, apperrors.ErrNotFound)

		user, err := mockService.GetUserByID("nonexistent")

		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		assert.Nil(t, user)
		repo.AssertExpectations(t)
	})
}

func TestGetUserByUsername(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mockService := setupTestUserService()
		expected := createTestUser()
		repo.On("FindByUsername", expected.Username).Return(expected, nil)

		user, err := mockService.GetUserByUsername(expected.Username)

		assert.NoError(t, err)
		assert.Equal(t, expected, user)
		repo.AssertExpectations(t)
	})
}

func TestGetUserByEmail(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mockService := setupTestUserService()
		expected := createTestUser()
		repo.On("FindByEmail", expected.Email).Return(expected, nil)

		user, err := mockService.GetUserByEmail(expected.Email)

		assert.NoError(t, err)
		assert.Equal(t, expected, user)
		repo.AssertExpectations(t)
	})
}
