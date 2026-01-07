package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/samik-k21/research-compute-queue/internal/api"
	"github.com/samik-k21/research-compute-queue/internal/auth"
	"github.com/samik-k21/research-compute-queue/internal/config"
	"github.com/samik-k21/research-compute-queue/internal/database"
	"github.com/samik-k21/research-compute-queue/internal/scheduler"
)

func main() {
	// Load configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal("Configuration validation failed:", err)
	}

	log.Println("========================================")
	log.Printf("Research Compute Queue API")
	log.Printf("Environment: %s", cfg.Environment)
	log.Println("========================================")

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	log.Println("✓ Database connection established")

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiryHours)
	log.Println("✓ JWT manager initialized")

	// Create necessary directories
	if err := os.MkdirAll(cfg.LogDirectory, 0755); err != nil {
		log.Fatal("Failed to create log directory:", err)
	}
	if err := os.MkdirAll(cfg.OutputDirectory, 0755); err != nil {
		log.Fatal("Failed to create output directory:", err)
	}
	log.Println("✓ Directories created")

	// Initialize scheduler
	sched := scheduler.NewScheduler(db, cfg.SchedulerIntervalSecs, cfg.MaxConcurrentJobs)

	// Start scheduler in background
	go sched.Start()
	log.Printf("✓ Scheduler started (interval: %ds, max concurrent: %d)",
		cfg.SchedulerIntervalSecs, cfg.MaxConcurrentJobs)

	// Set up API router
	router := api.SetupRouter(db, jwtManager)

	// Start server in a goroutine
	go func() {
		log.Printf("✓ API server starting on port %s", cfg.Port)
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	log.Println("========================================")
	log.Println("System is ready!")
	log.Printf("API: http://localhost:%s", cfg.Port)
	log.Println("Press Ctrl+C to stop")
	log.Println("========================================")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("")
	log.Println("========================================")
	log.Println("Shutting down gracefully...")
	log.Println("========================================")

	// Stop scheduler
	sched.Stop()
	log.Println("✓ Scheduler stopped")

	// Close database
	db.Close()
	log.Println("✓ Database connection closed")

	log.Println("✓ Server stopped successfully")
}