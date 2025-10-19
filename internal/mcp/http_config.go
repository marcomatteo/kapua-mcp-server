package mcp

import (
	"fmt"
	"os"
	"strings"
)

// HTTPConfig controls the HTTP transport settings for the MCP server.
type HTTPConfig struct {
	Host           string
	Port           int
	AllowedOrigins []string

	rawOrigins []string
}

// LoadHTTPConfig builds an HTTPConfig process environment.
// It defaults to localhost:8000 when no explicit host/port are provided.
func LoadHTTPConfig() (*HTTPConfig, error) {
	cfg := &HTTPConfig{
		Host: "localhost",
		Port: 8000,
	}

	var origins []string

	if envOrigins := os.Getenv("MCP_ALLOWED_ORIGINS"); envOrigins != "" {
		origins = append(origins, parseAllowedOrigins(envOrigins)...)
	}

	cfg.SetAllowedOrigins(origins)

	return cfg, nil
}

// SetPort updates the configured port and recomputes the derived origin list.
func (cfg *HTTPConfig) SetPort(port int) {
	cfg.Port = port
	cfg.recompute()
}

// SetHost updates the configured host.
func (cfg *HTTPConfig) SetHost(host string) {
	cfg.Host = host
}

// SetAllowedOrigins replaces the set of configured origins and recomputes the
// derived list with defaults applied.
func (cfg *HTTPConfig) SetAllowedOrigins(origins []string) {
	cfg.rawOrigins = append([]string(nil), origins...)
	cfg.recompute()
}

func (cfg *HTTPConfig) recompute() {
	cfg.AllowedOrigins = normalizeAllowedOrigins(cfg.rawOrigins, cfg.Port)
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
