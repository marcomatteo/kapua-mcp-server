package mcp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"kapua-mcp-server/pkg/utils"
)

func newTestLogger() *utils.Logger {
	return utils.NewDefaultLogger("OriginGuardTest")
}

func TestOriginMiddlewareAllowsMatchingHost(t *testing.T) {
	cfg := &HTTPConfig{
		AllowedOrigins: []string{"http://localhost"},
	}
	logger := newTestLogger()

	handler := newOriginMiddleware(cfg, logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "http://localhost/", nil)
	req.Header.Set("Origin", "http://localhost")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected OK, got %d", rr.Code)
	}
}

func TestOriginMiddlewareAllowsLoopbackEquivalence(t *testing.T) {
	cfg := &HTTPConfig{}
	logger := newTestLogger()

	handler := newOriginMiddleware(cfg, logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "http://localhost/", nil)
	req.Header.Set("Origin", "http://127.0.0.1:3456")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected OK, got %d", rr.Code)
	}
}

func TestOriginMiddlewareAllowsConfiguredPattern(t *testing.T) {
	cfg := &HTTPConfig{AllowedOrigins: []string{"https://example.com:9000"}}
	logger := newTestLogger()

	handler := newOriginMiddleware(cfg, logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "http://localhost/", nil)
	req.Header.Set("Origin", "https://example.com:9000")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected OK, got %d", rr.Code)
	}
}

func TestOriginMiddlewareRejectsUnknownHost(t *testing.T) {
	cfg := &HTTPConfig{}
	logger := newTestLogger()

	handler := newOriginMiddleware(cfg, logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "http://localhost/", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", rr.Code)
	}
}

func TestOriginMiddlewareWildcardAllowsAll(t *testing.T) {
	cfg := &HTTPConfig{AllowedOrigins: []string{"*"}}
	logger := newTestLogger()

	handler := newOriginMiddleware(cfg, logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "http://localhost/", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected OK, got %d", rr.Code)
	}
}

func TestOriginMiddlewareAllowsMissingOrigin(t *testing.T) {
	cfg := &HTTPConfig{}
	logger := newTestLogger()

	handler := newOriginMiddleware(cfg, logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "http://localhost/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected OK, got %d", rr.Code)
	}
}
