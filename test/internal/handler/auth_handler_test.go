package handler

import (
	"bytes"
	"encoding/json"
	"go-gin-api-server/internal/handler"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/pkg/apperrors"
	"go-gin-api-server/pkg/utils"
	mockServices "go-gin-api-server/test/mocks/service"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// Helper functions
func setupTestAuthHandler() (*handler.AuthHandler, *mockServices.AuthServiceMock) {
	mockAuthService := mockServices.NewAuthServiceMock()
	authHandler := handler.NewAuthHandler(mockAuthService)
	return authHandler, mockAuthService
}

func setupGinWithValidators() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		utils.RegisterCustomValidators(v)
	}

	return router
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

// Mock methods

// Testcases

func TestAuthHandler_Register(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		req := createTestRegisterRequest()
		tokenResponse := createTestTokenResponse()

		// Setup mock
		mockAuthService.On("Register", req).Return(tokenResponse, nil)

		// Setup Gin
		router := setupGinWithValidators()
		router.POST("/api/v1/auth/register", authHandler.Register)

		// Create request
		reqBody, _ := json.Marshal(req)
		httpReq, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)

		var response model.TokenResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, tokenResponse.AccessToken, response.AccessToken)
		assert.Equal(t, tokenResponse.RefreshToken, response.RefreshToken)
		assert.Equal(t, tokenResponse.TokenType, response.TokenType)
		assert.Equal(t, tokenResponse.ExpiresIn, response.ExpiresIn)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()

		// Setup Gin
		router := setupGinWithValidators()
		router.POST("/api/v1/auth/register", authHandler.Register)

		// Create request with invalid JSON
		httpReq, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString("invalid json"))
		httpReq.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockAuthService.AssertNotCalled(t, "Register")
	})

	t.Run("UserUnderAge", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		req := createTestRegisterRequest()

		// Setup mock
		mockAuthService.On("Register", req).Return(nil, apperrors.ErrUserUnderAge)

		// Setup Gin
		router := setupGinWithValidators()
		router.POST("/api/v1/auth/register", authHandler.Register)

		// Create request
		reqBody, _ := json.Marshal(req)
		httpReq, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("ReservedUsername", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		req := createTestRegisterRequest()
		req.Username = "admin" // 使用保留用戶名

		// Setup mock
		mockAuthService.On("Register", req).Return(nil, apperrors.ErrValidation)

		// Setup Gin
		router := setupGinWithValidators()
		router.POST("/api/v1/auth/register", authHandler.Register)

		// Create request
		reqBody, _ := json.Marshal(req)
		httpReq, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("UserExists", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		req := createTestRegisterRequest()

		// Setup mock
		mockAuthService.On("Register", req).Return(nil, apperrors.ErrUserExists)

		// Setup Gin
		router := setupGinWithValidators()
		router.POST("/api/v1/auth/register", authHandler.Register)

		// Create request
		reqBody, _ := json.Marshal(req)
		httpReq, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusConflict, w.Code)
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

		// Setup Gin
		router := setupGinWithValidators()
		router.POST("/api/v1/auth/login", authHandler.Login)

		// Create request
		reqBody, _ := json.Marshal(req)
		httpReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response model.TokenResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, tokenResponse.AccessToken, response.AccessToken)
		assert.Equal(t, tokenResponse.RefreshToken, response.RefreshToken)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()

		// Setup Gin
		router := setupGinWithValidators()
		router.POST("/api/v1/auth/login", authHandler.Login)

		// Create request with invalid JSON
		httpReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString("invalid json"))
		httpReq.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockAuthService.AssertNotCalled(t, "Login")
	})

	t.Run("Unauthorized", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		req := createTestLoginRequest()

		// Setup mock
		mockAuthService.On("Login", req).Return(nil, apperrors.ErrUnauthorized)

		// Setup Gin
		router := setupGinWithValidators()
		router.POST("/api/v1/auth/login", authHandler.Login)

		// Create request
		reqBody, _ := json.Marshal(req)
		httpReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockAuthService.AssertExpectations(t)
	})

	t.Run("Forbidden", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		req := createTestLoginRequest()

		// Setup mock
		mockAuthService.On("Login", req).Return(nil, apperrors.ErrForbidden)

		// Setup Gin
		router := setupGinWithValidators()
		router.POST("/api/v1/auth/login", authHandler.Login)

		// Create request
		reqBody, _ := json.Marshal(req)
		httpReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
		httpReq.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusForbidden, w.Code)

		mockAuthService.AssertExpectations(t)
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		refreshToken := "valid-refresh-token"
		tokenResponse := createTestTokenResponse()

		// Setup mock
		mockAuthService.On("RefreshToken", refreshToken).Return(tokenResponse, nil)

		// Setup Gin
		router := setupGinWithValidators()
		router.POST("/api/v1/auth/refresh", authHandler.RefreshToken)

		// Create request
		httpReq, _ := http.NewRequest("POST", "/api/v1/auth/refresh", nil)
		httpReq.AddCookie(&http.Cookie{
			Name:  "gin_api_refresh_token",
			Value: refreshToken,
		})

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response model.TokenResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, tokenResponse.AccessToken, response.AccessToken)
		assert.Equal(t, tokenResponse.RefreshToken, response.RefreshToken)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("MissingCookie", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()

		// Setup Gin
		router := setupGinWithValidators()
		router.POST("/api/v1/auth/refresh", authHandler.RefreshToken)

		// Create request without cookie
		httpReq, _ := http.NewRequest("POST", "/api/v1/auth/refresh", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		mockAuthService.AssertNotCalled(t, "RefreshToken")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		refreshToken := "invalid-refresh-token"

		// Setup mock
		mockAuthService.On("RefreshToken", refreshToken).Return(nil, apperrors.ErrInvalidToken)

		// Setup Gin
		router := setupGinWithValidators()
		router.POST("/api/v1/auth/refresh", authHandler.RefreshToken)

		// Create request
		httpReq, _ := http.NewRequest("POST", "/api/v1/auth/refresh", nil)
		httpReq.AddCookie(&http.Cookie{
			Name:  "gin_api_refresh_token",
			Value: refreshToken,
		})

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("Forbidden", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		refreshToken := "valid-but-user-inactive-token"

		// Setup mock
		mockAuthService.On("RefreshToken", refreshToken).Return(nil, apperrors.ErrForbidden)

		// Setup Gin
		router := setupGinWithValidators()
		router.POST("/api/v1/auth/refresh", authHandler.RefreshToken)

		// Create request
		httpReq, _ := http.NewRequest("POST", "/api/v1/auth/refresh", nil)
		httpReq.AddCookie(&http.Cookie{
			Name:  "gin_api_refresh_token",
			Value: refreshToken,
		})

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusForbidden, w.Code)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		authHandler, mockAuthService := setupTestAuthHandler()
		refreshToken := "expired-refresh-token"

		// Setup mock
		mockAuthService.On("RefreshToken", refreshToken).Return(nil, apperrors.ErrExpiredToken)

		// Setup Gin
		router := setupGinWithValidators()
		router.POST("/api/v1/auth/refresh", authHandler.RefreshToken)

		// Create request
		httpReq, _ := http.NewRequest("POST", "/api/v1/auth/refresh", nil)
		httpReq.AddCookie(&http.Cookie{
			Name:  "gin_api_refresh_token",
			Value: refreshToken,
		})

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, httpReq)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockAuthService.AssertExpectations(t)
	})
}
