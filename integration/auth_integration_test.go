package integration

import (
	"go-gin-api-server/internal/handler"
	"go-gin-api-server/internal/middleware"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/logger"
	"go-gin-api-server/pkg/utils"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupIntegrationAuthRouter(db *gorm.DB) *gin.Engine {
	// Setup dependencies
	userRepo := repository.NewUserRepositoryWithDB(db)
	authRepo := repository.NewAuthRepositoryWithDB(db)

	// Setup services
	authService := service.NewAuthService(userRepo, authRepo, globalJWTManager)

	// Setup handlers
	authHandler := handler.NewAuthHandler(authService, logger.Log)

	// Setup middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register custom validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		utils.RegisterCustomValidators(v)
	}

	// Register routes
	authHandler.RegisterRoutes(router)
	authHandler.RegisterProtectedRoutes(router, authMiddleware)

	return router
}

// validateJWTToken
func validateJWTToken(t *testing.T, tokenString string) *model.Claims {
	claims, err := globalJWTManager.ValidateToken(tokenString)
	assert.NoError(t, err, "JWT token should be valid")
	assert.NotNil(t, claims, "Claims should not be nil")

	return claims
}

func TestAuthIntegration_AuthLifecycle(t *testing.T) {
	db := setup()
	defer teardown(db)
	router := setupIntegrationAuthRouter(db)

	// 1. Register (public route)
	birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	registerReq := &model.RegisterRequest{
		Name:      "Test User",
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password123",
		BirthDate: &birthDate,
	}
	registerResp := makeHTTPRequest(t, router, "POST", "/api/v1/auth/register", registerReq, "")
	assert.Equal(t, http.StatusCreated, registerResp.Code)

	var registerResponse model.TokenResponse
	parseJSONResponse(t, registerResp, &registerResponse)

	// 驗證 token 格式和基本屬性
	assert.NotEmpty(t, registerResponse.AccessToken)
	assert.NotEmpty(t, registerResponse.RefreshToken)
	assert.Equal(t, "Bearer", registerResponse.TokenType)
	assert.Greater(t, registerResponse.ExpiresIn, int64(0))

	// 驗證 JWT token 內容正確性
	claims := validateJWTToken(t, registerResponse.AccessToken)
	assert.Equal(t, utils.JWTIssuer, claims.Issuer)
	assert.Equal(t, claims.UserID, claims.Subject) // UserID 應該等於 Subject

	// 驗證 refresh token 內容正確性
	refreshClaims := validateJWTToken(t, registerResponse.RefreshToken)
	assert.Equal(t, utils.JWTIssuer, refreshClaims.Issuer)
	assert.Equal(t, refreshClaims.UserID, refreshClaims.Subject)

	// 2. Login (public route)
	loginReq := &model.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	loginResp := makeHTTPRequest(t, router, "POST", "/api/v1/auth/login", loginReq, "")
	assert.Equal(t, http.StatusOK, loginResp.Code)

	var loginResponse model.TokenResponse
	parseJSONResponse(t, loginResp, &loginResponse)

	// 驗證 login token 格式和基本屬性
	assert.NotEmpty(t, loginResponse.AccessToken)
	assert.NotEmpty(t, loginResponse.RefreshToken)
	assert.Equal(t, "Bearer", loginResponse.TokenType)
	assert.Greater(t, loginResponse.ExpiresIn, int64(0))

	// 驗證 login JWT token 內容正確性
	loginClaims := validateJWTToken(t, loginResponse.AccessToken)
	assert.Equal(t, utils.JWTIssuer, loginClaims.Issuer)
	assert.Equal(t, loginClaims.UserID, loginClaims.Subject)

	// 驗證 login 和 register 的 UserID 一致
	assert.Equal(t, claims.UserID, loginClaims.UserID, "Login and register should have same user ID")

	// 3. Refresh Token (public route)
	refreshResp := makeHTTPRequestWithCookie(t, router, "POST", "/api/v1/auth/refresh", nil, loginResponse.RefreshToken)
	assert.Equal(t, http.StatusOK, refreshResp.Code)

	var refreshResponse model.TokenResponse
	parseJSONResponse(t, refreshResp, &refreshResponse)

	// 驗證 refresh token 格式和基本屬性
	assert.NotEmpty(t, refreshResponse.AccessToken)
	assert.NotEmpty(t, refreshResponse.RefreshToken)
	assert.Equal(t, "Bearer", refreshResponse.TokenType)
	assert.Greater(t, refreshResponse.ExpiresIn, int64(0))

	// 驗證 refresh JWT token 內容正確性
	refreshTokenClaims := validateJWTToken(t, refreshResponse.AccessToken)
	assert.Equal(t, utils.JWTIssuer, refreshTokenClaims.Issuer)
	assert.Equal(t, refreshTokenClaims.UserID, refreshTokenClaims.Subject)

	// 驗證 refresh 後 UserID 仍然一致
	assert.Equal(t, claims.UserID, refreshTokenClaims.UserID, "Refresh should maintain same user ID")

	// 4. Activate User (protected route - admin only)
	// First test normal user cannot activate
	activateResp := makeHTTPRequest(t, router, "POST", "/api/v1/auth/users/"+claims.UserID+"/activate", nil, loginResponse.AccessToken)
	assert.Equal(t, http.StatusForbidden, activateResp.Code)

	// Then test admin can activate
	adminUser := &model.User{ID: AdminUserID, IsActive: true}
	adminToken, _ := globalJWTManager.GenerateAccessToken(adminUser)

	activateResp = makeHTTPRequest(t, router, "POST", "/api/v1/auth/users/"+claims.UserID+"/activate", nil, adminToken)
	assert.Equal(t, http.StatusOK, activateResp.Code)

	// 5. Deactivate User (protected route - self or admin)
	deactivateResp := makeHTTPRequest(t, router, "POST", "/api/v1/auth/users/"+claims.UserID+"/deactivate", nil, loginResponse.AccessToken)
	assert.Equal(t, http.StatusOK, deactivateResp.Code)
}

