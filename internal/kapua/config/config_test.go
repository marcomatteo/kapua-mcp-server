package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFromEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".venv")
	content := "KAPUA_API_ENDPOINT=http://example.com/api\nKAPUA_USER=user\nKAPUA_PASSWORD=pass\n# comment\nINVALID_LINE\n"
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	cfg := &Config{}
	if err := loadFromEnvFile(cfg, envFile); err != nil {
		t.Fatalf("loadFromEnvFile returned error: %v", err)
	}

	if cfg.Kapua.APIEndpoint != "http://example.com/api" {
		t.Errorf("APIEndpoint not loaded, got %q", cfg.Kapua.APIEndpoint)
	}
	if cfg.Kapua.Username != "user" {
		t.Errorf("Username not loaded, got %q", cfg.Kapua.Username)
	}
	if cfg.Kapua.Password != "pass" {
		t.Errorf("Password not loaded, got %q", cfg.Kapua.Password)
	}
}

func TestLoadMissingRequired(t *testing.T) {
	t.Setenv("KAPUA_API_ENDPOINT", "")
	t.Setenv("KAPUA_USER", "")
	t.Setenv("KAPUA_PASSWORD", "")
	if _, err := Load(); err == nil {
		t.Fatalf("expected error when required fields missing")
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("KAPUA_API_ENDPOINT", "http://example.com/api")
	t.Setenv("KAPUA_USER", "user")
	t.Setenv("KAPUA_PASSWORD", "pass")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Kapua.APIEndpoint != "http://example.com/api" {
		t.Errorf("unexpected APIEndpoint: %q", cfg.Kapua.APIEndpoint)
	}
	if cfg.Kapua.Timeout != 30 {
		t.Errorf("expected default Timeout 30, got %d", cfg.Kapua.Timeout)
	}
}

func TestEnvOverridesFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".venv")
	content := "KAPUA_API_ENDPOINT=http://file/api\nKAPUA_USER=fileuser\nKAPUA_PASSWORD=filepass\n"
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	cfg := &Config{}
	if err := loadFromEnvFile(cfg, envFile); err != nil {
		t.Fatalf("loadFromEnvFile returned error: %v", err)
	}

	t.Setenv("KAPUA_API_ENDPOINT", "http://env/api")
	t.Setenv("KAPUA_USER", "envuser")
	t.Setenv("KAPUA_PASSWORD", "envpass")

	if err := loadFromEnv(cfg); err != nil {
		t.Fatalf("loadFromEnv returned error: %v", err)
	}

	if cfg.Kapua.APIEndpoint != "http://env/api" {
		t.Errorf("APIEndpoint not overridden, got %q", cfg.Kapua.APIEndpoint)
	}
	if cfg.Kapua.Username != "envuser" {
		t.Errorf("Username not overridden, got %q", cfg.Kapua.Username)
	}
	if cfg.Kapua.Password != "envpass" {
		t.Errorf("Password not overridden, got %q", cfg.Kapua.Password)
	}
}

func TestLoadTimeout(t *testing.T) {
	t.Setenv("KAPUA_API_ENDPOINT", "http://example.com/api")
	t.Setenv("KAPUA_USER", "user")
	t.Setenv("KAPUA_PASSWORD", "pass")
	t.Setenv("KAPUA_TIMEOUT", "60")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Kapua.Timeout != 60 {
		t.Errorf("expected Timeout 60, got %d", cfg.Kapua.Timeout)
	}
}

func TestLoadTimeoutInvalidEnv(t *testing.T) {
	t.Setenv("KAPUA_API_ENDPOINT", "http://example.com/api")
	t.Setenv("KAPUA_USER", "user")
	t.Setenv("KAPUA_PASSWORD", "pass")
	t.Setenv("KAPUA_TIMEOUT", "30s")

	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "invalid KAPUA_TIMEOUT") {
		t.Fatalf("expected invalid KAPUA_TIMEOUT error, got %v", err)
	}
}

func TestLoadTimeoutZeroEnv(t *testing.T) {
	t.Setenv("KAPUA_API_ENDPOINT", "http://example.com/api")
	t.Setenv("KAPUA_USER", "user")
	t.Setenv("KAPUA_PASSWORD", "pass")
	t.Setenv("KAPUA_TIMEOUT", "0")

	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "invalid KAPUA_TIMEOUT") {
		t.Fatalf("expected invalid KAPUA_TIMEOUT error, got %v", err)
	}
}

func TestLoadTimeoutFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".venv")
	content := "KAPUA_TIMEOUT=45\n"
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	cfg := &Config{Kapua: KapuaConfig{Timeout: 30}}
	if err := loadFromEnvFile(cfg, envFile); err != nil {
		t.Fatalf("loadFromEnvFile returned error: %v", err)
	}
	if cfg.Kapua.Timeout != 45 {
		t.Errorf("expected Timeout 45, got %d", cfg.Kapua.Timeout)
	}
}

func TestLoadTimeoutInvalidInFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".venv")
	content := "KAPUA_TIMEOUT=30s\n"
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	cfg := &Config{Kapua: KapuaConfig{Timeout: 30}}
	if err := loadFromEnvFile(cfg, envFile); err == nil || !strings.Contains(err.Error(), "invalid KAPUA_TIMEOUT") {
		t.Fatalf("expected invalid KAPUA_TIMEOUT error, got %v", err)
	}
}

func TestLoadTimeoutZeroInFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".venv")
	content := "KAPUA_TIMEOUT=0\n"
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	cfg := &Config{Kapua: KapuaConfig{Timeout: 30}}
	if err := loadFromEnvFile(cfg, envFile); err == nil || !strings.Contains(err.Error(), "invalid KAPUA_TIMEOUT") {
		t.Fatalf("expected invalid KAPUA_TIMEOUT error, got %v", err)
	}
}

func TestLoadAPIKeyAuthMethod(t *testing.T) {
	t.Setenv("KAPUA_API_ENDPOINT", "http://example.com/api")
	t.Setenv("KAPUA_AUTH_METHOD", "apikey")
	t.Setenv("KAPUA_API_KEY", "my-api-key")
	t.Setenv("KAPUA_USER", "")
	t.Setenv("KAPUA_PASSWORD", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Kapua.AuthMethod != "apikey" {
		t.Errorf("expected AuthMethod apikey, got %q", cfg.Kapua.AuthMethod)
	}
	if cfg.Kapua.APIKey != "my-api-key" {
		t.Errorf("expected APIKey my-api-key, got %q", cfg.Kapua.APIKey)
	}
}

func TestLoadAPIKeyMissingKey(t *testing.T) {
	t.Setenv("KAPUA_API_ENDPOINT", "http://example.com/api")
	t.Setenv("KAPUA_AUTH_METHOD", "apikey")
	t.Setenv("KAPUA_API_KEY", "")
	t.Setenv("KAPUA_USER", "")
	t.Setenv("KAPUA_PASSWORD", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected error when KAPUA_AUTH_METHOD=apikey but no KAPUA_API_KEY")
	}
}

func TestLoadUnsupportedAuthMethod(t *testing.T) {
	t.Setenv("KAPUA_API_ENDPOINT", "http://example.com/api")
	t.Setenv("KAPUA_AUTH_METHOD", "oauth")
	t.Setenv("KAPUA_USER", "user")
	t.Setenv("KAPUA_PASSWORD", "pass")

	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "unsupported KAPUA_AUTH_METHOD") {
		t.Fatalf("expected unsupported auth method error, got %v", err)
	}
}

func TestLoadDefaultAuthMethod(t *testing.T) {
	t.Setenv("KAPUA_API_ENDPOINT", "http://example.com/api")
	t.Setenv("KAPUA_USER", "user")
	t.Setenv("KAPUA_PASSWORD", "pass")
	t.Setenv("KAPUA_AUTH_METHOD", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Kapua.AuthMethod != "password" {
		t.Errorf("expected default AuthMethod password, got %q", cfg.Kapua.AuthMethod)
	}
}

func TestLoadAPIKeyFromEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".venv")
	content := "KAPUA_AUTH_METHOD=apikey\nKAPUA_API_KEY=file-api-key\n"
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	cfg := &Config{Kapua: KapuaConfig{AuthMethod: "password"}}
	if err := loadFromEnvFile(cfg, envFile); err != nil {
		t.Fatalf("loadFromEnvFile returned error: %v", err)
	}
	if cfg.Kapua.AuthMethod != "apikey" {
		t.Errorf("expected AuthMethod apikey, got %q", cfg.Kapua.AuthMethod)
	}
	if cfg.Kapua.APIKey != "file-api-key" {
		t.Errorf("expected APIKey file-api-key, got %q", cfg.Kapua.APIKey)
	}
}
