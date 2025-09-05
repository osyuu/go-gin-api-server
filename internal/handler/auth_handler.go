package handler

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/apperrors"
	"go-gin-api-server/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) RegisterRoutes(r *gin.Engine) {
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.RefreshToken)
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleAuthError(c, err, "Register")
		return
	}

	tokenResponse, err := h.authService.Register(&req)
	if err != nil {
		h.handleAuthError(c, err, "Register")
		return
	}

	c.JSON(http.StatusCreated, tokenResponse)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleAuthError(c, err, "Login")
		return
	}

	tokenResponse, err := h.authService.Login(&req)
	if err != nil {
		h.handleAuthError(c, err, "Login")
		return
	}

	c.JSON(http.StatusOK, tokenResponse)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleAuthError(c, err, "RefreshToken")
		return
	}

	tokenResponse, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		h.handleAuthError(c, err, "RefreshToken")
		return
	}

	c.JSON(http.StatusOK, tokenResponse)
}

func (h *AuthHandler) handleAuthError(c *gin.Context, err error, operation string) {
	switch err {
	case apperrors.ErrValidation:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
	case apperrors.ErrUserExists:
		c.JSON(http.StatusConflict, gin.H{
			"error": "User already exists",
		})
	case apperrors.ErrUserUnderAge:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User under age",
		})
	case apperrors.ErrUnauthorized:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
	case apperrors.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Forbidden",
		})
	case apperrors.ErrInvalidToken:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
	case apperrors.ErrExpiredToken:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Token has expired",
		})
	default:
		logger.Log.Error("Unexpected error in %s: %v", zap.String("operation", operation), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	}
}
