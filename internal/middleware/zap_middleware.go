package middleware

import (
	"blog_server/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GinZapMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime)
		if duration > 1*time.Second {
			logger.Log.Warn("slow request",
				zap.String("method", method),
				zap.String("path", path),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("duration", duration),
			)
		} else {
			logger.Log.Info("request completed",
				zap.String("method", method),
				zap.String("path", path),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("duration", duration),
			)
		}
	}
}
