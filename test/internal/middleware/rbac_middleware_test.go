package middleware

import (
	"go-gin-api-server/internal/middleware"
	"go-gin-api-server/internal/model"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Helper functions

func setupTestRBACMiddleware() *middleware.RBACMiddleware {
	return middleware.NewRBACMiddleware(zap.NewNop())
}

func setupTestRBACRouter(contextSetup gin.HandlerFunc, rbacMiddleware gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add context setup middleware first
	router.Use(contextSetup)

	// Add RBAC middleware
	router.Use(rbacMiddleware)

	router.GET("/protected", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"user_id": userID,
		})
	})

	router.GET("/protected/:id", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"user_id": userID,
		})
	})

	return router
}

// Test RequireAdmin

func TestRBACMiddleware_RequireAdmin(t *testing.T) {
	rbacMiddleware := setupTestRBACMiddleware()

	t.Run("Success_Admin", func(t *testing.T) {
		router := setupTestRBACRouter(func(c *gin.Context) {
			c.Set("user_role", model.RoleAdmin)
			c.Next()
		}, rbacMiddleware.RequireAdmin())

		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Forbidden_User", func(t *testing.T) {
		router := setupTestRBACRouter(func(c *gin.Context) {
			c.Set("user_role", model.RoleUser)
			c.Next()
		}, rbacMiddleware.RequireAdmin())

		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Unauthorized_NoRole", func(t *testing.T) {
		router := setupTestRBACRouter(func(c *gin.Context) {
			// Don't set user_role
			c.Next()
		}, rbacMiddleware.RequireAdmin())

		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// Test RequireOwnership

func TestRBACMiddleware_RequireOwnership(t *testing.T) {
	rbacMiddleware := setupTestRBACMiddleware()
	testUserID := "user-123"

	t.Run("Success_Owner", func(t *testing.T) {
		router := setupTestRBACRouter(func(c *gin.Context) {
			c.Set("user_id", testUserID)
			c.Next()
		}, rbacMiddleware.RequireOwnership())

		req, _ := http.NewRequest(http.MethodGet, "/protected/"+testUserID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Forbidden_DifferentUser", func(t *testing.T) {
		router := setupTestRBACRouter(func(c *gin.Context) {
			c.Set("user_id", "different-user")
			c.Next()
		}, rbacMiddleware.RequireOwnership())

		req, _ := http.NewRequest(http.MethodGet, "/protected/"+testUserID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Unauthorized_NoUserID", func(t *testing.T) {
		router := setupTestRBACRouter(func(c *gin.Context) {
			// Don't set user_id
			c.Next()
		}, rbacMiddleware.RequireOwnership())

		req, _ := http.NewRequest(http.MethodGet, "/protected/"+testUserID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// Test RequireOwnershipOrAdmin

func TestRBACMiddleware_RequireOwnershipOrAdmin(t *testing.T) {
	rbacMiddleware := setupTestRBACMiddleware()
	testUserID := "user-123"

	t.Run("Success_Owner", func(t *testing.T) {
		router := setupTestRBACRouter(func(c *gin.Context) {
			c.Set("user_id", testUserID)
			c.Set("user_role", model.RoleUser)
			c.Next()
		}, rbacMiddleware.RequireOwnershipOrAdmin())

		req, _ := http.NewRequest(http.MethodGet, "/protected/"+testUserID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Success_Admin", func(t *testing.T) {
		router := setupTestRBACRouter(func(c *gin.Context) {
			c.Set("user_id", "different-user")
			c.Set("user_role", model.RoleAdmin)
			c.Next()
		}, rbacMiddleware.RequireOwnershipOrAdmin())

		req, _ := http.NewRequest(http.MethodGet, "/protected/"+testUserID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Forbidden_NeitherOwnerNorAdmin", func(t *testing.T) {
		router := setupTestRBACRouter(func(c *gin.Context) {
			c.Set("user_id", "different-user")
			c.Set("user_role", model.RoleUser)
			c.Next()
		}, rbacMiddleware.RequireOwnershipOrAdmin())

		req, _ := http.NewRequest(http.MethodGet, "/protected/"+testUserID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
