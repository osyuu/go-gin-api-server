package handler

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/apperrors"
	"net/http"

	"github.com/gin-gonic/gin"
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

	c.JSON(http.StatusCreated, tokenResponse)
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

	c.JSON(http.StatusOK, tokenResponse)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 從 cookie 中獲取 refresh token
	refreshToken, err := c.Cookie("gin_api_refresh_token")
	if err != nil {
		h.handleAuthError(c, apperrors.ErrUnauthorized, "RefreshToken")
		return
	}

	tokenResponse, err := h.authService.RefreshToken(refreshToken)
	if err != nil {
		h.handleAuthError(c, err, "RefreshToken")
		return
	}

	// 更新 refresh token 到 cookie（7天有效期）
	c.SetCookie("gin_api_refresh_token", tokenResponse.RefreshToken,
		7*24*60*60, "/api", "", true, true) // 7天，限制路徑，Secure, HttpOnly

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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	}
}
