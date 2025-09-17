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

func setupIntegrationUserRouter(db *gorm.DB) *gin.Engine {
	// Setup dependencies
	userRepo := repository.NewUserRepositoryWithDB(db)
	authRepo := repository.NewAuthRepositoryWithDB(db)

	// Setup services
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, authRepo, globalJWTManager)

	// Setup handlers
	userHandler := handler.NewUserHandler(userService, logger.Log)

	// Setup middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Setup router
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Register custom validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		utils.RegisterCustomValidators(v)
	}

	// Register routes
	userHandler.RegisterRoutes(r)
	userHandler.RegisterProtectedRoutes(r, authMiddleware)

	return r
}

func TestUserIntegration_UserLifecycle(t *testing.T) {
	db := setup()
	defer teardown(db)
	router := setupIntegrationUserRouter(db)

	// 1. Create user and token
	user := createTestUser(t, db)
	token := createTestToken(t, user)
	accessToken := token.AccessToken

	// 2. Get user profile (public route)
	profileResp := makeHTTPRequest(t, router, "GET", "/api/v1/users/profile/"+*user.Username, nil, "")
	assert.Equal(t, http.StatusOK, profileResp.Code)

	var profile model.UserProfile
	parseJSONResponse(t, profileResp, &profile)
	assert.Equal(t, *user.Username, *profile.Username)
	assert.Equal(t, user.Name, profile.Name)
	assert.Equal(t, user.BirthDate.UTC(), profile.BirthDate.UTC())

	// 3. Get user by ID (protected route)
	userByIDResp := makeHTTPRequest(t, router, "GET", "/api/v1/users/"+user.ID, nil, accessToken)
	assert.Equal(t, http.StatusOK, userByIDResp.Code)

	// 4. Get user by username (protected route)
	userByUsernameResp := makeHTTPRequest(t, router, "GET", "/api/v1/users/username/"+*user.Username, nil, accessToken)
	assert.Equal(t, http.StatusOK, userByUsernameResp.Code)

	// 5. Get user by email (protected route)
	userByEmailResp := makeHTTPRequest(t, router, "GET", "/api/v1/users/email/"+*user.Email, nil, accessToken)
	assert.Equal(t, http.StatusOK, userByEmailResp.Code)

	// 6. Update user profile (protected route)
	updateReq := map[string]interface{}{
		"name":       "Updated User",
		"birth_date": time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	updateResp := makeHTTPRequest(t, router, "PATCH", "/api/v1/users/"+user.ID, updateReq, accessToken)
	assert.Equal(t, http.StatusOK, updateResp.Code)

	var updatedUser model.User
	parseJSONResponse(t, updateResp, &updatedUser)
	assert.Equal(t, "Updated User", updatedUser.Name)
	assert.Equal(t, time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC).UTC(), updatedUser.BirthDate.UTC())
	assert.True(t, updatedUser.UpdatedAt.UTC().After(updatedUser.CreatedAt.UTC()))
}

func TestUserIntegration_Unauthorized(t *testing.T) {
	db := setup()
	defer teardown(db)
	router := setupIntegrationUserRouter(db)

	// Create test user for the endpoints
	user := createTestUser(t, db)

	// Test unauthorized access to all protected routes
	t.Run("GetUserByID_Unauthorized", func(t *testing.T) {
		resp := makeHTTPRequest(t, router, "GET", "/api/v1/users/"+user.ID, nil, "")
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("GetUserByUsername_Unauthorized", func(t *testing.T) {
		resp := makeHTTPRequest(t, router, "GET", "/api/v1/users/username/"+*user.Username, nil, "")
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("GetUserByEmail_Unauthorized", func(t *testing.T) {
		resp := makeHTTPRequest(t, router, "GET", "/api/v1/users/email/"+*user.Email, nil, "")
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("UpdateUserProfile_Unauthorized", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"name": "Updated User",
		}
		resp := makeHTTPRequest(t, router, "PATCH", "/api/v1/users/"+user.ID, updateReq, "")
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}
