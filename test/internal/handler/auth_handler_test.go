package handler

import (
	"go-gin-api-server/internal/handler"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/pkg/apperrors"
	"go-gin-api-server/pkg/utils"
	mockService "go-gin-api-server/test/mocks/service"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Helper functions
func setupTestAuthHandler() (*handler.AuthHandler, *mockService.AuthServiceMock) {
	mockAuthService := mockService.NewAuthServiceMock()
	authHandler := handler.NewAuthHandler(mockAuthService, zap.NewNop())
	return authHandler, mockAuthService
}

func setupAuthRouter(authHandler *handler.AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("user_role", model.RoleUser)
		c.Set("user_id", testUserID)
		c.Next()
	})

	// Register custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		utils.RegisterCustomValidators(v)
	}

	r.POST("/api/v1/auth/register", authHandler.Register)
	r.POST("/api/v1/auth/login", authHandler.Login)
	r.POST("/api/v1/auth/refresh", authHandler.RefreshToken)
	r.POST("/api/v1/auth/users/:id/activate", authHandler.ActivateUser)
	r.POST("/api/v1/auth/users/:id/deactivate", authHandler.DeactivateUser)

	return r
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

func createTestTokenResponse() *model.TokenResponse {
	return &model.TokenResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}
}

// Testcases

func TestAuthHandler_Register(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		registerReq := createTestRegisterRequest()
		tokenResponse := createTestTokenResponse()

		// Setup mock
		mockAuthService.On("Register", registerReq).Return(tokenResponse, nil)

		// Setup router
		router := setupAuthRouter(authHandler)

		// Create json request
		httpReq := createTypedJSONRequest(http.MethodPost, "/api/v1/auth/register", registerReq)

		// run
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("ServerError", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		registerReq := createTestRegisterRequest()

		// Setup mock
		mockAuthService.On("Register", registerReq).Return(nil, apperrors.ErrUserUnderAge)

		// Setup router
		router := setupAuthRouter(authHandler)

		// Create json request
		httpReq := createTypedJSONRequest(http.MethodPost, "/api/v1/auth/register", registerReq)

		// run
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httpReq)

		// Create request
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockAuthService.AssertExpectations(t)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		req := createTestLoginRequest()
		tokenResponse := createTestTokenResponse()

		// Setup mock
		mockAuthService.On("Login", req).Return(tokenResponse, nil)

		// Setup router
		router := setupAuthRouter(authHandler)

		// Create json request
		httpReq := createTypedJSONRequest(http.MethodPost, "/api/v1/auth/login", req)

		// run
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("ServerError", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		req := createTestLoginRequest()
		tokenResponse := createTestTokenResponse()

		// Setup mock
		mockAuthService.On("Login", req).Return(tokenResponse, apperrors.ErrUnauthorized)

		// Setup router
		router := setupAuthRouter(authHandler)

		// Create json request
		httpReq := createTypedJSONRequest(http.MethodPost, "/api/v1/auth/login", req)

		// run
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockAuthService.AssertExpectations(t)
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		tokenResponse := createTestTokenResponse()

		// Setup mock - 不管傳入什麼參數都成功
		mockAuthService.On("RefreshToken", mock.Anything).Return(tokenResponse, nil)

		// Setup router
		router := setupAuthRouter(authHandler)

		// Create request
		httpReq := createTypedJSONRequest(http.MethodPost, "/api/v1/auth/refresh", nil)

		// run
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("ServerError", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()

		// Setup mock - Service 層驗證失敗
		mockAuthService.On("RefreshToken", mock.Anything).Return(nil, apperrors.ErrUnauthorized)

		// Setup router
		router := setupAuthRouter(authHandler)

		// Create request
		httpReq := createTypedJSONRequest(http.MethodPost, "/api/v1/auth/refresh", nil)

		// run
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockAuthService.AssertExpectations(t)
	})
}

func TestAuthHandler_ActivateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		userID := testUserID
		updatedUser := &model.User{ID: userID, IsActive: true}

		// Setup mock
		mockAuthService.On("ActivateUser", mock.Anything, mock.Anything).Return(updatedUser, nil)

		// Setup router
		router := setupAuthRouter(authHandler)

		// Create json request
		req := createTypedJSONRequest(http.MethodPost, "/api/v1/auth/users/"+userID+"/activate", nil)

		// run
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("ServerError", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()

		// Setup mock
		mockAuthService.On("ActivateUser", mock.Anything, mock.Anything).Return(nil, apperrors.ErrForbidden)

		// Setup router
		router := setupAuthRouter(authHandler)

		// Create json request
		req := createTypedJSONRequest(http.MethodPost, "/api/v1/auth/users/"+testUserID+"/activate", nil)
		// run
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, w.Code)
		mockAuthService.AssertExpectations(t)
	})
}

func TestAuthHandler_DeactivateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		userID := testUserID
		updatedUser := &model.User{ID: userID, IsActive: false}

		// Setup mock
		mockAuthService.On(
			"DeactivateUser",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(updatedUser, nil)

		// Setup router
		router := setupAuthRouter(authHandler)

		// Create json request
		req := createTypedJSONRequest(http.MethodPost, "/api/v1/auth/users/"+userID+"/deactivate", nil)

		// run
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("ServerError", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()

		// Setup mock
		mockAuthService.On(
			"DeactivateUser",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(nil, apperrors.ErrForbidden)

		// Setup router
		router := setupAuthRouter(authHandler)

		// Create json request
		req := createTypedJSONRequest(http.MethodPost, "/api/v1/auth/users/"+testUserID+"/deactivate", nil)

		// run
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusForbidden, w.Code)
		mockAuthService.AssertExpectations(t)
	})
}
