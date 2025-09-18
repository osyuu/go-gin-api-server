package handler

import (
	"go-gin-api-server/internal/middleware"
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/apperrors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService service.AuthService
	logger      *zap.Logger
}

func NewAuthHandler(authService service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
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

func (h *AuthHandler) RegisterProtectedRoutes(r *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	// Protected auth routes (admin operations)
	protected := r.Group("/api/v1/auth")
	protected.Use(authMiddleware.RequireAuth())
	{
		protected.POST("/users/:id/activate", h.ActivateUser)
		protected.POST("/users/:id/deactivate", h.DeactivateUser)
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := BindJSON(c, &req); err != nil {
		return
	}

	tokenResponse, err := h.authService.Register(&req)
	if err != nil {
		h.handleAuthError(c, err, "Register")
		return
	}

	// 設置 refresh token 到 cookie（7天有效期）
	c.SetCookie("gin_api_refresh_token", tokenResponse.RefreshToken,
		7*24*60*60, "/api", "", true, true) // 7天，限制路徑，Secure, HttpOnly

	h.handleAuthSuccess(c, tokenResponse, http.StatusCreated)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := BindJSON(c, &req); err != nil {
		return
	}

	tokenResponse, err := h.authService.Login(&req)
	if err != nil {
		h.handleAuthError(c, err, "Login")
		return
	}

	// 設置 refresh token 到 cookie（7天有效期）
	c.SetCookie("gin_api_refresh_token", tokenResponse.RefreshToken,
		7*24*60*60, "/api", "", true, true) // 7天，限制路徑，Secure, HttpOnly

	h.handleAuthSuccess(c, tokenResponse, http.StatusOK)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 從 cookie 中獲取 refresh token
	refreshToken, err := c.Cookie("gin_api_refresh_token")
	if err != nil {
		refreshToken = ""
	}

	tokenResponse, err := h.authService.RefreshToken(refreshToken)
	if err != nil {
		h.handleAuthError(c, err, "RefreshToken")
		return
	}

	// 更新 refresh token 到 cookie（7天有效期）
	c.SetCookie("gin_api_refresh_token", tokenResponse.RefreshToken,
		7*24*60*60, "/api", "", true, true) // 7天，限制路徑，Secure, HttpOnly

	h.handleAuthSuccess(c, tokenResponse, http.StatusOK)
}

func (h *AuthHandler) ActivateUser(c *gin.Context) {
	userID := c.Param("id")

	currentUserRole, err := GetUserRole(c)
	if err != nil {
		h.handleAuthError(c, err, "ActivateUser")
		return
	}

	user, err := h.authService.ActivateUser(userID, currentUserRole)
	if err != nil {
		h.handleAuthError(c, err, "ActivateUser")
		return
	}

	h.handleAuthSuccess(c, user, http.StatusOK)
}

func (h *AuthHandler) DeactivateUser(c *gin.Context) {
	userID := c.Param("id")
	currentUserID, currentUserRole, err := GetUserIDAndRole(c)
	if err != nil {
		h.handleAuthError(c, err, "DeactivateUser")
		return
	}

	user, err := h.authService.DeactivateUser(userID, currentUserID, currentUserRole)
	if err != nil {
		h.handleAuthError(c, err, "DeactivateUser")
		return
	}

	h.handleAuthSuccess(c, user, http.StatusOK)
}

func (h *AuthHandler) handleAuthError(c *gin.Context, err error, _ string) {
	switch err {
	case apperrors.ErrValidation:
		h.logger.Error("Invalid request format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
	case apperrors.ErrUserExists:
		h.logger.Error("User already exists", zap.Error(err))
		c.JSON(http.StatusConflict, gin.H{
			"error": "User already exists",
		})
	case apperrors.ErrUserUnderAge:
		h.logger.Error("User under age", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User under age",
		})
	case apperrors.ErrUnauthorized:
		h.logger.Error("Unauthorized", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
	case apperrors.ErrForbidden:
		h.logger.Error("Forbidden", zap.Error(err))
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Forbidden",
		})
	case apperrors.ErrInvalidToken:
		h.logger.Error("Invalid token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
	case apperrors.ErrExpiredToken:
		h.logger.Error("Token has expired", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Token has expired",
		})
	default:
		h.logger.Error("Internal server error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	}
}

func (h *AuthHandler) handleAuthSuccess(c *gin.Context, data interface{}, statusCode int) {
	if data != nil {
		c.JSON(statusCode, data)
	} else {
		c.Status(statusCode)
	}
}
