package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthRequired checks for valid JWT token
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement JWT validation
		// For now, we'll skip authentication to test other endpoints
		
		// Get token from header
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// TODO: Validate JWT token here
		// For now, accept any token
		
		c.Next()
	}
}