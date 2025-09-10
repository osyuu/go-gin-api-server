package service

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/apperrors"
	mockRepository "go-gin-api-server/test/mocks/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper functions

func setupTestUserService() (*mockRepository.UserRepositoryMock, service.UserService) {
	mockRepo := mockRepository.NewUserRepositoryMock()
	mockService := service.NewUserService(mockRepo)
	return mockRepo, mockService
}

func createTestUser() *model.User {
	birthDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	username := "mock_user"
	email := "mock_user@test.com"
	return model.CreateUser(
		"Mock User",
		&username,
		&email,
		&birthDate,
	)
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
		reservedUsername := "administrator" // reserved username
		expected.Username = &reservedUsername

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
			UpdatedAt: created.CreatedAt.Add(time.Second), // 確保 UpdatedAt 晚於 CreatedAt
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
		repo.On("FindByUsername", *expected.Username).Return(expected, nil)

		user, err := mockService.GetUserByUsername(*expected.Username)

		assert.NoError(t, err)
		assert.Equal(t, expected, user)
		repo.AssertExpectations(t)
	})
}

func TestGetUserByEmail(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo, mockService := setupTestUserService()
		expected := createTestUser()
		repo.On("FindByEmail", *expected.Email).Return(expected, nil)

		user, err := mockService.GetUserByEmail(*expected.Email)

		assert.NoError(t, err)
		assert.Equal(t, expected, user)
		repo.AssertExpectations(t)
	})
}
