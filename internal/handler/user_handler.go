package handler

import (
	"errors"
	"go-gin-api-server/internal/middleware"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/apperrors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Request structures for better type safety and validation
type CreateUserRequest struct {
	Name      string     `json:"name" binding:"required,min=3"`
	Username  string     `json:"username" binding:"required,min=3,max=50,username"`
	Email     string     `json:"email" binding:"required,email"`
	BirthDate *time.Time `json:"birth_date,omitempty"`
}

type UpdateUserProfileRequest struct {
	Name      string     `json:"name,omitempty" binding:"omitempty,min=3"`
	BirthDate *time.Time `json:"birth_date,omitempty"`
}

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) RegisterRoutes(r *gin.Engine) {
	// Public routes
	r.POST("/api/v1/users", h.CreateUser)
}

func (h *UserHandler) RegisterProtectedRoutes(r *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	// Protected routes
	protected := r.Group("/api/v1/users")
	protected.Use(authMiddleware.RequireAuth())
	{
		// Get user info
		protected.GET("/:id", h.GetUserByID)
		protected.GET("/username/:username", h.GetUserByUsername)
		protected.GET("/email/:email", h.GetUserByEmail)

		// User management operations
		protected.PATCH("/:id", h.UpdateUserProfile)
		protected.PATCH("/:id/activate", h.ActivateUser)
		protected.PATCH("/:id/deactivate", h.DeactivateUser)
		protected.DELETE("/:id", h.DeleteUser)
	}
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	user, err := h.service.GetUserByID(id)

	if err != nil {
		h.handleUserError(c, err, "GetUserByID")
		return
	}

	h.handleSuccess(c, user, http.StatusOK)
}

func (h *UserHandler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	user, err := h.service.GetUserByUsername(username)
	if err != nil {
		h.handleUserError(c, err, "GetUserByUsername")
		return
	}
	h.handleSuccess(c, user, http.StatusOK)
}

func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	user, err := h.service.GetUserByEmail(email)
	if err != nil {
		h.handleUserError(c, err, "GetUserByEmail")
		return
	}
	h.handleSuccess(c, user, http.StatusOK)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := h.bindRequest(c, &req); err != nil {
		return
	}
	createdUser, err := h.service.CreateUser(
		req.Name,
		req.Username,
		req.Email,
		req.BirthDate,
	)
	if err != nil {
		h.handleUserError(c, err, "CreateUser")
		return
	}
	h.handleSuccess(c, createdUser, http.StatusCreated)
}

func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	userID := c.Param("id")

	// 從請求體獲取更新數據
	var update UpdateUserProfileRequest

	if err := h.bindRequest(c, &update); err != nil {
		return
	}

	// 檢查是否有任何更新選項
	if update.Name == "" && update.BirthDate == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No update fields provided",
		})
		return
	}

	updated, err := h.service.UpdateUserProfile(userID, update.Name, update.BirthDate)
	if err != nil {
		h.handleUserError(c, err, "UpdateUserProfile")
		return
	}

	h.handleSuccess(c, updated, http.StatusOK)
}

func (h *UserHandler) ActivateUser(c *gin.Context) {
	userID := c.Param("id")
	err := h.service.ActivateUser(userID)
	if err != nil {
		h.handleUserError(c, err, "ActivateUser")
		return
	}
	h.handleSuccess(c, nil, http.StatusNoContent)
}

func (h *UserHandler) DeactivateUser(c *gin.Context) {
	userID := c.Param("id")
	err := h.service.DeactivateUser(userID)
	if err != nil {
		h.handleUserError(c, err, "DeactivateUser")
		return
	}
	h.handleSuccess(c, nil, http.StatusNoContent)
}

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

func (h *UserHandler) bindRequest(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindBodyWithJSON(obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return err
	}
	return nil
}

func (h *UserHandler) handleUserError(c *gin.Context, err error, operation string) {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
	case errors.Is(err, apperrors.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Validation failed",
		})
	case errors.Is(err, apperrors.ErrUserExists):
		c.JSON(http.StatusConflict, gin.H{
			"error": "User already exists",
		})
	case errors.Is(err, apperrors.ErrUserUnderAge):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User under age",
		})
	case errors.Is(err, apperrors.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
	default:
		log.Printf("Unexpected error in %s: %v", operation, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	}
}

// 處理成功響應
func (h *UserHandler) handleSuccess(c *gin.Context, data interface{}, statusCode int) {
	if data != nil {
		c.JSON(statusCode, data)
	} else {
		c.Status(statusCode)
	}
}
