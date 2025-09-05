package middleware

import (
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/apperrors"
	"go-gin-api-server/pkg/logger"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthMiddleware struct {
	authService service.AuthService
}

func NewAuthMiddleware(authService service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.handleAuthError(c, apperrors.ErrUnauthorized, "Missing authorization header")
			return
		}

		// 2. check Bearer format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.handleAuthError(c, apperrors.ErrUnauthorized, "Invalid authorization header format")
			return
		}

		token := parts[1]

		// 3. validate token
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			m.handleAuthError(c, err, "Token validation failed")
			return
		}

		// 4. store user ID to context
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// OptionalAuth 可選認證的中間件（用於某些需要知道用戶身份但不需要強制登錄的場景）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 沒有token，繼續執行但不設置user_id
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// 格式錯誤，繼續執行但不設置user_id
			c.Next()
			return
		}

		token := parts[1]
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			// token無效，繼續執行但不設置user_id
			c.Next()
			return
		}

		// token有效，設置user ID
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// handleAuthError 處理認證錯誤
func (m *AuthMiddleware) handleAuthError(c *gin.Context, err error, operation string) {
	switch err {
	case apperrors.ErrInvalidToken:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
		})
	case apperrors.ErrExpiredToken:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Token has expired",
		})
	case apperrors.ErrUnauthorized:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
	default:
		logger.Log.Error("Unexpected error in auth middleware",
			zap.String("operation", operation),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	}
	c.Abort()
}
