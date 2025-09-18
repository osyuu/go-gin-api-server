package middleware

import (
	"go-gin-api-server/internal/service"
	"go-gin-api-server/pkg/apperrors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthMiddleware struct {
	authService service.AuthService
	logger      *zap.Logger
}

func NewAuthMiddleware(authService service.AuthService, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
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
			// 如果是 Access Token 過期，嘗試自動刷新
			if err == apperrors.ErrExpiredToken {
				if m.tryAutoRefresh(c) {
					return // 自動刷新成功，已經設置了新的 token
				}
			}
			m.handleAuthError(c, err, "Token validation failed")
			return
		}

		// 4. store user ID, role to context
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
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
			// 如果是 Access Token 過期，嘗試自動刷新
			if err == apperrors.ErrExpiredToken {
				if m.tryAutoRefresh(c) {
					return // 自動刷新成功，已經設置了新的 token
				}
			}
			// token無效，繼續執行但不設置user_id
			c.Next()
			return
		}

		// token is valid, set user ID and role to context
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// tryAutoRefresh 嘗試自動刷新 Access Token
func (m *AuthMiddleware) tryAutoRefresh(c *gin.Context) bool {
	// 1. 從 cookie 中獲取 refresh token
	refreshToken, err := c.Cookie("gin_api_refresh_token")
	if err != nil {
		m.logger.Debug("No refresh token found in cookie", zap.Error(err))
		return false
	}

	// 2. 使用 refresh token 獲取新的 access token（不刷新 refresh token）
	newAccessToken, err := m.authService.RefreshAccessToken(refreshToken)
	if err != nil {
		m.logger.Debug("Failed to refresh access token", zap.Error(err))
		return false
	}

	// 3. 在響應頭中設置新的 access token
	c.Header("X-New-Access-Token", newAccessToken)
	c.Header("X-Token-Type", "Bearer")

	// 4. 驗證新的 access token 並設置 user_id
	claims, err := m.authService.ValidateToken(newAccessToken)
	if err != nil {
		m.logger.Error("Failed to validate new access token", zap.Error(err))
		return false
	}

	c.Set("user_id", claims.UserID)
	c.Set("user_role", claims.Role)
	c.Next()
	return true
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
		m.logger.Error("Unexpected error in auth middleware",
			zap.String("operation", operation),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	}
	c.Abort()
}
