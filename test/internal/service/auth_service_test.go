package service

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/apperrors"
	"go-gin-api-server/pkg/utils"
	mockRepository "go-gin-api-server/test/mocks/repository"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test constants
const (
	testUserName1 = "User 1"
	testUserName2 = "User 2"
	testUserID1   = "user1-id"
)

const (
	testUserID      = "test-user-id"
	testOtherUserID = "test-other-user-id"
	testAdminUserID = "admin-user-id-6734"
)

// Helper functions
func setupTestAuthService() (*mockRepository.UserRepositoryMock, *mockRepository.AuthRepositoryMock, *utils.JWTManager, service.AuthService) {
	mockUserRepo := mockRepository.NewUserRepositoryMock()
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
		mockUserRepo.On("Create", mock.AnythingOfType("*model.User")).Return(&model.User{ID: testUserID}, nil)
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

		userID := testUserID
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

		userID := testUserID
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

		userID := testUserID
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

		userID := testUserID
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

	t.Run("MissingToken", func(t *testing.T) {
		_, _, _, authService := setupTestAuthService()

		// run
		result, err := authService.RefreshToken("")

		// assert
		assert.ErrorIs(t, err, apperrors.ErrUnauthorized)
		assert.Nil(t, result)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockUserRepo, _, jwtMgr, authService := setupTestAuthService()
		userID := testUserID
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
		userID := testUserID
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
		userID := testUserID
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

func TestAuthService_ActivateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockUserRepo, _, _, authService := setupTestAuthService()
		userID := testUserID
		user := &model.User{ID: userID, IsActive: false}
		updatedUser := &model.User{ID: userID, IsActive: true}

		// Setup mocks
		mockUserRepo.On("FindByID", mock.Anything).Return(user, nil)
		mockUserRepo.On("Update", mock.Anything, mock.Anything).Return(updatedUser, nil)

		// run
		result, err := authService.ActivateUser(userID, testAdminUserID)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, user.ID, result.ID)
		assert.True(t, result.IsActive)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("NotAdmin_Forbidden", func(t *testing.T) {
		_, _, _, authService := setupTestAuthService()
		userID := testUserID

		// run
		result, err := authService.ActivateUser(userID, userID)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Nil(t, result)
	})
}

func TestAuthService_DeactivateUser(t *testing.T) {
	t.Run("Success_Admin", func(t *testing.T) {
		mockUserRepo, _, _, authService := setupTestAuthService()
		userID := testUserID
		user := &model.User{ID: userID, IsActive: true}
		updatedUser := &model.User{ID: userID, IsActive: false}

		// Setup mocks
		mockUserRepo.On("FindByID", mock.Anything).Return(user, nil)
		mockUserRepo.On("Update", mock.Anything, mock.Anything).Return(updatedUser, nil)

		// run
		result, err := authService.DeactivateUser(userID, testAdminUserID)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, user.ID, result.ID)
		assert.False(t, result.IsActive)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Success_CurrentUser", func(t *testing.T) {
		mockUserRepo, _, _, authService := setupTestAuthService()
		userID := testUserID
		user := &model.User{ID: userID, IsActive: true}
		updatedUser := &model.User{ID: userID, IsActive: false}

		// Setup mocks
		mockUserRepo.On("FindByID", mock.Anything).Return(user, nil)
		mockUserRepo.On("Update", mock.Anything, mock.Anything).Return(updatedUser, nil)

		// run
		result, err := authService.DeactivateUser(userID, userID)

		// assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, user.ID, result.ID)
		assert.False(t, result.IsActive)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Forbidden", func(t *testing.T) {
		_, _, _, authService := setupTestAuthService()
		userID := testUserID

		// run
		result, err := authService.DeactivateUser(userID, testOtherUserID)

		// assert
		assert.ErrorIs(t, err, apperrors.ErrForbidden)
		assert.Nil(t, result)
	})
}

