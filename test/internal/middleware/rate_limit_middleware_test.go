package middleware

import (
	"go-gin-api-server/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupRateLimitTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	rateLimitMiddleware := middleware.NewRateLimitMiddleware(zap.NewNop())

	r.GET("/test", rateLimitMiddleware.StrictRateLimit(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 使用IP限流
	r.GET("/test-ip", rateLimitMiddleware.IPRateLimit(2, time.Minute), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 使用用戶限流
	r.GET("/test-user", rateLimitMiddleware.UserRateLimit(2, time.Minute), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	return r
}

func TestRateLimitMiddleware_StrictLimit(t *testing.T) {
	router := setupRateLimitTestRouter()

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("X-RateLimit-Limit"), "10")
	assert.Contains(t, w.Header().Get("X-RateLimit-Remaining"), "9")
}

func TestRateLimitMiddleware_ExceedLimit(t *testing.T) {
	router := setupRateLimitTestRouter()

	// 發送超過限制的請求
	for i := 0; i < 12; i++ { // 超過10次限制
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if i < 10 {
			// 前10次應該成功
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			// 第11次開始應該被限流
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}

func TestRateLimitMiddleware_IPBasedLimit(t *testing.T) {
	router := setupRateLimitTestRouter()

	// 模擬同一個IP的請求
	for i := 0; i < 4; i++ {
		req, _ := http.NewRequest("GET", "/test-ip", nil)
		req.RemoteAddr = "127.0.0.1:8080" // 固定IP
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if i < 2 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}

func TestRateLimitMiddleware_UserBasedLimit(t *testing.T) {
	router := setupRateLimitTestRouter()

	// 模擬已登錄用戶的請求
	for i := 0; i < 4; i++ {
		req, _ := http.NewRequest("GET", "/test-user", nil)
		w := httptest.NewRecorder()

		// 模擬設置用戶ID
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", "test-user-123")

		router.ServeHTTP(w, req)

		if i < 2 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}
