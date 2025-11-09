package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates Bearer token for protected endpoints
func AuthMiddleware(apiToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health and docs endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/health") ||
			strings.HasPrefix(c.Request.URL.Path, "/ready") ||
			strings.HasPrefix(c.Request.URL.Path, "/live") ||
			strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			c.Next()
			return
		}

		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization format, expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		// Validate token
		token := parts[1]
		if token != apiToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			c.Abort()
			return
		}

		// Token is valid, continue
		c.Next()
	}
}
