package handler

import (
	"bytes"
	"encoding/json"
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
)

// Test constants
const (
	username          = "test_user"
	email             = "test_user@test.com"
	testUserID        = "test-id"
	NonExistentUserID = "550e8400-e29b-41d4-a716-446655440000"
)

// Helper functions

func setupTestUserHandler() (*mockService.UserServiceMock, *handler.UserHandler) {
	mockService := mockService.NewUserServiceMock()
	userHandler := handler.NewUserHandler(mockService)
	return mockService, userHandler
}

func setupUserRouter(handlerFunc gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Register custom validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("username", utils.UsernameValidator)
	}

	r.GET("/users/:id", handlerFunc)
	r.GET("/users/username/:username", handlerFunc)
	r.GET("/users/email/:email", handlerFunc)
	r.POST("/users", handlerFunc)
	r.PATCH("/users/:id", handlerFunc)
	r.PATCH("/users/:id/activate", handlerFunc)
	r.PATCH("/users/:id/deactivate", handlerFunc)
	r.DELETE("/users/:id", handlerFunc)
	return r
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

// Helper function to create JSON request body
func createJSONRequest(data interface{}) *bytes.Buffer {
	jsonData, _ := json.Marshal(data)
	return bytes.NewBuffer(jsonData)
}

// Helper function to create HTTP request with JSON body
func createJSONHTTPRequest(method, url string, data interface{}) *http.Request {
	req, _ := http.NewRequest(method, url, createJSONRequest(data))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// Test request structures are now defined in user_handler.go

// Helper function to create typed request
func createTypedJSONRequest(method, url string, data interface{}) *http.Request {
	return createJSONHTTPRequest(method, url, data)
}

// TestCases

func TestCreateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.CreateUser)

		// 創建期望的用戶數據
		birthDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		username := username
		email := email
		expectedUser := model.CreateUser(
			"Test User",
			&username,
			&email,
			&birthDate,
		)
		expectedUser.ID = testUserID

		// Mock 期望與實際請求參數匹配
		mockService.On("CreateUser",
			"Test User",
			&username,
			&email,
			&birthDate,
		).Return(expectedUser, nil)

		requestData := handler.CreateUserRequest{
			Name:      "Test User",
			Username:  "test_user",
			Email:     "test_user@test.com",
			BirthDate: &birthDate,
		}

		req := createTypedJSONRequest(http.MethodPost, "/users", requestData)

		// run
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusCreated, response.Code)
		var responseUser model.User
		err := json.Unmarshal(response.Body.Bytes(), &responseUser)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, responseUser.ID)
		assert.Equal(t, expectedUser.Name, responseUser.Name)
		assert.Equal(t, expectedUser.Username, responseUser.Username)
		assert.Equal(t, expectedUser.Email, responseUser.Email)
		assert.Equal(t, expectedUser.BirthDate, responseUser.BirthDate)
		assert.Equal(t, expectedUser.IsActive, responseUser.IsActive)

		mockService.AssertExpectations(t)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.CreateUser)

		jsonBody := `{"invalid": json}`
		req := createTypedJSONRequest(http.MethodPost, "/users", jsonBody)

		// run
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusBadRequest, response.Code)
		mockService.AssertNotCalled(t, "CreateUser")
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.CreateUser)

		// missing required fields
		requestData := handler.CreateUserRequest{
			Name: "",
		}

		req := createTypedJSONRequest(http.MethodPost, "/users", requestData)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusBadRequest, response.Code)
		mockService.AssertNotCalled(t, "CreateUser")
	})

	t.Run("InvalidEmailFormat", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.CreateUser)

		// invalid email format
		requestData := handler.CreateUserRequest{
			Name:      "Test User",
			Username:  "test_user",
			Email:     "invalid-email",
			BirthDate: nil,
		}

		req := createTypedJSONRequest(http.MethodPost, "/users", requestData)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusBadRequest, response.Code)
		mockService.AssertNotCalled(t, "CreateUser")
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.CreateUser)

		birthDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		requestData := handler.CreateUserRequest{
			Name:      "Test User",
			Username:  "test_user",
			Email:     "test_user@test.com",
			BirthDate: &birthDate,
		}

		username := username
		email := email
		mockService.On("CreateUser",
			"Test User",
			&username,
			&email,
			&birthDate,
		).Return(nil, apperrors.ErrUserExists)

		req := createTypedJSONRequest(http.MethodPost, "/users", requestData)

		// run
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusConflict, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("UserUnderAge", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.CreateUser)

		// 設置一個未滿13歲的生日
		underAgeBirthDate := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC) // 假設現在是2024年
		requestData := handler.CreateUserRequest{
			Name:      "Test User",
			Username:  "test_user",
			Email:     "test_user@test.com",
			BirthDate: &underAgeBirthDate,
		}

		// Mock Service 返回 ErrUserUnderAge
		username := username
		email := email
		mockService.On("CreateUser",
			"Test User",
			&username,
			&email,
			&underAgeBirthDate,
		).Return(nil, apperrors.ErrUserUnderAge)

		req := createTypedJSONRequest(http.MethodPost, "/users", requestData)

		// run
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusBadRequest, response.Code) // ErrUserUnderAge 映射到 400
		mockService.AssertExpectations(t)
	})
}

func TestGetUserByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserByID)

		expectedUser := createTestUser()
		expectedUser.ID = testUserID

		mockService.On("GetUserByID", testUserID).Return(expectedUser, nil)

		req, _ := http.NewRequest(http.MethodGet, "/users/test-id", nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusOK, response.Code)
		var responseUser model.User
		err := json.Unmarshal(response.Body.Bytes(), &responseUser)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, responseUser.ID)
		assert.Equal(t, expectedUser.Name, responseUser.Name)
		assert.Equal(t, expectedUser.Username, responseUser.Username)
		assert.Equal(t, expectedUser.Email, responseUser.Email)

		mockService.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserByID)

		// Mock Service 返回 ErrNotFound
		mockService.On("GetUserByID", NonExistentUserID).Return(nil, apperrors.ErrNotFound)

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
		testUsername := "test-username"
		expectedUser.Username = &testUsername

		mockService.On("GetUserByUsername", "test-username").Return(expectedUser, nil)

		req, _ := http.NewRequest(http.MethodGet, "/users/username/test-username", nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusOK, response.Code)
		var responseUser model.User
		err := json.Unmarshal(response.Body.Bytes(), &responseUser)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, responseUser.ID)
		assert.Equal(t, expectedUser.Name, responseUser.Name)
		assert.Equal(t, expectedUser.Username, responseUser.Username)
		assert.Equal(t, expectedUser.Email, responseUser.Email)

		mockService.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserByUsername)

		// Mock Service 返回 ErrNotFound
		mockService.On("GetUserByUsername", "non-existent-username").Return(nil, apperrors.ErrNotFound)

		req, _ := http.NewRequest(http.MethodGet, "/users/username/non-existent-username", nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNotFound, response.Code) // ErrNotFound 映射到 404
		mockService.AssertExpectations(t)
	})
}

