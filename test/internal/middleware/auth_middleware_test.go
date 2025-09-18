package middleware

import (
	"go-gin-api-server/internal/middleware"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/pkg/apperrors"
	mockServices "go-gin-api-server/test/mocks/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

const (
	expiredTokenValue = "expired-token"
)

func setupTestAuthMiddleware() (*middleware.AuthMiddleware, *mockServices.AuthServiceMock) {
	mockAuthService := mockServices.NewAuthServiceMock()
	authMiddleware := middleware.NewAuthMiddleware(mockAuthService, zap.NewNop())
	return authMiddleware, mockAuthService
}

func setupTestAuthRouter(middlewareFunc gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(middlewareFunc)

	// Add a test route that requires auth
	router.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"user_id": userID,
		})
	})

	return router
}

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	t.Run("ValidToken", func(t *testing.T) {
		authMiddleware, mockAuthService := setupTestAuthMiddleware()
		validToken := "valid-access-token"
		claims := &model.Claims{UserID: "user-123"}

		// Setup mock
		mockAuthService.On("ValidateToken", validToken).Return(claims, nil)

		// Setup router
		router := setupTestAuthRouter(authMiddleware.RequireAuth())

		// Create request
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user-123")

		mockAuthService.AssertExpectations(t)
	})

	t.Run("MissingAuthorizationHeader", func(t *testing.T) {
		authMiddleware, mockAuthService := setupTestAuthMiddleware()

		// Setup router
		router := setupTestAuthRouter(authMiddleware.RequireAuth())

		// Create request without Authorization header
		req, _ := http.NewRequest("GET", "/protected", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Unauthorized")

		mockAuthService.AssertNotCalled(t, "ValidateToken")
	})

	t.Run("InvalidAuthorizationFormat", func(t *testing.T) {
		authMiddleware, mockAuthService := setupTestAuthMiddleware()

		// Setup router
		router := setupTestAuthRouter(authMiddleware.RequireAuth())

		// Create request with invalid Authorization format
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat token")

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Unauthorized")

		mockAuthService.AssertNotCalled(t, "ValidateToken")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		authMiddleware, mockAuthService := setupTestAuthMiddleware()
		invalidToken := "invalid-token"

		// Setup mock
		mockAuthService.On("ValidateToken", invalidToken).Return(nil, apperrors.ErrInvalidToken)

		// Setup router
		router := setupTestAuthRouter(authMiddleware.RequireAuth())

		// Create request
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+invalidToken)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")

		mockAuthService.AssertExpectations(t)
	})

	t.Run("ExpiredTokenWithAutoRefresh", func(t *testing.T) {
		authMiddleware, mockAuthService := setupTestAuthMiddleware()
		expiredToken := expiredTokenValue
		refreshToken := "valid-refresh-token"
		newAccessToken := "new-access-token"
		claims := &model.Claims{UserID: "user-123"}

		// Setup mocks
		mockAuthService.On("ValidateToken", expiredToken).Return(nil, apperrors.ErrExpiredToken)
		mockAuthService.On("RefreshAccessToken", refreshToken).Return(newAccessToken, nil)
		mockAuthService.On("ValidateToken", newAccessToken).Return(claims, nil)

		// Setup router
		router := setupTestAuthRouter(authMiddleware.RequireAuth())

		// Create request
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		req.AddCookie(&http.Cookie{
			Name:  "gin_api_refresh_token",
			Value: refreshToken,
		})

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, newAccessToken, w.Header().Get("X-New-Access-Token"))
		assert.Equal(t, "Bearer", w.Header().Get("X-Token-Type"))
		assert.Contains(t, w.Body.String(), "user-123")

		mockAuthService.AssertExpectations(t)
	})

	t.Run("ExpiredTokenWithoutRefreshToken", func(t *testing.T) {
		authMiddleware, mockAuthService := setupTestAuthMiddleware()
		expiredToken := expiredTokenValue

		// Setup mock
		mockAuthService.On("ValidateToken", expiredToken).Return(nil, apperrors.ErrExpiredToken)

		// Setup router
		router := setupTestAuthRouter(authMiddleware.RequireAuth())

		// Create request without refresh token cookie
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Token has expired")

		mockAuthService.AssertExpectations(t)
	})

	t.Run("ExpiredTokenWithInvalidRefreshToken", func(t *testing.T) {
		authMiddleware, mockAuthService := setupTestAuthMiddleware()
		expiredToken := expiredTokenValue
		invalidRefreshToken := "invalid-refresh-token"

		// Setup mocks
		mockAuthService.On("ValidateToken", expiredToken).Return(nil, apperrors.ErrExpiredToken)
		mockAuthService.On("RefreshAccessToken", invalidRefreshToken).Return("", apperrors.ErrInvalidToken)

		// Setup router
		router := setupTestAuthRouter(authMiddleware.RequireAuth())

		// Create request
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		req.AddCookie(&http.Cookie{
			Name:  "gin_api_refresh_token",
			Value: invalidRefreshToken,
		})

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Token has expired")

		mockAuthService.AssertExpectations(t)
	})
}

func TestAuthMiddleware_OptionalAuth(t *testing.T) {
	t.Run("ValidToken", func(t *testing.T) {
		authMiddleware, mockAuthService := setupTestAuthMiddleware()
		validToken := "valid-access-token"
		claims := &model.Claims{UserID: "user-123"}

		// Setup mock
		mockAuthService.On("ValidateToken", validToken).Return(claims, nil)

		// Setup router
		router := setupTestAuthRouter(authMiddleware.OptionalAuth())

		// Create request
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user-123")

		mockAuthService.AssertExpectations(t)
	})

	t.Run("NoAuthorizationHeader", func(t *testing.T) {
		authMiddleware, mockAuthService := setupTestAuthMiddleware()

		// Setup router
		router := setupTestAuthRouter(authMiddleware.OptionalAuth())

		// Create request without Authorization header
		req, _ := http.NewRequest("GET", "/protected", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "null") // user_id should be null

		mockAuthService.AssertNotCalled(t, "ValidateToken")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		authMiddleware, mockAuthService := setupTestAuthMiddleware()
		invalidToken := "invalid-token"

		// Setup mock
		mockAuthService.On("ValidateToken", invalidToken).Return(nil, apperrors.ErrInvalidToken)

		// Setup router
		router := setupTestAuthRouter(authMiddleware.OptionalAuth())

		// Create request
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+invalidToken)

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "null") // user_id should be null

		mockAuthService.AssertExpectations(t)
	})

	t.Run("ExpiredTokenWithAutoRefresh", func(t *testing.T) {
		authMiddleware, mockAuthService := setupTestAuthMiddleware()
		expiredToken := expiredTokenValue
		refreshToken := "valid-refresh-token"
		newAccessToken := "new-access-token"
		claims := &model.Claims{UserID: "user-123"}

		// Setup mocks
		mockAuthService.On("ValidateToken", expiredToken).Return(nil, apperrors.ErrExpiredToken)
		mockAuthService.On("RefreshAccessToken", refreshToken).Return(newAccessToken, nil)
		mockAuthService.On("ValidateToken", newAccessToken).Return(claims, nil)

		// Setup router
		router := setupTestAuthRouter(authMiddleware.OptionalAuth())

		// Create request
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		req.AddCookie(&http.Cookie{
			Name:  "gin_api_refresh_token",
			Value: refreshToken,
		})

		// Create response recorder
		w := httptest.NewRecorder()

		// Perform request
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, newAccessToken, w.Header().Get("X-New-Access-Token"))
		assert.Contains(t, w.Body.String(), "user-123")

		mockAuthService.AssertExpectations(t)
	})
}
