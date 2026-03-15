package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
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
	APIKey      string `json:"api_key"`    // API key for KAPUA_AUTH_METHOD=apikey
	AuthMethod  string `json:"auth_method"` // "password" (default) or "apikey"
	Timeout     int    `json:"timeout"`    // in seconds
}

// Load loads configuration from environment variables and .venv file
func Load() (*Config, error) {
	config := &Config{
		Kapua: KapuaConfig{
			Timeout:    30, // default timeout
			AuthMethod: "password",
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

	switch config.Kapua.AuthMethod {
	case "apikey":
		if config.Kapua.APIKey == "" {
			return nil, fmt.Errorf("KAPUA_API_KEY is required when KAPUA_AUTH_METHOD=apikey")
		}
	default:
		config.Kapua.AuthMethod = "password"
		if config.Kapua.Username == "" {
			return nil, fmt.Errorf("KAPUA_USER is required")
		}
		if config.Kapua.Password == "" {
			return nil, fmt.Errorf("KAPUA_PASSWORD is required")
		}
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
		case "KAPUA_API_KEY":
			config.Kapua.APIKey = value
		case "KAPUA_AUTH_METHOD":
			config.Kapua.AuthMethod = value
		case "KAPUA_TIMEOUT":
			if v, err := strconv.Atoi(value); err == nil {
				config.Kapua.Timeout = v
			}
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
	if apiKey := os.Getenv("KAPUA_API_KEY"); apiKey != "" {
		config.Kapua.APIKey = apiKey
	}
	if authMethod := os.Getenv("KAPUA_AUTH_METHOD"); authMethod != "" {
		config.Kapua.AuthMethod = authMethod
	}
	if timeout := os.Getenv("KAPUA_TIMEOUT"); timeout != "" {
		if v, err := strconv.Atoi(timeout); err == nil {
			config.Kapua.Timeout = v
		}
	}
}