func TestGetUserByEmail(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserByEmail)

		expectedUser := createTestUser()
		testEmail := "test-email"
		expectedUser.Email = &testEmail

		mockService.On("GetUserByEmail", "test-email").Return(expectedUser, nil)

		req, _ := http.NewRequest(http.MethodGet, "/users/email/test-email", nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusOK, response.Code)
		var responseUser model.User
		err := json.Unmarshal(response.Body.Bytes(), &responseUser)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, responseUser.ID)
		assert.Equal(t, expectedUser.Name, responseUser.Name)
		assert.Equal(t, expectedUser.Username, responseUser.Username)
		assert.Equal(t, expectedUser.Email, responseUser.Email)

		mockService.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.GetUserByEmail)

		// Mock Service 返回 ErrNotFound
		mockService.On("GetUserByEmail", "non-existent-email").Return(nil, apperrors.ErrNotFound)

		req, _ := http.NewRequest(http.MethodGet, "/users/email/non-existent-email", nil)
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

		birthDate := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
		testUsername := username
		testEmail := email
		expectedUser := model.CreateUser(
			"Updated User",
			&testUsername,
			&testEmail,
			&birthDate,
		)
		expectedUser.ID = testUserID

		mockService.On("UpdateUserProfile",
			testUserID,
			"Updated User",
			&birthDate,
		).Return(expectedUser, nil)

		requestData := handler.UpdateUserProfileRequest{
			Name:      "Updated User",
			BirthDate: &birthDate,
		}

		req := createTypedJSONRequest(http.MethodPatch, "/users/test-id", requestData)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusOK, response.Code)
		var responseUser model.User
		err := json.Unmarshal(response.Body.Bytes(), &responseUser)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.Name, responseUser.Name)
		assert.Equal(t, expectedUser.BirthDate, responseUser.BirthDate)

		mockService.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.UpdateUserProfile)

		// Mock Service 返回 ErrNotFound
		mockService.On("UpdateUserProfile", NonExistentUserID, "Updated User", mock.Anything).Return(nil, apperrors.ErrNotFound)

		requestData := handler.UpdateUserProfileRequest{
			Name: "Updated User",
		}

		req := createTypedJSONRequest(http.MethodPatch, "/users/"+NonExistentUserID, requestData)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNotFound, response.Code) // ErrNotFound 映射到 404
		mockService.AssertExpectations(t)
	})

	t.Run("NoUpdateFields", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.UpdateUserProfile)

		// empty update request
		requestData := handler.UpdateUserProfileRequest{}

		req := createTypedJSONRequest(http.MethodPatch, "/users/test-id", requestData)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusBadRequest, response.Code)
		mockService.AssertNotCalled(t, "UpdateUserProfile")
	})

	t.Run("UserUnderAge", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.UpdateUserProfile)

		// 設置一個未滿13歲的生日
		underAgeBirthDate := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC) // 假設現在是2024年
		requestData := handler.UpdateUserProfileRequest{
			Name:      "Updated User",
			BirthDate: &underAgeBirthDate,
		}

		// Mock Service 返回 ErrUserUnderAge
		mockService.On("UpdateUserProfile", testUserID, "Updated User", &underAgeBirthDate).Return(nil, apperrors.ErrUserUnderAge)

		req := createTypedJSONRequest(http.MethodPatch, "/users/test-id", requestData)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusBadRequest, response.Code) // ErrUserUnderAge 映射到 400
		mockService.AssertExpectations(t)
	})
}

func TestActivateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.ActivateUser)

		// Mock Service 返回成功
		mockService.On("ActivateUser", testUserID).Return(nil)

		req, _ := http.NewRequest(http.MethodPatch, "/users/test-id/activate", nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNoContent, response.Code) // 204 No Content
		mockService.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.ActivateUser)

		// Mock Service 返回 ErrNotFound
		mockService.On("ActivateUser", NonExistentUserID).Return(apperrors.ErrNotFound)

		req, _ := http.NewRequest(http.MethodPatch, "/users/"+NonExistentUserID+"/activate", nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNotFound, response.Code) // ErrNotFound 映射到 404
		mockService.AssertExpectations(t)
	})
}

func TestDeactivateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.DeactivateUser)

		// Mock Service 返回成功
		mockService.On("DeactivateUser", testUserID).Return(nil)

		req, _ := http.NewRequest(http.MethodPatch, "/users/test-id/deactivate", nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNoContent, response.Code) // 204 No Content
		mockService.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.DeactivateUser)

		// Mock Service 返回 ErrNotFound
		mockService.On("DeactivateUser", NonExistentUserID).Return(apperrors.ErrNotFound)

		req, _ := http.NewRequest(http.MethodPatch, "/users/"+NonExistentUserID+"/deactivate", nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNotFound, response.Code) // ErrNotFound 映射到 404
		mockService.AssertExpectations(t)
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.DeleteUser)

		// Mock Service 返回成功
		mockService.On("DeleteUser", testUserID).Return(nil)

		req, _ := http.NewRequest(http.MethodDelete, "/users/test-id", nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNoContent, response.Code) // 204 No Content
		mockService.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.DeleteUser)

		// Mock Service 返回 ErrNotFound
		mockService.On("DeleteUser", NonExistentUserID).Return(apperrors.ErrNotFound)

		req, _ := http.NewRequest(http.MethodDelete, "/users/"+NonExistentUserID, nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNotFound, response.Code) // ErrNotFound 映射到 404
		mockService.AssertExpectations(t)
	})
}
