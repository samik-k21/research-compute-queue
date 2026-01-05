package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	DatabaseURL  string
	Port         string
	Environment  string
	JWTSecret    string
	JWTExpiryHours         int
	SchedulerIntervalSecs  int
	MaxConcurrentJobs      int
	LogDirectory           string
	OutputDirectory        string
}

// Load reads configuration from environment variables
func Load() *Config {
	// Load .env file in development
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		Port:                  getEnv("PORT", "8080"),
		Environment:           getEnv("ENVIRONMENT", "development"),
		JWTSecret:             getEnv("JWT_SECRET", ""),
		JWTExpiryHours:        getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		SchedulerIntervalSecs: getEnvAsInt("SCHEDULER_INTERVAL_SECONDS", 30),
		MaxConcurrentJobs:     getEnvAsInt("MAX_CONCURRENT_JOBS", 10),
		LogDirectory:          getEnv("LOG_DIRECTORY", "./logs"),
		OutputDirectory:       getEnv("OUTPUT_DIRECTORY", "./output"),
	}
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt reads an environment variable as an integer or returns a default
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// Validate checks if required configuration values are set
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	return nil
}