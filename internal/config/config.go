package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port string
	Env  string

	// Firebase configuration
	FirebaseProjectID   string
	FirebasePrivateKey  string
	FirebaseClientEmail string
	FirebaseDatabaseURL string

	// OpenAI configuration
	OpenAIAPIKey string

	// Gemini configuration
	GeminiAPIKey string

	// Google Cloud configuration
	GoogleCredentialsPath string
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	err := godotenv.Load(filepath.Join(dir, "../../.env"))
	if err != nil {
		// Try to load from env.example if .env doesn't exist
		_ = godotenv.Load(filepath.Join(dir, "../../env.example"))
	}

	config := &Config{
		Port: getEnv("PORT", "8080"),
		Env:  getEnv("ENV", "development"),

		FirebaseProjectID:   getEnv("FIREBASE_PROJECT_ID", ""),
		FirebasePrivateKey:  getEnv("FIREBASE_PRIVATE_KEY", ""),
		FirebaseClientEmail: getEnv("FIREBASE_CLIENT_EMAIL", ""),
		FirebaseDatabaseURL: getEnv("FIREBASE_DATABASE_URL", ""),

		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),

		GoogleCredentialsPath: getEnv("GOOGLE_APPLICATION_CREDENTIALS", ""),
	}

	// Validate required configuration
	if config.Env == "production" {
		if config.FirebaseProjectID == "" {
			return nil, fmt.Errorf("FIREBASE_PROJECT_ID is required in production")
		}
		if config.FirebasePrivateKey == "" {
			return nil, fmt.Errorf("FIREBASE_PRIVATE_KEY is required in production")
		}
		if config.FirebaseClientEmail == "" {
			return nil, fmt.Errorf("FIREBASE_CLIENT_EMAIL is required in production")
		}
		if config.OpenAIAPIKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is required in production")
		}
	}

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
