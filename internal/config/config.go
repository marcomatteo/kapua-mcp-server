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
	MCP   MCPConfig   `json:"mcp"`
}

// KapuaConfig holds Kapua-specific configuration
type KapuaConfig struct {
	APIEndpoint string `json:"api_endpoint"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Timeout     int    `json:"timeout"` // in seconds
}

// MCPConfig holds MCP server configuration
type MCPConfig struct {
	Host           string   `json:"host"`
	Port           int      `json:"port"`
	AllowedOrigins []string `json:"allowed_origins"`
}

// Load loads configuration from environment variables and .venv file
func Load() (*Config, error) {
	config := &Config{
		Kapua: KapuaConfig{
			Timeout: 30, // default timeout
		},
		MCP: MCPConfig{
			Host: "localhost",
			Port: 8000,
		},
	}

	// Load from .venv file first
	if err := loadFromEnvFile(config, ".venv"); err != nil {
		// .venv file not found or not readable, continue with env variables
	}

	// Override with environment variables
	loadFromEnv(config)

	config.MCP.AllowedOrigins = normalizeAllowedOrigins(config.MCP.AllowedOrigins, config.MCP.Port)

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
		case "MCP_ALLOWED_ORIGINS":
			config.MCP.AllowedOrigins = append(config.MCP.AllowedOrigins, parseAllowedOrigins(value)...)
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
	if origins := os.Getenv("MCP_ALLOWED_ORIGINS"); origins != "" {
		config.MCP.AllowedOrigins = append(config.MCP.AllowedOrigins, parseAllowedOrigins(origins)...)
	}
}

func parseAllowedOrigins(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p != "" {
			origins = append(origins, p)
		}
	}
	return origins
}

func normalizeAllowedOrigins(origins []string, port int) []string {
	originMap := make(map[string]struct{})
	ordered := make([]string, 0, len(origins))

	add := func(values []string) {
		for _, v := range values {
			key := strings.TrimSpace(v)
			if key == "" {
				continue
			}
			if _, exists := originMap[key]; exists {
				continue
			}
			originMap[key] = struct{}{}
			ordered = append(ordered, key)
		}
	}

	add(origins)

	defaults := defaultAllowedOrigins(port)
	add(defaults)

	if _, ok := originMap["*"]; ok {
		return []string{"*"}
	}

	return ordered
}

func defaultAllowedOrigins(port int) []string {
	hosts := []string{"localhost", "127.0.0.1", "::1", "[::1]", "0.0.0.0", "host.docker.internal"}
	schemes := []string{"http", "https"}
	var origins []string
	for _, scheme := range schemes {
		for _, host := range hosts {
			origins = append(origins, fmt.Sprintf("%s://%s", scheme, host))
			if port > 0 {
				origins = append(origins, fmt.Sprintf("%s://%s:%d", scheme, host, port))
			}
		}
	}
	return origins
}
