package mcp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"unsafe"

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

	// Access the server's internal tools registry via reflection to ensure the new
	// tool has been registered. The MCP SDK does not currently expose a public API
	// for enumerating tools outside of the JSON-RPC surface area, so reflection is
	// used purely for test verification.
	toolsField := reflect.ValueOf(server).Elem().FieldByName("tools")
	if !toolsField.IsValid() {
		t.Fatal("tools field not found on MCP server")
	}

	toolsValue := reflect.NewAt(toolsField.Type(), unsafe.Pointer(toolsField.UnsafeAddr())).Elem()
	featuresField := toolsValue.Elem().FieldByName("features")
	if !featuresField.IsValid() {
		t.Fatal("features map not found on tools registry")
	}
	expectedTools := []string{
		"kapua-list-devices",
		"kapua-list-device-events",
		"kapua-list-data-messages",
		"kapua-configurations-read",
		"kapua-inventory-read",
		"kapua-inventory-bundles",
		"kapua-inventory-bundle-start",
		"kapua-inventory-bundle-stop",
		"kapua-inventory-containers",
		"kapua-inventory-container-start",
		"kapua-inventory-container-stop",
		"kapua-inventory-system-packages",
		"kapua-inventory-deployment-packages",
	}

	if got := featuresField.Len(); got < len(expectedTools) {
		t.Fatalf("expected at least %d tools to be registered, got %d", len(expectedTools), got)
	}

	registered := make(map[string]struct{})
	for _, key := range featuresField.MapKeys() {
		registered[key.String()] = struct{}{}
	}

	for _, name := range expectedTools {
		if _, ok := registered[name]; !ok {
			t.Fatalf("expected %s tool to be registered", name)
		}
	}
}
