package handler

import (
	"blog_server/internal/handler"
	"blog_server/internal/model"
	"blog_server/pkg/utils"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

func (m *mockUserService) GetUserById(id string) (*model.User, error) {
	args := m.Called(id)
	if user := args.Get(0); user != nil {
		return user.(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserService) CreateUser(user *model.User) error {
	args := m.Called(user)

	// Simulate ID generation if no error
	if args.Error(0) == nil && user.ID == "" {
		user.ID = "1"
	}

	return args.Error(0)
}

func setupUserRouter(handlerFunc gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Register custom validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("custom_validator", utils.CustomValidator)
	}

	r.GET("users/:id", handlerFunc)
	r.POST("users", handlerFunc)
	return r
}

func TestGetUser_Success(t *testing.T) {

	mockService := NewMockUserService()
	mockService.On("GetUserById", mock.Anything).Return(&model.User{ID: "1", Name: "Test User", Age: 20}, nil)
	userHandler := handler.NewUserHandler(mockService)
	r := setupUserRouter(userHandler.GetUser)

	// mock request
	req, _ := http.NewRequest(http.MethodGet, "/users/1", nil)

	// mock response
	response := httptest.NewRecorder()

	// run request
	r.ServeHTTP(response, req)

	// assert response
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"id": "1", "name": "Test User", "age": 20}`, response.Body.String())
}

func TestGetUser_NotFound(t *testing.T) {
	mockUserService := NewMockUserService()
	mockUserService.On("GetUserById", mock.Anything).Return(nil, errors.New("user not found"))

	userHandler := handler.NewUserHandler(mockUserService)
	r := setupUserRouter(userHandler.GetUser)

	// mock request
	req, _ := http.NewRequest(http.MethodGet, "/users/999", nil)

	// mock response
	response := httptest.NewRecorder()

	// run request
	r.ServeHTTP(response, req)

	// assert response
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.JSONEq(t, `{"error": "User not found"}`, response.Body.String())
}

func TestCreateUser_Success(t *testing.T) {
	mockUserService := NewMockUserService()
	mockUserService.On("CreateUser", mock.Anything).Return(nil)
	userHandler := handler.NewUserHandler(mockUserService)
	r := setupUserRouter(userHandler.CreateUser)

	// mock request with JSON body
	jsonBody := `{"name": "Test User", "age": 20}`
	req, _ := http.NewRequest(http.MethodPost, "/users", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// mock response
	response := httptest.NewRecorder()

	// run request
	r.ServeHTTP(response, req)

	// assert response
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.JSONEq(t, `{"id": "1", "name": "Test User", "age": 20}`, response.Body.String())
}

func TestCreateUser_Error(t *testing.T) {
	mockUserService := NewMockUserService()
	mockUserService.On("CreateUser", mock.Anything).Return(errors.New("database error"))
	userHandler := handler.NewUserHandler(mockUserService)
	r := setupUserRouter(userHandler.CreateUser)

	// mock request with JSON body
	jsonBody := `{"name": "Test User", "age": 20}`
	req, _ := http.NewRequest(http.MethodPost, "/users", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// mock response
	response := httptest.NewRecorder()

	// run request
	r.ServeHTTP(response, req)

	// assert response
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"error": "database error"}`, response.Body.String())
}
