package handler

import (
	"errors"
	"go-gin-api-server/internal/handler"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/repository"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPostService struct {
	mock.Mock
}

func NewMockPostService() *mockPostService {
	return &mockPostService{}
}

func setupPostRouter(handlerFunc gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.GET("/posts", handlerFunc)
	r.GET("/posts/:id", handlerFunc)
	r.POST("/posts", handlerFunc)
	r.PATCH("/posts/:id", handlerFunc)
	r.DELETE("/posts/:id", handlerFunc)
	return r
}

// Mock methods

func (m *mockPostService) Create(post *model.Post) (*model.Post, error) {
	args := m.Called(post)
	if p := args.Get(0); p != nil {
		return p.(*model.Post), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPostService) GetAll() ([]model.Post, error) {
	args := m.Called()
	if list := args.Get(0); list != nil {
		return list.([]model.Post), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPostService) GetByID(id string) (*model.Post, error) {
	args := m.Called(id)
	if p := args.Get(0); p != nil {
		return p.(*model.Post), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPostService) Update(id string, post *model.Post) (*model.Post, error) {
	args := m.Called(id, post)
	if p := args.Get(0); p != nil {
		return p.(*model.Post), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockPostService) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// Testcases

func TestGetPostByID_Success(t *testing.T) {
	mockService := NewMockPostService()
	mockService.On("GetByID", "1").Return(&model.Post{ID: "1", Title: "Test Post", Content: "Test Content", Author: "Test Author"}, nil)
	handler := handler.NewPostHandler(mockService)
	r := setupPostRouter(handler.GetPostByID)

	// run
	req, _ := http.NewRequest(http.MethodGet, "/posts/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)

	// assert
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"id": "1", "title": "Test Post", "content": "Test Content", "author": "Test Author"}`, response.Body.String())
	mockService.AssertExpectations(t)
}

func TestGetPostByID_NotFound(t *testing.T) {
	mockService := NewMockPostService()
	mockService.On("GetByID", "999").Return(nil, repository.ErrNotFound)
	handler := handler.NewPostHandler(mockService)
	r := setupPostRouter(handler.GetPostByID)

	// run
	req, _ := http.NewRequest(http.MethodGet, "/posts/999", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)

	// assert
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.JSONEq(t, `{"error": "Post not found"}`, response.Body.String())
	mockService.AssertExpectations(t)
}

func TestGetPosts_Success(t *testing.T) {
	mockService := NewMockPostService()
	mockService.On("GetAll").Return([]model.Post{
		{ID: "1", Title: "Test Post", Content: "Test Content", Author: "Test Author"},
		{ID: "2", Title: "Test Post 2", Content: "Test Content 2", Author: "Test Author 2"},
	}, nil)
	handler := handler.NewPostHandler(mockService)
	r := setupPostRouter(handler.GetPosts)

	// run
	req, _ := http.NewRequest(http.MethodGet, "/posts", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)

	// assert
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `[
		{"id": "1", "title": "Test Post", "content": "Test Content", "author": "Test Author"},
		{"id": "2", "title": "Test Post 2", "content": "Test Content 2", "author": "Test Author 2"}
	]`, response.Body.String())
	mockService.AssertExpectations(t)
}

func TestGetPosts_Error(t *testing.T) {
	mockService := NewMockPostService()
	mockService.On("GetAll").Return(nil, errors.New("database error"))
	handler := handler.NewPostHandler(mockService)
	r := setupPostRouter(handler.GetPosts)

	// run
	req, _ := http.NewRequest(http.MethodGet, "/posts", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)

	// assert
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"error": "Failed to retrieve posts"}`, response.Body.String())
	mockService.AssertExpectations(t)
}

func TestCreatePost_Success(t *testing.T) {
	mockService := NewMockPostService()
	mockService.On("Create", mock.Anything).Return(&model.Post{ID: "1", Title: "Test Post", Content: "Test Content", Author: "Test Author"}, nil)
	handler := handler.NewPostHandler(mockService)
	r := setupPostRouter(handler.CreatePost)

	// mock request with JSON body
	jsonBody := `{"title": "Title"}`
	req, _ := http.NewRequest(http.MethodPost, "/posts", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// run
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)

	// assert
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.JSONEq(t, `{"id": "1", "title": "Test Post", "content": "Test Content", "author": "Test Author"}`, response.Body.String())
	mockService.AssertExpectations(t)
}

func TestCreatePost_Error(t *testing.T) {
	mockService := NewMockPostService()
	mockService.On("Create", mock.Anything).Return(nil, errors.New("database error"))
	handler := handler.NewPostHandler(mockService)
	r := setupPostRouter(handler.CreatePost)

	// mock request with JSON body
	jsonBody := `{"title": "Title"}`
	req, _ := http.NewRequest(http.MethodPost, "/posts", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// run
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)

	// assert
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"error": "Failed to create post"}`, response.Body.String())
	mockService.AssertExpectations(t)
}

func TestUpdatePost_Success(t *testing.T) {
	mockService := NewMockPostService()
	mockService.On("Update", "1", mock.Anything).Return(&model.Post{ID: "1", Title: "Updated Title", Content: "Updated Content", Author: "Updated Author"}, nil)
	handler := handler.NewPostHandler(mockService)
	r := setupPostRouter(handler.UpdatePost)

	// mock request with JSON body
	jsonBody := `{"title": "Title"}`
	req, _ := http.NewRequest(http.MethodPatch, "/posts/1", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// run
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)

	// assert
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"id": "1", "title": "Updated Title", "content": "Updated Content", "author": "Updated Author"}`, response.Body.String())
	mockService.AssertExpectations(t)
}

func TestUpdatePost_NotFound(t *testing.T) {
	mockService := NewMockPostService()
	mockService.On("Update", "1", mock.Anything).Return(nil, repository.ErrNotFound)
	handler := handler.NewPostHandler(mockService)
	r := setupPostRouter(handler.UpdatePost)

	// mock request with JSON body
	jsonBody := `{"title": "Title"}`
	req, _ := http.NewRequest(http.MethodPatch, "/posts/1", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// run
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)

	// assert
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.JSONEq(t, `{"error": "Post not found"}`, response.Body.String())
	mockService.AssertExpectations(t)
}

func TestUpdatePost_Error(t *testing.T) {
	mockService := NewMockPostService()
	mockService.On("Update", "1", mock.Anything).Return(nil, errors.New("database error"))
	handler := handler.NewPostHandler(mockService)
	r := setupPostRouter(handler.UpdatePost)

	// mock request with JSON body
	jsonBody := `{"title": "Title"}`
	req, _ := http.NewRequest(http.MethodPatch, "/posts/1", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// run
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)

	// assert
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"error": "Failed to update post"}`, response.Body.String())
	mockService.AssertExpectations(t)
}

func TestDeletePost_Success(t *testing.T) {
	mockService := NewMockPostService()
	mockService.On("Delete", "1").Return(nil)
	handler := handler.NewPostHandler(mockService)
	r := setupPostRouter(handler.DeletePost)

	// run
	req, _ := http.NewRequest(http.MethodDelete, "/posts/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)

	// assert
	assert.Equal(t, http.StatusNoContent, response.Code)
	mockService.AssertExpectations(t)
}

func TestDeletePost_NotFound(t *testing.T) {
	mockService := NewMockPostService()
	mockService.On("Delete", "1").Return(repository.ErrNotFound)
	handler := handler.NewPostHandler(mockService)
	r := setupPostRouter(handler.DeletePost)

	// run
	req, _ := http.NewRequest(http.MethodDelete, "/posts/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)

	// assert
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.JSONEq(t, `{"error": "Post not found"}`, response.Body.String())
	mockService.AssertExpectations(t)
}

func TestDeletePost_Error(t *testing.T) {
	mockService := NewMockPostService()
	mockService.On("Delete", "1").Return(errors.New("database error"))
	handler := handler.NewPostHandler(mockService)
	r := setupPostRouter(handler.DeletePost)

	// run
	req, _ := http.NewRequest(http.MethodDelete, "/posts/1", nil)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)

	// assert
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{"error": "Failed to delete post"}`, response.Body.String())
	mockService.AssertExpectations(t)
}
