package handler

import (
	"errors"
	"go-gin-api-server/internal/middleware"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/apperrors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PostHandler struct {
	service service.PostService
	logger  *zap.Logger
}

func NewPostHandler(service service.PostService, logger *zap.Logger) *PostHandler {
	return &PostHandler{
		service: service,
		logger:  logger,
	}
}

func (h *PostHandler) RegisterRoutes(r *gin.Engine) {
	router := r.Group("/api/v1")
	{
		router.GET("/posts", h.GetPosts)
		router.GET("/posts/:id", h.GetPostByID)
	}
}

func (h *PostHandler) RegisterProtectedRoutes(r *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	protected := r.Group("/api/v1/posts")
	protected.Use(authMiddleware.RequireAuth())
	{
		protected.POST("", h.CreatePost)
		protected.PATCH("/:id", h.UpdatePost)
		protected.DELETE("/:id", h.DeletePost)
	}
}

// GetPosts retrieves a paginated list of posts
//
// Examples:
//
//	GET /api/v1/posts?limit=10
//	GET /api/v1/posts?limit=10&cursor=eyJpZCI6IjEiLCJjcmVhdGVkX2F0IjoiMjAyNC0wMS0wMVQwODowMDowMFoifQ==
//	GET /api/v1/posts?limit=10&author_id=user123
func (h *PostHandler) GetPosts(c *gin.Context) {
	// Parse cursor request parameters
	var cursorReq model.CursorRequest
	if err := BindQuery(c, &cursorReq); err != nil {
		h.handlePostError(c, apperrors.ErrValidation, "GetPosts")
		return
	}

	// Get posts with cursor pagination
	response, err := h.service.List(cursorReq)
	if err != nil {
		h.handlePostError(c, err, "GetPosts")
		return
	}

	h.handlePostSuccess(c, response, http.StatusOK)
}

// GetPostByID retrieves a single post by its ID
//
// Example:
//
//	GET /api/v1/posts/123
func (h *PostHandler) GetPostByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.handlePostError(c, apperrors.ErrValidation, "GetPostByID")
		return
	}

	found, err := h.service.GetByID(id)
	if err != nil {
		h.handlePostError(c, err, "GetPostByID")
		return
	}

	h.handlePostSuccess(c, found, http.StatusOK)
}

// CreatePost creates a new post (requires authentication)
//
// Example:
//
//	POST /api/v1/posts
//	{
//	  "content": "This is my new post content"
//	}
func (h *PostHandler) CreatePost(c *gin.Context) {
	var newPost model.Post

	if err := BindJSON(c, &newPost); err != nil {
		return
	}

	// Get current user ID from auth context
	userID := c.GetString("user_id")
	newPost.AuthorID = userID

	created, err := h.service.Create(&newPost)
	if err != nil {
		h.handlePostError(c, err, "CreatePost")
		return
	}

	h.handlePostSuccess(c, created, http.StatusCreated)
}

// UpdatePost updates an existing post (requires authentication and ownership)
//
// Example:
//
//	PATCH /api/v1/posts/123
//	{
//	  "content": "Updated post content"
//	}
func (h *PostHandler) UpdatePost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.handlePostError(c, apperrors.ErrValidation, "UpdatePost")
		return
	}

	var update model.Post
	if err := BindJSON(c, &update); err != nil {
		return
	}

	// Get current user ID from auth context
	userID := c.GetString("user_id")

	updated, err := h.service.Update(id, &update, userID)
	if err != nil {
		h.handlePostError(c, err, "UpdatePost")
		return
	}

	h.handlePostSuccess(c, updated, http.StatusOK)
}

// DeletePost deletes a post (requires authentication and ownership)
//
// Example:
//
//	DELETE /api/v1/posts/123
func (h *PostHandler) DeletePost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.handlePostError(c, apperrors.ErrValidation, "DeletePost")
		return
	}

	// Get current user ID from auth context
	userID := c.GetString("user_id")

	if err := h.service.Delete(id, userID); err != nil {
		h.handlePostError(c, err, "DeletePost")
		return
	}

	h.handlePostSuccess(c, nil, http.StatusNoContent)
}

// Helper functions

func (h *PostHandler) handlePostError(c *gin.Context, err error, operation string) {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		h.logger.Info("Post not found", zap.String("operation", operation), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Post not found",
		})
	case errors.Is(err, apperrors.ErrValidation):
		h.logger.Info("Validation error", zap.String("operation", operation), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Validation failed",
		})
	case errors.Is(err, apperrors.ErrForbidden):
		h.logger.Info("Permission denied", zap.String("operation", operation), zap.Error(err))
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Permission denied",
		})
	case errors.Is(err, apperrors.ErrUnauthorized):
		h.logger.Info("Unauthorized", zap.String("operation", operation), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
	case errors.Is(err, apperrors.ErrPostContentTooLong):
		h.logger.Info("Post content too long", zap.String("operation", operation), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Post content is too long",
		})
	case errors.Is(err, apperrors.ErrPostContentTooShort):
		h.logger.Info("Post content too short", zap.String("operation", operation), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Post content is too short",
		})
	case errors.Is(err, apperrors.ErrPostContentSensitiveWords):
		h.logger.Info("Post content contains sensitive words", zap.String("operation", operation), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Post content contains inappropriate language",
		})
	default:
		h.logger.Error("Unexpected error", zap.String("operation", operation), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	}
}

func (h *PostHandler) handlePostSuccess(c *gin.Context, data interface{}, statusCode int) {
	if data != nil {
		c.JSON(statusCode, data)
	} else {
		c.Status(statusCode)
	}
}
