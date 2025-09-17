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

// Test constants
const (
	testName            = "Test User"
	testUsername        = "test_user"
	testEmail           = "test_user@test.com"
	testUserID          = "test-id"
	NonExistentUserID   = "550e8400-e29b-41d4-a716-446655440000"
	NonExistentUsername = "non-existent-username"
	NonExistentEmail    = "non-existent-email@test.com"
)

// Helper functions

func setupTestUserHandler() (*mockService.UserServiceMock, *handler.UserHandler) {
	mockService := mockService.NewUserServiceMock()
	userHandler := handler.NewUserHandler(mockService, zap.NewNop())
	return mockService, userHandler
}

func setupUserRouter(handlerFunc gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Register custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		utils.RegisterCustomValidators(v)
	}

	r.GET("/users/:id", handlerFunc)
	r.GET("/users/username/:username", handlerFunc)
	r.GET("/users/email/:email", handlerFunc)
	r.GET("/users/profile/:username", handlerFunc)
	r.PATCH("/users/:id", handlerFunc)
	r.DELETE("/users/:id", handlerFunc)
	return r
}

func createTestUser(overrides ...map[string]interface{}) *model.User {
	// Default
	id := testUserID
	birthDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	name := testName
	username := testUsername
	email := testEmail
	now := time.Now().UTC().Truncate(time.Second)
	createdAt := now
	updatedAt := now

	if len(overrides) > 0 {
		override := overrides[0]
		if val, ok := override["id"]; ok {
			id = val.(string)
		}

		if val, ok := override["name"]; ok {
			name = val.(string)
		}

		if val, ok := override["birth_date"]; ok {
			birthDate = val.(time.Time)
		}

		if val, ok := override["username"]; ok {
			username = val.(string)
		}

		if val, ok := override["email"]; ok {
			email = val.(string)
		}

		if val, ok := override["created_at"]; ok {
			createdAt = val.(time.Time)
		}

		if val, ok := override["updated_at"]; ok {
			updatedAt = val.(time.Time)
		}
	}

	return &model.User{
		ID:        id,
		Name:      name,
		BirthDate: &birthDate,
		Username:  &username,
		Email:     &email,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

// TestCases

func TestGetUserByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserByID)

		expectedUser := createTestUser()

		mockService.On("GetUserByID", mock.Anything).Return(expectedUser, nil)

		req, _ := http.NewRequest(http.MethodGet, "/users/test-id", nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusOK, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserByID)

		mockService.On("GetUserByID", mock.Anything).Return(nil, apperrors.ErrNotFound)

		req, _ := http.NewRequest(http.MethodGet, "/users/"+NonExistentUserID, nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNotFound, response.Code) // ErrNotFound 映射到 404
		mockService.AssertExpectations(t)
	})

}

func TestGetUserByUsername(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserByUsername)

		expectedUser := createTestUser()

		mockService.On("GetUserByUsername", mock.Anything).Return(expectedUser, nil)

		req, _ := http.NewRequest(http.MethodGet, "/users/username/"+testUsername, nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusOK, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("BindingError_InvalidUsername", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserByUsername)

		// Test invalid username format (starts with number)
		req, _ := http.NewRequest(http.MethodGet, "/users/username/123invalid", nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusBadRequest, response.Code) // Invalid format should return 400
		mockService.AssertExpectations(t)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserByUsername)

		// Mock Service 返回 ErrNotFound
		mockService.On("GetUserByUsername", mock.Anything).Return(nil, apperrors.ErrNotFound)

		req, _ := http.NewRequest(http.MethodGet, "/users/username/"+NonExistentUsername, nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNotFound, response.Code)
		mockService.AssertExpectations(t)
	})
}

func TestGetUserByEmail(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserByEmail)

		expectedUser := createTestUser()

		mockService.On("GetUserByEmail", mock.Anything).Return(expectedUser, nil)

		req, _ := http.NewRequest(http.MethodGet, "/users/email/"+testEmail, nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusOK, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("BindingError_InvalidEmail", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserByEmail)

		// Test invalid email format
		req, _ := http.NewRequest(http.MethodGet, "/users/email/invalid-email", nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusBadRequest, response.Code) // Invalid format should return 400
		mockService.AssertExpectations(t)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserByEmail)

		// Mock Service 返回 ErrNotFound
		mockService.On("GetUserByEmail", mock.Anything).Return(nil, apperrors.ErrNotFound)

		req, _ := http.NewRequest(http.MethodGet, "/users/email/"+NonExistentEmail, nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNotFound, response.Code) // ErrNotFound 映射到 404
		mockService.AssertExpectations(t)
	})
}

func TestGetUserProfile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserProfile)

		username := testUsername
		birthDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		expectedProfile := &model.UserProfile{
			Name:      testName,
			Username:  &username,
			BirthDate: &birthDate,
		}

		mockService.On("GetUserProfile", mock.Anything).Return(expectedProfile, nil)

		req, _ := http.NewRequest(http.MethodGet, "/users/profile/"+username, nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusOK, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("BindingError_InvalidUsername", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserProfile)

		// Test invalid username format (starts with number)
		req, _ := http.NewRequest(http.MethodGet, "/users/profile/123invalid", nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusBadRequest, response.Code) // Invalid format should return 400
		mockService.AssertExpectations(t)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserProfile)

		// Mock Service 返回 ErrNotFound
		mockService.On("GetUserProfile", mock.Anything).Return(nil, apperrors.ErrNotFound)

		req, _ := http.NewRequest(http.MethodGet, "/users/profile/"+NonExistentUsername, nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNotFound, response.Code) // ErrNotFound 映射到 404
		mockService.AssertExpectations(t)
	})

}

func TestUpdateUserProfile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.UpdateUserProfile)

		createdUser := createTestUser()
		mockService.On("UpdateUserProfile",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(createdUser, nil)

		requestData := model.UpdateUserProfileRequest{
			Name: "Updated User",
		}

		req := createTypedJSONRequest(http.MethodPatch, "/users/"+testUserID, requestData)

		// run
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusOK, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.UpdateUserProfile)

		createdUser := createTestUser()
		mockService.On("UpdateUserProfile",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(createdUser, apperrors.ErrNotFound)

		requestData := model.UpdateUserProfileRequest{
			Name: "Updated User",
		}

		req := createTypedJSONRequest(http.MethodPatch, "/users/"+NonExistentUserID, requestData)

		// run
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNotFound, response.Code) // ErrNotFound 映射到 404
		mockService.AssertExpectations(t)
	})

	t.Run("NoUpdateFields", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.UpdateUserProfile)

		requestData := model.UpdateUserProfileRequest{}

		req := createTypedJSONRequest(http.MethodPatch, "/users/"+testUserID, requestData)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusBadRequest, response.Code)
		mockService.AssertExpectations(t)
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.DeleteUser)

		// Mock Service 返回成功
		mockService.On("DeleteUser", testUserID).Return(nil)

		req, _ := http.NewRequest(http.MethodDelete, "/users/"+testUserID, nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNoContent, response.Code) // 204 No Content
		mockService.AssertExpectations(t)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.DeleteUser)

		mockService.On("DeleteUser", mock.Anything).Return(apperrors.ErrNotFound)

		req, _ := http.NewRequest(http.MethodDelete, "/users/"+NonExistentUserID, nil)

		// run
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNotFound, response.Code) // ErrNotFound 映射到 404
		mockService.AssertExpectations(t)
	})
}
