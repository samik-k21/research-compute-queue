package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck returns API health status
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "Research Compute Queue API is running",
		"version": "1.0.0",
	})
}