package handler

import (
	"go-gin-api-server/internal/handler"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/pkg/apperrors"
	mockService "go-gin-api-server/test/mocks/service"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func setupTestPostHandler() (*mockService.PostServiceMock, *handler.PostHandler) {
	mockService := mockService.NewPostServiceMock()
	postHandler := handler.NewPostHandler(mockService, zap.NewNop())
	return mockService, postHandler
}

func setupPostRouter(postHandler *handler.PostHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Set("user_role", model.RoleUser)
		c.Set("user_id", authorID)
		c.Next()
	})

	r.GET("/posts", postHandler.GetPosts)
	r.GET("/posts/:id", postHandler.GetPostByID)
	r.POST("/posts", postHandler.CreatePost)
	r.PATCH("/posts/:id", postHandler.UpdatePost)
	r.DELETE("/posts/:id", postHandler.DeletePost)
	return r
}

func createTestPost(overrides ...map[string]interface{}) *model.Post {
	// Default
	id := uint64(1)
	content := "Test Content"
	createdAt := time.Now()
	updatedAt := time.Now()
	authorID := authorID

	if len(overrides) > 0 {
		override := overrides[0]
		if val, ok := override["id"]; ok {
			id = val.(uint64)
		}

		if val, ok := override["author_id"]; ok {
			authorID = val.(string)
		}

		if val, ok := override["content"]; ok {
			content = val.(string)
		}

		if val, ok := override["created_at"]; ok {
			createdAt = val.(time.Time)
		}

		if val, ok := override["updated_at"]; ok {
			updatedAt = val.(time.Time)
		}
	}

	return &model.Post{
		ID:        id,
		Content:   content,
		AuthorID:  authorID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

var (
	authorID          = "author-e29b-41d4-a716-446655440000"
	NonExistentPostID = uint64(999)
)

// Testcases

func TestCreatePost(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, postHandler := setupTestPostHandler()
		r := setupPostRouter(postHandler)

		expected := createTestPost()
		mockService.On("Create", mock.Anything).Return(expected, nil)

		requestData := &model.Post{
			Content: "Test Content",
		}

		req := createTypedJSONRequest(http.MethodPost, "/posts", requestData)

		// run
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		// assert
		assert.Equal(t, http.StatusCreated, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("BindingError", func(t *testing.T) {
		mockService, postHandler := setupTestPostHandler()
		r := setupPostRouter(postHandler)

		requestData := &model.Post{
			Content: "", // empty content, trigger required validation
		}

		req := createTypedJSONRequest(http.MethodPost, "/posts", requestData)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		assert.Equal(t, http.StatusBadRequest, response.Code)
		mockService.AssertNotCalled(t, "Create")
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService, postHandler := setupTestPostHandler()
		r := setupPostRouter(postHandler)

		// Mock service return error
		mockService.On("Create", mock.Anything).Return(nil, apperrors.ErrPostContentTooLong)

		requestData := &model.Post{
			Content: "This is a valid content that should pass validation",
		}

		req := createTypedJSONRequest(http.MethodPost, "/posts", requestData)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		assert.Equal(t, http.StatusBadRequest, response.Code)
		mockService.AssertExpectations(t)
	})

}

func TestUpdatePost(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, postHandler := setupTestPostHandler()
		r := setupPostRouter(postHandler)

		expected := createTestPost()
		mockService.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(expected, nil)

		requestData := &model.Post{
			Content: "Updated Content",
		}

		req := createTypedJSONRequest(http.MethodPatch, "/posts/1", requestData)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("BindingError_InvalidID", func(t *testing.T) {
		mockService, postHandler := setupTestPostHandler()
		r := setupPostRouter(postHandler)

		// invalid id parameter
		req := createTypedJSONRequest(http.MethodPatch, "/posts/invalid", nil)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		assert.Equal(t, http.StatusBadRequest, response.Code)
		mockService.AssertNotCalled(t, "Update")
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService, postHandler := setupTestPostHandler()
		r := setupPostRouter(postHandler)

		mockService.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil, apperrors.ErrNotFound)

		requestData := &model.Post{
			Content: "Updated Content",
		}

		req := createTypedJSONRequest(http.MethodPatch, "/posts/1", requestData)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		assert.Equal(t, http.StatusNotFound, response.Code)
		mockService.AssertExpectations(t)
	})
}

func TestDeletePost(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, postHandler := setupTestPostHandler()
		r := setupPostRouter(postHandler)

		mockService.On("Delete", mock.Anything, mock.Anything).Return(nil)

		req := createTypedJSONRequest(http.MethodDelete, "/posts/1", nil)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		assert.Equal(t, http.StatusNoContent, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("BindingError_InvalidID", func(t *testing.T) {
		mockService, postHandler := setupTestPostHandler()
		r := setupPostRouter(postHandler)

		// invalid id parameter
		req := createTypedJSONRequest(http.MethodDelete, "/posts/invalid", nil)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		assert.Equal(t, http.StatusBadRequest, response.Code)
		mockService.AssertNotCalled(t, "Delete")
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService, postHandler := setupTestPostHandler()
		r := setupPostRouter(postHandler)

		mockService.On("Delete", mock.Anything, mock.Anything).Return(apperrors.ErrNotFound)

		req := createTypedJSONRequest(http.MethodDelete, "/posts/1", nil)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		assert.Equal(t, http.StatusNotFound, response.Code)
		mockService.AssertExpectations(t)
	})
}

func TestGetPosts(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, postHandler := setupTestPostHandler()
		r := setupPostRouter(postHandler)

		expectedResponse := &model.CursorResponse[model.Post]{
			Data:    []model.Post{*createTestPost()},
			Next:    "",
			HasMore: false,
		}

		mockService.On("List", mock.Anything).Return(expectedResponse, nil)

		req := createTypedJSONRequest(http.MethodGet, "/posts?limit=10", nil)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("BindQueryError", func(t *testing.T) {
		mockService, postHandler := setupTestPostHandler()
		r := setupPostRouter(postHandler)

		// invalid limit parameter (exceeds max value 100)
		req := createTypedJSONRequest(http.MethodGet, "/posts?limit=200", nil)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		assert.Equal(t, http.StatusBadRequest, response.Code)
		mockService.AssertNotCalled(t, "List")
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockService, postHandler := setupTestPostHandler()
		r := setupPostRouter(postHandler)

		mockService.On("List", mock.Anything).Return(nil, apperrors.ErrValidation)

		req := createTypedJSONRequest(http.MethodGet, "/posts?limit=10", nil)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		assert.Equal(t, http.StatusBadRequest, response.Code)
		mockService.AssertExpectations(t)
	})
}

func TestGetPostByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockService, postHandler := setupTestPostHandler()
		r := setupPostRouter(postHandler)

		expected := createTestPost()
		mockService.On("GetByID", mock.Anything).Return(expected, nil)

		NonExistentPostIDStr := strconv.FormatUint(NonExistentPostID, 10)
		req := createTypedJSONRequest(http.MethodGet, "/posts/"+NonExistentPostIDStr, nil)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("BindingError_InvalidID", func(t *testing.T) {
		mockService, postHandler := setupTestPostHandler()
		r := setupPostRouter(postHandler)

		// invalid id parameter
		req := createTypedJSONRequest(http.MethodGet, "/posts/invalid", nil)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		assert.Equal(t, http.StatusBadRequest, response.Code)
		mockService.AssertNotCalled(t, "GetByID")
	})

	t.Run("NotFound", func(t *testing.T) {
		mockService, postHandler := setupTestPostHandler()
		r := setupPostRouter(postHandler)

		mockService.On("GetByID", mock.Anything).Return(nil, apperrors.ErrNotFound)

		NonExistentPostIDStr := strconv.FormatUint(NonExistentPostID, 10)
		req := createTypedJSONRequest(http.MethodGet, "/posts/"+NonExistentPostIDStr, nil)

		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)

		assert.Equal(t, http.StatusNotFound, response.Code)
		mockService.AssertExpectations(t)
	})
}
