package service

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/apperrors"
	"go-gin-api-server/pkg/utils"
	mockRepository "go-gin-api-server/test/mocks/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper functions
func setupTestAuthService() (*mockUserRepository, *mockRepository.AuthRepositoryMock, *utils.JWTManager, service.AuthService) {
	mockUserRepo := NewMockUserRepository()
	mockAuthRepo := mockRepository.NewAuthRepositoryMock()
	jwtMgr := utils.NewJWTManager("test-secret", 15*time.Minute)
	authService := service.NewAuthService(mockUserRepo, mockAuthRepo, jwtMgr)
	return mockUserRepo, mockAuthRepo, jwtMgr, authService
}

func createTestRegisterRequest() *model.RegisterRequest {
	birthDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	return &model.RegisterRequest{
		Name:      "Test User",
		Username:  "testuser",
		Email:     "test@example.com",
		BirthDate: &birthDate,
		Password:  "password123",
	}
}

func createTestLoginRequest() *model.LoginRequest {
	return &model.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
}

// Testcases

func TestAuthService_Register(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockUserRepo, mockAuthRepo, _, authService := setupTestAuthService()
		req := createTestRegisterRequest()

		// Setup mocks
		mockUserRepo.On("Create", mock.AnythingOfType("*model.User")).Return(&model.User{ID: "test-user-id"}, nil)
		mockAuthRepo.On("CreateCredentials", mock.AnythingOfType("*model.UserCredentials")).Return(&model.UserCredentials{}, nil)

		// run
		result, err := authService.Register(req)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)
		assert.Equal(t, "Bearer", result.TokenType)
		assert.Greater(t, result.ExpiresIn, int64(0))

		mockUserRepo.AssertExpectations(t)
		mockAuthRepo.AssertExpectations(t)
	})

	t.Run("UserUnderAge", func(t *testing.T) {
		mockUserRepo, mockAuthRepo, _, authService := setupTestAuthService()
		birthDate := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC) // 9 years old
		req := &model.RegisterRequest{
			Name:      "Test User",
			Username:  "testuser",
			Email:     "test@example.com",
			BirthDate: &birthDate,
			Password:  "password123",
		}

		// run
		result, err := authService.Register(req)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrUserUnderAge)
		assert.Nil(t, result)

		mockUserRepo.AssertNotCalled(t, "Create")
		mockAuthRepo.AssertNotCalled(t, "CreateCredentials")
	})

	t.Run("ReservedUsername", func(t *testing.T) {
		mockUserRepo, mockAuthRepo, _, authService := setupTestAuthService()
		req := &model.RegisterRequest{
			Name:     "Test User",
			Username: "admin", // reserved username
			Email:    "test@example.com",
			Password: "password123",
		}

		// run
		result, err := authService.Register(req)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrValidation)
		assert.Nil(t, result)

		mockUserRepo.AssertNotCalled(t, "Create")
		mockAuthRepo.AssertNotCalled(t, "CreateCredentials")
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Run("SuccessWithUsername", func(t *testing.T) {
		mockUserRepo, mockAuthRepo, _, authService := setupTestAuthService()
		req := createTestLoginRequest()

		userID := "test-user-id"
		user := &model.User{ID: userID, IsActive: true}

		// generate bcrypt hash
		hashedPassword, _ := utils.HashPassword("password123")
		credentials := &model.UserCredentials{
			UserID:   userID,
			Password: hashedPassword,
		}

		// Setup mocks
		mockUserRepo.On("FindByUsername", "testuser").Return(user, nil)
		mockAuthRepo.On("FindByUserID", userID).Return(credentials, nil)

		// run
		result, err := authService.Login(req)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)

		mockUserRepo.AssertExpectations(t)
		mockAuthRepo.AssertExpectations(t)
	})

	t.Run("SuccessWithEmail", func(t *testing.T) {
		mockUserRepo, mockAuthRepo, _, authService := setupTestAuthService()
		req := &model.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		userID := "test-user-id"
		user := &model.User{ID: userID, IsActive: true}

		// 生成真實的 bcrypt hash
		hashedPassword, _ := utils.HashPassword("password123")
		credentials := &model.UserCredentials{
			UserID:   userID,
			Password: hashedPassword,
		}

		// Setup mocks
		mockUserRepo.On("FindByEmail", "test@example.com").Return(user, nil)
		mockAuthRepo.On("FindByUserID", userID).Return(credentials, nil)

		// run
		result, err := authService.Login(req)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, result)

		mockUserRepo.AssertExpectations(t)
		mockAuthRepo.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockUserRepo, mockAuthRepo, _, authService := setupTestAuthService()
		req := createTestLoginRequest()

		// Setup mocks
		mockUserRepo.On("FindByUsername", "testuser").Return(nil, apperrors.ErrNotFound)

		// run
		result, err := authService.Login(req)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
		assert.Nil(t, result)

		mockUserRepo.AssertExpectations(t)
		mockAuthRepo.AssertNotCalled(t, "FindByUserID")
	})

	t.Run("UserInactive", func(t *testing.T) {
		mockUserRepo, mockAuthRepo, _, authService := setupTestAuthService()
		req := createTestLoginRequest()

		userID := "test-user-id"
		user := &model.User{ID: userID, IsActive: false}

		// generate bcrypt hash
		hashedPassword, _ := utils.HashPassword("password123")
		credentials := &model.UserCredentials{
			UserID:   userID,
			Password: hashedPassword,
		}

		// Setup mocks
		mockUserRepo.On("FindByUsername", "testuser").Return(user, nil)
		mockAuthRepo.On("FindByUserID", userID).Return(credentials, nil)

		// run
		result, err := authService.Login(req)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Nil(t, result)

		mockUserRepo.AssertExpectations(t)
		mockAuthRepo.AssertExpectations(t)
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockUserRepo, _, jwtMgr, authService := setupTestAuthService()

		userID := "test-user-id"
		user := &model.User{ID: userID, IsActive: true}

		// Generate a valid refresh token
		tokenResponse, _ := jwtMgr.GenerateToken(user)
		refreshToken := tokenResponse.RefreshToken

		// Setup mocks
		mockUserRepo.On("FindByID", userID).Return(user, nil)

		// run
		result, err := authService.RefreshToken(refreshToken)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.AccessToken)
		assert.NotEmpty(t, result.RefreshToken)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		_, _, _, authService := setupTestAuthService()

		// run
		result, err := authService.RefreshToken("invalid-token")

		// assert
		assert.ErrorIs(t, err, apperrors.ErrInvalidToken)
		assert.Nil(t, result)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockUserRepo, _, jwtMgr, authService := setupTestAuthService()
		userID := "test-user-id"
		user := &model.User{ID: userID, IsActive: true}

		// Generate a valid refresh token
		tokenResponse, _ := jwtMgr.GenerateToken(user)
		refreshToken := tokenResponse.RefreshToken

		// Setup mocks
		mockUserRepo.On("FindByID", user.ID).Return(nil, apperrors.ErrNotFound)

		// run
		result, err := authService.RefreshToken(refreshToken)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
		assert.Nil(t, result)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("UserInactive", func(t *testing.T) {
		mockUserRepo, _, jwtMgr, authService := setupTestAuthService()
		userID := "test-user-id"
		user := &model.User{ID: userID, IsActive: false}

		// Generate a valid refresh token
		tokenResponse, _ := jwtMgr.GenerateToken(user)
		refreshToken := tokenResponse.RefreshToken

		// Setup mocks
		mockUserRepo.On("FindByID", user.ID).Return(user, nil)

		// run
		result, err := authService.RefreshToken(refreshToken)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Nil(t, result)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		_, _, jwtMgr, authService := setupTestAuthService()
		userID := "test-user-id"
		user := &model.User{ID: userID, IsActive: true}

		// Generate a valid token
		tokenResponse, _ := jwtMgr.GenerateToken(user)
		accessToken := tokenResponse.AccessToken

		// run
		claims, err := authService.ValidateToken(accessToken)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, user.ID, claims.UserID)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		_, _, _, authService := setupTestAuthService()

		// run
		claims, err := authService.ValidateToken("invalid-token")

		// assert
		assert.ErrorIs(t, err, apperrors.ErrInvalidToken)
		assert.Nil(t, claims)
	})
}
