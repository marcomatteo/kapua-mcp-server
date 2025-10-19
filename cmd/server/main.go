package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"kapua-mcp-server/internal/kapua/config"
	mcpserver "kapua-mcp-server/internal/mcp"
)

type kapuaServer interface {
	Handler(*mcpserver.HTTPConfig) http.Handler
	ListenAndServe(string, http.Handler) error
	RunTransport(context.Context, string, mcpsdk.Transport) error
}

var newServer = func(ctx context.Context, cfg *config.Config) (kapuaServer, error) {
	return mcpserver.NewServer(ctx, cfg)
}

var (
	httpMode = flag.Bool("http", false, "Run the MCP server with the HTTP streamable transport instead of stdio")
	host     = flag.String("host", "localhost", "For http-streamable server, the host to listen on")
	port     = flag.Int("port", 8000, "For http-streamable server, the port number to listen on")
)

func main() {
	out := flag.CommandLine.Output()
	flag.Usage = func() {
		fmt.Fprintf(out, "Usage: %s [-http] [-port <port>] [-host <host>]\n\n", os.Args[0])
		fmt.Fprintf(out, "Kapua MCP Server for Eclipse Kapua IoT Device Management.\n")
		fmt.Fprintf(out, "Options:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()

	kapuaCfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	var httpCfg *mcpserver.HTTPConfig
	if *httpMode {
		httpCfg, err = mcpserver.LoadHTTPConfig()
		if err != nil {
			log.Fatalf("Failed to load MCP HTTP configuration: %v", err)
		}
		if *host != "localhost" {
			httpCfg.SetHost(*host)
		}
		if *port != 8000 {
			httpCfg.SetPort(*port)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	srv, err := newServer(ctx, kapuaCfg)
	if err != nil {
		log.Fatalf("Failed to initialise MCP server: %v", err)
	}

	if *httpMode {
		if err := runHTTPServer(srv, httpCfg); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	} else {
		if err := runStdioServer(srv); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}
}

func runHTTPServer(srv kapuaServer, httpCfg *mcpserver.HTTPConfig) error {
	if httpCfg == nil {
		return fmt.Errorf("http transport requires configuration")
	}

	addr := fmt.Sprintf("%s:%d", httpCfg.Host, httpCfg.Port)
	handlerWithLogging := LoggingHandler(srv.Handler(httpCfg))
	if err := srv.ListenAndServe(addr, handlerWithLogging); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

func runStdioServer(srv kapuaServer) error {
	ctx := context.Background()
	loggingTransport := &mcpsdk.LoggingTransport{Transport: &mcpsdk.StdioTransport{}, Writer: os.Stderr}
	if err := srv.RunTransport(ctx, "stdio", loggingTransport); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
