package api

import (
	"github.com/gin-gonic/gin"

	"github.com/samik-k21/research-compute-queue/internal/api/handlers"
	"github.com/samik-k21/research-compute-queue/internal/api/middleware"
	"github.com/samik-k21/research-compute-queue/internal/database"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(db *database.DB) *gin.Engine {
	// Create router
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())      // Recover from panics
	router.Use(middleware.Logger()) // Log requests

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	jobHandler := handlers.NewJobHandler(db)

	// Health check (no auth required)
	router.GET("/health", handlers.HealthCheck)

	// API routes
	api := router.Group("/api")
	{
		// Authentication routes (no auth required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Job routes (auth required)
		jobs := api.Group("/jobs")
		jobs.Use(middleware.AuthRequired()) // Apply auth middleware
		{
			jobs.POST("", jobHandler.SubmitJob)
			jobs.GET("", jobHandler.ListJobs)
			jobs.GET("/:id", jobHandler.GetJob)
			jobs.DELETE("/:id", jobHandler.CancelJob)
		}

		// TODO: Queue routes
		// TODO: Admin routes
	}

	return router
}