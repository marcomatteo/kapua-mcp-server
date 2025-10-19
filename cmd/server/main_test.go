package main

import (
	"context"
	"errors"
	"net/http"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	mcpserver "kapua-mcp-server/internal/mcp"
)

type stubKapuaServer struct {
	handler                http.Handler
	listenErr              error
	listenAddr             string
	receivedHandler        http.Handler
	receivedHTTPConfig     *mcpserver.HTTPConfig
	listenCalled           bool
	runTransportCalled     bool
	runTransportErr        error
	runTransportName       string
	runTransportTransport  mcpsdk.Transport
	runTransportContextNil bool
}

func (s *stubKapuaServer) Handler(cfg *mcpserver.HTTPConfig) http.Handler {
	s.receivedHTTPConfig = cfg
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

func TestRunHTTPServer(t *testing.T) {
	stub := &stubKapuaServer{}
	httpCfg := &mcpserver.HTTPConfig{Host: "localhost", Port: 0}

	if err := runHTTPServer(stub, httpCfg); err != nil {
		t.Fatalf("runHTTPServer returned error: %v", err)
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
	if stub.receivedHTTPConfig != httpCfg {
		t.Fatalf("expected HTTP config to be forwarded")
	}
	if stub.runTransportCalled {
		t.Fatalf("expected RunTransport not to be called for HTTP transport")
	}
}

func TestRunHTTPServerNilConfig(t *testing.T) {
	if err := runHTTPServer(&stubKapuaServer{}, nil); err == nil || err.Error() != "http transport requires configuration" {
		t.Fatalf("expected configuration error, got %v", err)
	}
}

func TestRunHTTPServerListenError(t *testing.T) {
	stub := &stubKapuaServer{listenErr: errors.New("listen failure")}

	err := runHTTPServer(stub, &mcpserver.HTTPConfig{Host: "localhost", Port: 0})
	if err == nil || !errors.Is(err, stub.listenErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunStdioServer(t *testing.T) {
	stub := &stubKapuaServer{}

	if err := runStdioServer(stub); err != nil {
		t.Fatalf("runStdioServer returned error: %v", err)
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

func TestRunStdioServerTransportError(t *testing.T) {
	stub := &stubKapuaServer{runTransportErr: errors.New("transport failure")}

	err := runStdioServer(stub)
	if err == nil || !errors.Is(err, stub.runTransportErr) {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stub.runTransportCalled {
		t.Fatalf("expected RunTransport to be called for stdio transport")
	}
}
