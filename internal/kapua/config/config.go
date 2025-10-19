package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config holds all configuration for the Kapua MCP server
type Config struct {
	Kapua KapuaConfig `json:"kapua"`
}

// KapuaConfig holds Kapua-specific configuration
type KapuaConfig struct {
	APIEndpoint string `json:"api_endpoint"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Timeout     int    `json:"timeout"` // in seconds
}

// Load loads configuration from environment variables and .venv file
func Load() (*Config, error) {
	config := &Config{
		Kapua: KapuaConfig{
			Timeout: 30, // default timeout
		},
	}

	// Load from .venv file first
	if err := loadFromEnvFile(config, ".venv"); err != nil {
		// .venv file not found or not readable, continue with env variables
	}
	loadFromEnv(config)

	// Validate required fields
	if config.Kapua.APIEndpoint == "" {
		return nil, fmt.Errorf("KAPUA_API_ENDPOINT is required")
	}
	if config.Kapua.Username == "" {
		return nil, fmt.Errorf("KAPUA_USER is required")
	}
	if config.Kapua.Password == "" {
		return nil, fmt.Errorf("KAPUA_PASSWORD is required")
	}

	return config, nil
}

// loadFromEnvFile loads configuration from a .env style file
func loadFromEnvFile(config *Config, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Set configuration values
		switch key {
		case "KAPUA_API_ENDPOINT":
			config.Kapua.APIEndpoint = value
		case "KAPUA_USER":
			config.Kapua.Username = value
		case "KAPUA_PASSWORD":
			config.Kapua.Password = value
		}
	}

	return scanner.Err()
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	if endpoint := os.Getenv("KAPUA_API_ENDPOINT"); endpoint != "" {
		config.Kapua.APIEndpoint = endpoint
	}
	if username := os.Getenv("KAPUA_USER"); username != "" {
		config.Kapua.Username = username
	}
	if password := os.Getenv("KAPUA_PASSWORD"); password != "" {
		config.Kapua.Password = password
	}
}
