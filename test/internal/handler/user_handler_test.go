package handler

import (
	"bytes"
	"encoding/json"
	"go-gin-api-server/internal/handler"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/pkg/apperrors"
	"go-gin-api-server/pkg/utils"
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

type mockUserService struct {
	mock.Mock
}

func NewMockUserService() *mockUserService {
	return &mockUserService{}
}

// Helper functions

func setupTestUserHandler() (*mockUserService, *handler.UserHandler) {
	mockService := NewMockUserService()
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
	return model.CreateUser(
		"Mock User",
		"mock_user",
		"mock_user@test.com",
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

// Mock methods

func (m *mockUserService) CreateUser(name, username, email string, birthDate *time.Time) (*model.User, error) {
	args := m.Called(name, username, email, birthDate)
	if user := args.Get(0); user != nil {
		return user.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserService) GetUserByID(id string) (*model.User, error) {
	args := m.Called(id)
	if user := args.Get(0); user != nil {
		return user.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserService) GetUserByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if user := args.Get(0); user != nil {
		return user.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserService) GetUserByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if user := args.Get(0); user != nil {
		return user.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserService) UpdateUserProfile(userID string, name string, birthDate *time.Time) (*model.User, error) {
	args := m.Called(userID, name, birthDate)
	if user := args.Get(0); user != nil {
		return user.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserService) ActivateUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *mockUserService) DeactivateUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *mockUserService) DeleteUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

// TestCases

func TestCreateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, userHandler := setupTestUserHandler()
		r := setupUserRouter(userHandler.CreateUser)

		// 創建期望的用戶數據
		birthDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		expectedUser := &model.User{
			ID:        "test-id",
			Name:      "Test User",
			Username:  "test_user",
			Email:     "test_user@test.com",
			BirthDate: &birthDate,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Mock 期望與實際請求參數匹配
		mockService.On("CreateUser",
			"Test User",
			"test_user",
			"test_user@test.com",
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

		mockService.On("CreateUser",
			"Test User",
			"test_user",
			"test_user@test.com",
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
		mockService.On("CreateUser",
			"Test User",
			"test_user",
			"test_user@test.com",
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
		expectedUser.ID = "test-id"

		mockService.On("GetUserByID", "test-id").Return(expectedUser, nil)

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
		mockService.On("GetUserByID", "non-existent-id").Return(nil, apperrors.ErrNotFound)

		req, _ := http.NewRequest(http.MethodGet, "/users/non-existent-id", nil)
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
		expectedUser.Username = "test-username"

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
		expectedUser.Email = "test-email"

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
		expectedUser := &model.User{
			ID:        "test-id",
			Name:      "Updated User",
			Username:  "test_user",
			Email:     "test_user@test.com",
			BirthDate: &birthDate,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("UpdateUserProfile",
			"test-id",
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
		mockService.On("UpdateUserProfile", "non-existent-id", "Updated User", mock.Anything).Return(nil, apperrors.ErrNotFound)

		requestData := handler.UpdateUserProfileRequest{
			Name: "Updated User",
		}

		req := createTypedJSONRequest(http.MethodPatch, "/users/non-existent-id", requestData)
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
		mockService.On("UpdateUserProfile", "test-id", "Updated User", &underAgeBirthDate).Return(nil, apperrors.ErrUserUnderAge)

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
		mockService.On("ActivateUser", "test-id").Return(nil)

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
		mockService.On("ActivateUser", "non-existent-id").Return(apperrors.ErrNotFound)

		req, _ := http.NewRequest(http.MethodPatch, "/users/non-existent-id/activate", nil)
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
		mockService.On("DeactivateUser", "test-id").Return(nil)

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
		mockService.On("DeactivateUser", "non-existent-id").Return(apperrors.ErrNotFound)

		req, _ := http.NewRequest(http.MethodPatch, "/users/non-existent-id/deactivate", nil)
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
		mockService.On("DeleteUser", "test-id").Return(nil)

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
		mockService.On("DeleteUser", "non-existent-id").Return(apperrors.ErrNotFound)

		req, _ := http.NewRequest(http.MethodDelete, "/users/non-existent-id", nil)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusNotFound, response.Code) // ErrNotFound 映射到 404
		mockService.AssertExpectations(t)
	})
}
