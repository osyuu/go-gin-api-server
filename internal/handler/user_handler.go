package handler

import (
	"errors"
	"go-gin-api-server/internal/middleware"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/apperrors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	service service.UserService
	logger  *zap.Logger
}

func NewUserHandler(service service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

func (h *UserHandler) RegisterRoutes(r *gin.Engine) {
	// Public routes - only safe user queries
	r.GET("/api/v1/users/profile/:username", h.GetUserProfile)
}

func (h *UserHandler) RegisterProtectedRoutes(r *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	// Protected routes
	protected := r.Group("/api/v1/users")
	protected.Use(authMiddleware.RequireAuth())
	{
		// Get user info (sensitive data)
		protected.GET("/:id", h.GetUserByID)
		protected.GET("/username/:username", h.GetUserByUsername)
		protected.GET("/email/:email", h.GetUserByEmail)

		// User management operations
		protected.PATCH("/:id", h.UpdateUserProfile)

		// Admin operations (not implemented)
		// protected.DELETE("/:id", h.DeleteUser)
	}
}

// GetUserProfile Get user public profile
//
// Example:
//
//	GET /api/v1/users/john_doe
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	var req struct {
		Username string `uri:"username" binding:"required,username"`
	}

	if err := BindUri(c, &req); err != nil {
		return
	}

	publicInfo, err := h.service.GetUserProfile(req.Username)
	if err != nil {
		h.handleUserError(c, err, "GetUserProfile")
		return
	}
	h.handleSuccess(c, publicInfo, http.StatusOK)
}

// GetUserByID Get user by ID
//
// Example:
//
//	GET /api/v1/users/550e8400-e29b-41d4-a716-446655440000
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.GetUserByID(id)

	if err != nil {
		h.handleUserError(c, err, "GetUserByID")
		return
	}

	h.handleSuccess(c, user, http.StatusOK)
}

// GetUserByUsername Get user by username
//
// Example:
//
//	GET /api/v1/users/username/john_doe
func (h *UserHandler) GetUserByUsername(c *gin.Context) {
	var req struct {
		Username string `uri:"username" binding:"required,username"`
	}

	if err := BindUri(c, &req); err != nil {
		return
	}

	user, err := h.service.GetUserByUsername(req.Username)
	if err != nil {
		h.handleUserError(c, err, "GetUserByUsername")
		return
	}
	h.handleSuccess(c, user, http.StatusOK)
}

// GetUserByEmail Get user by email
//
// Example:
//
//	GET /api/v1/users/email/user@example.com
func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	var req struct {
		Email string `uri:"email" binding:"required,email"`
	}

	if err := BindUri(c, &req); err != nil {
		return
	}

	user, err := h.service.GetUserByEmail(req.Email)
	if err != nil {
		h.handleUserError(c, err, "GetUserByEmail")
		return
	}
	h.handleSuccess(c, user, http.StatusOK)
}

// UpdateUserProfile Update user profile
//
// Example:
//
//	PATCH /api/v1/users/550e8400-e29b-41d4-a716-446655440000
//	{
//		"name": "New Name",
//		"birth_date": "1990-01-01T00:00:00Z"
//	}
func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	userID := c.Param("id")
	var update model.UpdateUserProfileRequest

	if err := BindJSON(c, &update); err != nil {
		return
	}

	if update.Name == "" && update.BirthDate == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No update fields provided",
		})
		return
	}

	// Get current user ID from auth context
	currentUserID := c.GetString("user_id")

	updated, err := h.service.UpdateUserProfile(userID, currentUserID, update)
	if err != nil {
		h.handleUserError(c, err, "UpdateUserProfile")
		return
	}

	h.handleSuccess(c, updated, http.StatusOK)
}

// DeleteUser Delete user
//
// Example:
//
//	DELETE /api/v1/users/550e8400-e29b-41d4-a716-446655440000
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	err := h.service.DeleteUser(userID)
	if err != nil {
		h.handleUserError(c, err, "DeleteUser")
		return
	}
	h.handleSuccess(c, nil, http.StatusNoContent)
}

// Helper functions

func (h *UserHandler) handleUserError(c *gin.Context, err error, _ string) {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		h.logger.Error("User not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
	case errors.Is(err, apperrors.ErrValidation):
		h.logger.Error("Validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Validation failed",
		})
	case errors.Is(err, apperrors.ErrUserExists):
		h.logger.Error("User already exists", zap.Error(err))
		c.JSON(http.StatusConflict, gin.H{
			"error": "User already exists",
		})
	case errors.Is(err, apperrors.ErrUserUnderAge):
		h.logger.Error("User under age", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User under age",
		})
	case errors.Is(err, apperrors.ErrUnauthorized):
		h.logger.Error("Unauthorized", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
	case errors.Is(err, apperrors.ErrForbidden):
		h.logger.Error("Forbidden", zap.Error(err))
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Forbidden",
		})
	default:
		h.logger.Error("Internal server error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	}
}

func (h *UserHandler) handleSuccess(c *gin.Context, data interface{}, statusCode int) {
	if data != nil {
		c.JSON(statusCode, data)
	} else {
		c.Status(statusCode)
	}
}
