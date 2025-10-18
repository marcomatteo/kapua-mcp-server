package main

import (
	"context"
	"errors"
	"net/http"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"kapua-mcp-server/internal/config"
)

type stubKapuaServer struct {
	handler                http.Handler
	listenErr              error
	listenAddr             string
	receivedHandler        http.Handler
	listenCalled           bool
	runTransportCalled     bool
	runTransportErr        error
	runTransportName       string
	runTransportTransport  mcpsdk.Transport
	runTransportContextNil bool
}

func (s *stubKapuaServer) Handler() http.Handler {
	if s.handler == nil {
		s.handler = http.NewServeMux()
	}
	return s.handler
}

func (s *stubKapuaServer) ListenAndServe(addr string, handler http.Handler) error {
	s.listenCalled = true
	s.listenAddr = addr
	s.receivedHandler = handler
	return s.listenErr
}

func (s *stubKapuaServer) RunTransport(ctx context.Context, name string, transport mcpsdk.Transport) error {
	s.runTransportCalled = true
	s.runTransportName = name
	s.runTransportTransport = transport
	s.runTransportContextNil = ctx == nil
	return s.runTransportErr
}

func TestRunServerHTTP(t *testing.T) {
	stub := &stubKapuaServer{}
	oldNewServer := newServer
	defer func() { newServer = oldNewServer }()

	newServer = func(ctx context.Context, cfg *config.Config) (kapuaServer, error) {
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
		return stub, nil
	}

	err := runServer(&config.Config{}, "http", "localhost:0")
	if err != nil {
		t.Fatalf("runServer returned error: %v", err)
	}
	if !stub.listenCalled {
		t.Fatalf("expected ListenAndServe to be called")
	}
	if stub.listenAddr != "localhost:0" {
		t.Fatalf("expected address localhost:0, got %s", stub.listenAddr)
	}
	if stub.receivedHandler == nil {
		t.Fatalf("expected logging handler to be passed")
	}
	if stub.runTransportCalled {
		t.Fatalf("expected RunTransport not to be called for HTTP transport")
	}
}

func TestRunServerNewServerError(t *testing.T) {
	oldNewServer := newServer
	defer func() { newServer = oldNewServer }()

	wantErr := errors.New("boom")
	newServer = func(ctx context.Context, cfg *config.Config) (kapuaServer, error) {
		return nil, wantErr
	}

	err := runServer(&config.Config{}, "http", "localhost:0")
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

	err := runServer(&config.Config{}, "http", "localhost:0")
	if err == nil || !errors.Is(err, listenErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunServerStdio(t *testing.T) {
	stub := &stubKapuaServer{}
	oldNewServer := newServer
	defer func() { newServer = oldNewServer }()

	newServer = func(ctx context.Context, cfg *config.Config) (kapuaServer, error) {
		return stub, nil
	}

	err := runServer(&config.Config{}, "stdio", "ignored")
	if err != nil {
		t.Fatalf("runServer returned error: %v", err)
	}
	if !stub.runTransportCalled {
		t.Fatalf("expected RunTransport to be called for stdio transport")
	}
	if stub.runTransportName != "stdio" {
		t.Fatalf("unexpected transport name: %s", stub.runTransportName)
	}
	if stub.runTransportTransport == nil {
		t.Fatal("expected transport to be provided")
	}
	if _, ok := stub.runTransportTransport.(*mcpsdk.LoggingTransport); !ok {
		t.Fatalf("expected LoggingTransport, got %T", stub.runTransportTransport)
	}
	if stub.runTransportContextNil {
		t.Fatal("expected non-nil context for RunTransport")
	}
	if stub.listenCalled {
		t.Fatalf("expected ListenAndServe not to be called for stdio transport")
	}
}

func TestRunServerStdioError(t *testing.T) {
	transportErr := errors.New("transport failure")
	stub := &stubKapuaServer{runTransportErr: transportErr}
	oldNewServer := newServer
	defer func() { newServer = oldNewServer }()

	newServer = func(ctx context.Context, cfg *config.Config) (kapuaServer, error) {
		return stub, nil
	}

	err := runServer(&config.Config{}, "stdio", "ignored")
	if err == nil || !errors.Is(err, transportErr) {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stub.runTransportCalled {
		t.Fatalf("expected RunTransport to be called for stdio transport")
	}
}

func TestRunServerUnsupportedTransport(t *testing.T) {
	stub := &stubKapuaServer{}
	oldNewServer := newServer
	defer func() { newServer = oldNewServer }()

	newServer = func(ctx context.Context, cfg *config.Config) (kapuaServer, error) {
		return stub, nil
	}

	err := runServer(&config.Config{}, "invalid", "ignored")
	if err == nil || err.Error() != "unsupported transport \"invalid\"" {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.runTransportCalled || stub.listenCalled {
		t.Fatalf("expected no transport methods to be invoked for unsupported transport")
	}
}
