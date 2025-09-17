package mcp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"kapua-mcp-server/internal/config"
	"kapua-mcp-server/internal/kapua/handlers"
	"kapua-mcp-server/internal/kapua/services"
)

func TestNewServerSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/authentication/user":
			io.WriteString(w, `{"tokenId":"token","refreshToken":"refresh","expiresOn":"2025-01-02T15:04:05Z","refreshExpiresOn":"2025-01-03T15:04:05Z","scopeId":"tenant"}`)
		case strings.HasSuffix(r.URL.Path, "/devices"):
			io.WriteString(w, `{"items":[]}`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	cfg := &config.Config{Kapua: config.KapuaConfig{APIEndpoint: ts.URL, Timeout: 5}}

	srv, err := NewServer(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewServer returned error: %v", err)
	}
	if srv.Handler() == nil {
		t.Fatal("expected handler")
	}
}

func TestNewServerAuthFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"code":"ERR"}`)
	}))
	defer ts.Close()

	cfg := &config.Config{Kapua: config.KapuaConfig{APIEndpoint: ts.URL, Timeout: 5}}

	_, err := NewServer(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected authentication failure")
	}
	if !strings.Contains(err.Error(), "failed to authenticate") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegisterKapuaHelpers(t *testing.T) {
	kapuaHandler := handlers.NewKapuaHandler(&services.KapuaClient{})
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "test", Version: "dev"}, nil)

	registerKapuaTools(server, kapuaHandler)
	registerKapuaResources(server, kapuaHandler)
}
