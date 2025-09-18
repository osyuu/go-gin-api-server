package middleware

import (
	"go-gin-api-server/internal/model"
	"go-gin-api-server/pkg/apperrors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RBACMiddleware provides role-based access control middleware functions
type RBACMiddleware struct {
	logger *zap.Logger
}

// NewRBACMiddleware creates a new RBAC middleware instance
func NewRBACMiddleware(logger *zap.Logger) *RBACMiddleware {
	return &RBACMiddleware{
		logger: logger,
	}
}

// RequireAdmin requires admin role to access
func (r *RBACMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			r.handleRBACError(c, apperrors.ErrUnauthorized, "RequireAdmin")
			return
		}

		userRole, ok := role.(model.UserRole)
		if !ok {
			r.handleRBACError(c, apperrors.ErrUnauthorized, "RequireAdmin")
			return
		}

		if !userRole.IsAdmin() {
			r.handleRBACError(c, apperrors.ErrForbidden, "RequireAdmin")
			return
		}

		c.Next()
	}
}

// RequireOwnershipOrAdmin requires user to be the resource owner or admin
func (r *RBACMiddleware) RequireOwnershipOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		// Get current user ID
		currentUserID, exists := c.Get("user_id")
		if !exists {
			r.handleRBACError(c, apperrors.ErrUnauthorized, "RequireOwnershipOrAdmin")
			return
		}

		userIDStr, ok := currentUserID.(string)
		if !ok {
			r.handleRBACError(c, apperrors.ErrUnauthorized, "RequireOwnershipOrAdmin")
			return
		}

		// Get user role
		role, exists := c.Get("user_role")
		if !exists {
			r.handleRBACError(c, apperrors.ErrUnauthorized, "RequireOwnershipOrAdmin")
			return
		}

		userRole, ok := role.(model.UserRole)
		if !ok {
			r.handleRBACError(c, apperrors.ErrUnauthorized, "RequireOwnershipOrAdmin")
			return
		}

		// Allow if user is admin or owns the resource
		if userRole.IsAdmin() || userIDStr == userID {
			c.Next()
			return
		}

		r.handleRBACError(c, apperrors.ErrForbidden, "RequireOwnershipOrAdmin")
	}
}

// RequireOwnership requires user to be the resource owner
func (r *RBACMiddleware) RequireOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		currentUserID, exists := c.Get("user_id")
		if !exists {
			r.handleRBACError(c, apperrors.ErrUnauthorized, "RequireOwnership")
			return
		}

		userIDStr, ok := currentUserID.(string)
		if !ok {
			r.handleRBACError(c, apperrors.ErrUnauthorized, "RequireOwnership")
			return
		}

		if userIDStr != userID {
			r.handleRBACError(c, apperrors.ErrForbidden, "RequireOwnership")
			return
		}

		c.Next()
	}
}

// handleRBACError handles RBAC-related errors
func (r *RBACMiddleware) handleRBACError(c *gin.Context, err error, operation string) {
	switch err {
	case apperrors.ErrUnauthorized:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
	case apperrors.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
	default:
		r.logger.Error("Unexpected error in RBAC middleware",
			zap.String("operation", operation),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	}
	c.Abort()
}
