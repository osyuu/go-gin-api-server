package middleware

import "github.com/gin-gonic/gin"

func ApiMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		println("Before request...")
		c.Next()
		println("After request...")
	}
}