func TestAuthIntegration_Unauthorized(t *testing.T) {
	db := setup()
	defer teardown(db)
	router := setupIntegrationAuthRouter(db)

	// Create test user for testing protected routes
	user := createTestUser(t, db)

	t.Run("ActivateUser_Unauthorized", func(t *testing.T) {
		resp := makeHTTPRequest(t, router, "POST", "/api/v1/auth/users/"+user.ID+"/activate", nil, "")
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("DeactivateUser_Unauthorized", func(t *testing.T) {
		resp := makeHTTPRequest(t, router, "POST", "/api/v1/auth/users/"+user.ID+"/deactivate", nil, "")
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("RefreshToken_Unauthorized", func(t *testing.T) {
		resp := makeHTTPRequest(t, router, "POST", "/api/v1/auth/refresh", nil, "")
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}

func TestAuthIntegration_EdgeCases(t *testing.T) {
	db := setup()
	defer teardown(db)
	router := setupIntegrationAuthRouter(db)

	t.Run("Register_DuplicateUser", func(t *testing.T) {
		birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
		registerReq := &model.RegisterRequest{
			Name:      "Test User",
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			BirthDate: &birthDate,
		}

		// First registration should succeed
		resp1 := makeHTTPRequest(t, router, "POST", "/api/v1/auth/register", registerReq, "")
		assert.Equal(t, http.StatusCreated, resp1.Code)

		// Second registration with same email should fail
		resp2 := makeHTTPRequest(t, router, "POST", "/api/v1/auth/register", registerReq, "")
		assert.Equal(t, http.StatusConflict, resp2.Code)
	})

	t.Run("RefreshToken_InvalidToken", func(t *testing.T) {
		resp := makeHTTPRequestWithCookie(t, router, "POST", "/api/v1/auth/refresh", nil, "invalid-refresh-token")
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}
