package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.uber.org/zap"
)

type RateLimitMiddleware struct {
	store  limiter.Store
	logger *zap.Logger
}

// RateLimitConfig
type RateLimitConfig struct {
	// Limit
	Limit int64
	// Period (seconds)
	Period time.Duration
	// By User ID (true: by user ID, false: by IP)
	ByUser bool
}

func NewRateLimitMiddleware(logger *zap.Logger) *RateLimitMiddleware {
	store := memory.NewStore()

	return &RateLimitMiddleware{
		store:  store,
		logger: logger,
	}
}

func (m *RateLimitMiddleware) RateLimit(config RateLimitConfig) gin.HandlerFunc {
	// Create rate limiter
	rate := limiter.Rate{
		Limit:  config.Limit,
		Period: config.Period,
	}

	instance := limiter.New(m.store, rate)

	return func(c *gin.Context) {
		key := m.getRateLimitKey(c, config.ByUser)

		// --------------------
		// type Context struct {
		// Limit     int64     // 限制次數 (例如: 10)
		// Remaining int64     // Remaining (例如: 7)
		// Reset     int64     // Reset (例如: 1640995200)
		// Reached   bool      // Reached (true/false)
		// --------------------

		// Check rate limit
		context, err := instance.Get(c, key)
		if err != nil {
			m.logger.Error("Failed to get rate limit", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
			return
		}

		// Set rate limit information to response header
		c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		// Check if exceeded limit
		if context.Reached {
			m.logger.Warn("Rate limit exceeded",
				zap.String("key", key),
				zap.Int64("limit", context.Limit),
				zap.Int64("remaining", context.Remaining),
			)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": context.Reset - time.Now().Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Get Rate Limit Key
func (m *RateLimitMiddleware) getRateLimitKey(c *gin.Context, byUser bool) string {
	if byUser {
		// By User ID
		userID, exists := c.Get("user_id")
		if exists {
			return fmt.Sprintf("user:%s", userID)
		}
		// If no user ID, fallback to IP limit
	}

	// By IP
	clientIP := c.ClientIP()
	return fmt.Sprintf("ip:%s", clientIP)
}

// Default Rate Limit

func (m *RateLimitMiddleware) StrictRateLimit() gin.HandlerFunc {
	return m.RateLimit(RateLimitConfig{
		Limit:  10,          // 10次
		Period: time.Minute, // 1分鐘
		ByUser: true,
	})
}

func (m *RateLimitMiddleware) ModerateRateLimit() gin.HandlerFunc {
	return m.RateLimit(RateLimitConfig{
		Limit:  100,         // 100次
		Period: time.Minute, // 1分鐘
		ByUser: true,
	})
}

func (m *RateLimitMiddleware) LooseRateLimit() gin.HandlerFunc {
	return m.RateLimit(RateLimitConfig{
		Limit:  1000,      // 1000次
		Period: time.Hour, // 1小時
		ByUser: false,     // 按IP限流
	})
}

// IPRateLimit
func (m *RateLimitMiddleware) IPRateLimit(limit int64, period time.Duration) gin.HandlerFunc {
	return m.RateLimit(RateLimitConfig{
		Limit:  limit,
		Period: period,
		ByUser: false,
	})
}

// UserRateLimit
func (m *RateLimitMiddleware) UserRateLimit(limit int64, period time.Duration) gin.HandlerFunc {
	return m.RateLimit(RateLimitConfig{
		Limit:  limit,
		Period: period,
		ByUser: true,
	})
}
