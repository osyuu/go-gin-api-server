package integration

import (
	"bytes"
	"encoding/json"
	"go-gin-api-server/config"
	"go-gin-api-server/internal/handler"
	"go-gin-api-server/internal/middleware"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/logger"
	"go-gin-api-server/pkg/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func setupIntegrationAuthRouter() *gin.Engine {
	// Initialize logger
	logger.Init("test")

	// Load config
	cfg := config.LoadConfig()

	// Setup dependencies
	userRepo := repository.NewUserRepository()
	authRepo := repository.NewAuthRepository()
	jwtMgr := utils.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessTokenExpiration)
	authService := service.NewAuthService(userRepo, authRepo, jwtMgr)

	// Setup handlers and middleware
	authHandler := handler.NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register custom validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		utils.RegisterCustomValidators(v)
	}

	// Auth routes
	authHandler.RegisterRoutes(router)

	// Setup user handler for protected routes testing
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)
	userHandler.RegisterRoutes(router)
	userHandler.RegisterProtectedRoutes(router, authMiddleware)

	return router
}

func TestAuthIntegration_RegisterAndLogin(t *testing.T) {
	router := setupIntegrationAuthRouter()

	// Test data
	birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	registerReq := &model.RegisterRequest{
		Name:      "Test User",
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password123",
		BirthDate: &birthDate,
	}

	t.Run("Register", func(t *testing.T) {
		reqBody, _ := json.Marshal(registerReq)
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")
	})

	t.Run("Login", func(t *testing.T) {
		loginReq := &model.LoginRequest{
			Username: "testuser",
			Password: "password123",
		}

		reqBody, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")
	})
}

func TestAuthIntegration_ProtectedRoute(t *testing.T) {
	router := setupIntegrationAuthRouter()

	// First register a user
	birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	registerReq := &model.RegisterRequest{
		Name:      "Test User",
		Username:  "testuser2",
		Email:     "test2@example.com",
		Password:  "password123",
		BirthDate: &birthDate,
	}

	reqBody, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var registerResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &registerResponse)
	assert.NoError(t, err)

	accessToken := registerResponse["access_token"].(string)

	t.Run("AccessProtectedRouteWithValidToken", func(t *testing.T) {
		// 使用實際的用戶路由來測試受保護的路由
		req, _ := http.NewRequest("GET", "/api/v1/users/username/testuser2", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "id")
		assert.Contains(t, response, "username")
	})

	t.Run("AccessProtectedRouteWithoutToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/users/username/testuser2", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("AccessProtectedRouteWithInvalidToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/users/username/testuser2", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthIntegration_RefreshToken(t *testing.T) {
	router := setupIntegrationAuthRouter()

	// First register a user
	birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	registerReq := &model.RegisterRequest{
		Name:      "Test User",
		Username:  "testuser4",
		Email:     "test4@example.com",
		Password:  "password123",
		BirthDate: &birthDate,
	}

	reqBody, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var registerResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &registerResponse)
	assert.NoError(t, err)

	refreshToken := registerResponse["refresh_token"].(string)

	t.Run("RefreshTokenWithValidRefreshToken", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", nil)
		req.AddCookie(&http.Cookie{
			Name:  "gin_api_refresh_token",
			Value: refreshToken,
		})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")
	})

	t.Run("RefreshTokenWithoutCookie", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
