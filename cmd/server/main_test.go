package main

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"kapua-mcp-server/internal/config"
)

type stubKapuaServer struct {
	handler         http.Handler
	listenErr       error
	listenAddr      string
	receivedHandler http.Handler
	called          bool
}

func (s *stubKapuaServer) Handler() http.Handler {
	if s.handler == nil {
		s.handler = http.NewServeMux()
	}
	return s.handler
}

func (s *stubKapuaServer) ListenAndServe(addr string, handler http.Handler) error {
	s.called = true
	s.listenAddr = addr
	s.receivedHandler = handler
	return s.listenErr
}

func TestRunServerSuccess(t *testing.T) {
	stub := &stubKapuaServer{}
	oldNewServer := newServer
	defer func() { newServer = oldNewServer }()

	newServer = func(ctx context.Context, cfg *config.Config) (kapuaServer, error) {
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
		return stub, nil
	}

	err := runServer(&config.Config{}, "localhost:0")
	if err != nil {
		t.Fatalf("runServer returned error: %v", err)
	}
	if !stub.called {
		t.Fatalf("expected ListenAndServe to be called")
	}
	if stub.listenAddr != "localhost:0" {
		t.Fatalf("expected address localhost:0, got %s", stub.listenAddr)
	}
	if stub.receivedHandler == nil {
		t.Fatalf("expected logging handler to be passed")
	}
}

func TestRunServerNewServerError(t *testing.T) {
	oldNewServer := newServer
	defer func() { newServer = oldNewServer }()

	wantErr := errors.New("boom")
	newServer = func(ctx context.Context, cfg *config.Config) (kapuaServer, error) {
		return nil, wantErr
	}

	err := runServer(&config.Config{}, "localhost:0")
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped boom error, got %v", err)
	}
}

func TestRunServerListenError(t *testing.T) {
	listenErr := errors.New("listen failure")
	stub := &stubKapuaServer{listenErr: listenErr}
	oldNewServer := newServer
	defer func() { newServer = oldNewServer }()

	newServer = func(ctx context.Context, cfg *config.Config) (kapuaServer, error) {
		return stub, nil
	}

	err := runServer(&config.Config{}, "localhost:0")
	if err == nil || !errors.Is(err, listenErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}