// TestAuthService_ConcurrentRegistration 測試併發註冊場景
func TestAuthService_ConcurrentRegistration(t *testing.T) {
	t.Run("ConcurrentRegistrationWithSameUsername", func(t *testing.T) {
		// 測試多個用戶同時註冊相同 username 的情況
		mockUserRepo, mockAuthRepo, _, authService := setupTestAuthService()

		// 準備測試數據
		birthDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		req1 := &model.RegisterRequest{
			Name:      "User 1",
			Username:  "testuser",
			Email:     "user1@example.com",
			BirthDate: &birthDate,
			Password:  "password123",
		}
		req2 := &model.RegisterRequest{
			Name:      "User 2",
			Username:  "testuser", // 相同的 username
			Email:     "user2@example.com",
			BirthDate: &birthDate,
			Password:  "password123",
		}

		// 設置 mock：第一個請求成功，第二個請求失敗（用戶已存在）
		user1 := &model.User{
			ID:       "user1-id",
			Name:     "User 1",
			Username: &req1.Username,
			Email:    &req1.Email,
			IsActive: true,
		}

		// 第一個請求的 mock 設置
		mockUserRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
			return u.Name == testUserName1
		})).Return(user1, nil).Once()

		mockAuthRepo.On("CreateCredentials", mock.MatchedBy(func(c *model.UserCredentials) bool {
			return c.UserID == testUserID1
		})).Return(&model.UserCredentials{
			ID:       "cred1-id",
			UserID:   testUserID1,
			Password: "hashed_password",
		}, nil).Once()

		// 第二個請求的 mock 設置（用戶已存在）
		mockUserRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
			return u.Name == testUserName2
		})).Return(nil, apperrors.ErrUserExists).Once()

		// 併發執行
		var wg sync.WaitGroup
		results := make([]*model.TokenResponse, 2)
		errors := make([]error, 2)

		wg.Add(2)
		go func() {
			defer wg.Done()
			results[0], errors[0] = authService.Register(req1)
		}()
		go func() {
			defer wg.Done()
			results[1], errors[1] = authService.Register(req2)
		}()

		wg.Wait()

		// 驗證結果
		// 第一個請求應該成功
		assert.NoError(t, errors[0])
		assert.NotNil(t, results[0])
		assert.NotEmpty(t, results[0].AccessToken)
		assert.NotEmpty(t, results[0].RefreshToken)

		// 第二個請求應該失敗（用戶已存在）
		assert.ErrorIs(t, errors[1], apperrors.ErrUserExists)
		assert.Nil(t, results[1])

		// 驗證 mock 調用
		mockUserRepo.AssertExpectations(t)
		mockAuthRepo.AssertExpectations(t)
	})

	t.Run("ConcurrentRegistrationWithSameEmail", func(t *testing.T) {
		// 測試多個用戶同時註冊相同 email 的情況
		mockUserRepo, mockAuthRepo, _, authService := setupTestAuthService()

		birthDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		req1 := &model.RegisterRequest{
			Name:      "User 1",
			Username:  "user1",
			Email:     "test@example.com", // 相同的 email
			BirthDate: &birthDate,
			Password:  "password123",
		}
		req2 := &model.RegisterRequest{
			Name:      "User 2",
			Username:  "user2",
			Email:     "test@example.com", // 相同的 email
			BirthDate: &birthDate,
			Password:  "password123",
		}

		user1 := &model.User{
			ID:       "user1-id",
			Name:     "User 1",
			Username: &req1.Username,
			Email:    &req1.Email,
			IsActive: true,
		}

		// 第一個請求成功
		mockUserRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
			return u.Name == "User 1"
		})).Return(user1, nil).Once()

		mockAuthRepo.On("CreateCredentials", mock.MatchedBy(func(c *model.UserCredentials) bool {
			return c.UserID == "user1-id"
		})).Return(&model.UserCredentials{
			ID:       "cred1-id",
			UserID:   "user1-id",
			Password: "hashed_password",
		}, nil).Once()

		// 第二個請求失敗（email 已存在）
		mockUserRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
			return u.Name == "User 2"
		})).Return(nil, apperrors.ErrUserExists).Once()

		// 併發執行
		var wg sync.WaitGroup
		results := make([]*model.TokenResponse, 2)
		errors := make([]error, 2)

		wg.Add(2)
		go func() {
			defer wg.Done()
			results[0], errors[0] = authService.Register(req1)
		}()
		go func() {
			defer wg.Done()
			results[1], errors[1] = authService.Register(req2)
		}()

		wg.Wait()

		// 驗證結果
		assert.NoError(t, errors[0])
		assert.NotNil(t, results[0])
		assert.ErrorIs(t, errors[1], apperrors.ErrUserExists)
		assert.Nil(t, results[1])

		mockUserRepo.AssertExpectations(t)
		mockAuthRepo.AssertExpectations(t)
	})

	t.Run("ConcurrentRegistrationWithDifferentUsers", func(t *testing.T) {
		// 測試多個不同用戶同時註冊（應該都成功）
		mockUserRepo, mockAuthRepo, _, authService := setupTestAuthService()

		birthDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		req1 := &model.RegisterRequest{
			Name:      "User 1",
			Username:  "user1",
			Email:     "user1@example.com",
			BirthDate: &birthDate,
			Password:  "password123",
		}
		req2 := &model.RegisterRequest{
			Name:      "User 2",
			Username:  "user2",
			Email:     "user2@example.com",
			BirthDate: &birthDate,
			Password:  "password123",
		}

		user1 := &model.User{
			ID:       "user1-id",
			Name:     "User 1",
			Username: &req1.Username,
			Email:    &req1.Email,
			IsActive: true,
		}
		user2 := &model.User{
			ID:       "user2-id",
			Name:     "User 2",
			Username: &req2.Username,
			Email:    &req2.Email,
			IsActive: true,
		}

		// 兩個請求都成功
		mockUserRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
			return u.Name == "User 1"
		})).Return(user1, nil).Once()
		mockUserRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
			return u.Name == "User 2"
		})).Return(user2, nil).Once()

		mockAuthRepo.On("CreateCredentials", mock.MatchedBy(func(c *model.UserCredentials) bool {
			return c.UserID == "user1-id"
		})).Return(&model.UserCredentials{
			ID:       "cred1-id",
			UserID:   "user1-id",
			Password: "hashed_password",
		}, nil).Once()
		mockAuthRepo.On("CreateCredentials", mock.MatchedBy(func(c *model.UserCredentials) bool {
			return c.UserID == "user2-id"
		})).Return(&model.UserCredentials{
			ID:       "cred2-id",
			UserID:   "user2-id",
			Password: "hashed_password",
		}, nil).Once()

		// 併發執行
		var wg sync.WaitGroup
		results := make([]*model.TokenResponse, 2)
		errors := make([]error, 2)

		wg.Add(2)
		go func() {
			defer wg.Done()
			results[0], errors[0] = authService.Register(req1)
		}()
		go func() {
			defer wg.Done()
			results[1], errors[1] = authService.Register(req2)
		}()

		wg.Wait()

		// 驗證結果：兩個請求都應該成功
		assert.NoError(t, errors[0])
		assert.NotNil(t, results[0])
		assert.NoError(t, errors[1])
		assert.NotNil(t, results[1])

		// 驗證生成的 token 不同
		assert.NotEqual(t, results[0].AccessToken, results[1].AccessToken)

		mockUserRepo.AssertExpectations(t)
		mockAuthRepo.AssertExpectations(t)
	})
}
