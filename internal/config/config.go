package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	GroqAPIKey string
}

// LoadConfig loads the configuration from environment variables and .env file
func LoadConfig() (*Config, error) {
	// Load environment variables from global .env file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting user home directory: %v", err)
	}

	globalEnvPath := filepath.Join(homeDir, ".gitai.env")
	err = godotenv.Load(globalEnvPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error loading global .env file (%s): %v\n", globalEnvPath, err)
		fmt.Fprintf(os.Stderr, "Continuing with environment variables...\n")
	}

	// Check if GROQ_API_KEY is set
	groqAPIKey := os.Getenv("GROQ_API_KEY")
	if groqAPIKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY is not set. Please set it in %s or as an environment variable", globalEnvPath)
	}

	return &Config{
		GroqAPIKey: groqAPIKey,
	}, nil
}
