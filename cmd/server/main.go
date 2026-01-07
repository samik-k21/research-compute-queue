package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/samik-k21/research-compute-queue/internal/api"
	"github.com/samik-k21/research-compute-queue/internal/config"
	"github.com/samik-k21/research-compute-queue/internal/database"
)

func main() {
	// Load configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal("Configuration validation failed:", err)
	}

	log.Printf("Starting Research Compute Queue API in %s mode", cfg.Environment)

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	log.Println("Database connection established")

	// Create necessary directories
	if err := os.MkdirAll(cfg.LogDirectory, 0755); err != nil {
		log.Fatal("Failed to create log directory:", err)
	}
	if err := os.MkdirAll(cfg.OutputDirectory, 0755); err != nil {
		log.Fatal("Failed to create output directory:", err)
	}

	// Set up API router
	router := api.SetupRouter(db)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	log.Println("API server is running. Press Ctrl+C to stop")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}