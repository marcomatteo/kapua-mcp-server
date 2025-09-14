package services

import (
	"kapua-mcp-server/internal/config"
	"testing"
	"time"
)

func TestNewKapuaClientBaseURL(t *testing.T) {
	cfg := &config.KapuaConfig{APIEndpoint: "http://example.com/api", Timeout: 10}
	client := NewKapuaClient(cfg)
	expected := "http://example.com/api/v1"
	if client.baseURL != expected {
		t.Fatalf("expected baseURL %q, got %q", expected, client.baseURL)
	}
	if client.httpClient.Timeout != 10*time.Second {
		t.Fatalf("expected timeout 10s, got %v", client.httpClient.Timeout)
	}
	if !client.autoRefresh {
		t.Fatalf("expected autoRefresh to be true")
	}
}

func TestNewKapuaClientBaseURLAlreadyV1(t *testing.T) {
	cfg := &config.KapuaConfig{APIEndpoint: "http://example.com/api/v1", Timeout: 5}
	client := NewKapuaClient(cfg)
	expected := "http://example.com/api/v1"
	if client.baseURL != expected {
		t.Fatalf("expected baseURL %q, got %q", expected, client.baseURL)
	}
	if client.httpClient.Timeout != 5*time.Second {
		t.Fatalf("expected timeout 5s, got %v", client.httpClient.Timeout)
	}
}
