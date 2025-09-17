// Copyright 2025 The Go MCP SDK Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"kapua-mcp-server/internal/config"
	mcpserver "kapua-mcp-server/internal/mcp"
)

type kapuaServer interface {
	Handler() http.Handler
	ListenAndServe(string, http.Handler) error
}

var newServer = func(ctx context.Context, cfg *config.Config) (kapuaServer, error) {
	return mcpserver.NewServer(ctx, cfg)
}

var (
	host = flag.String("host", "localhost", "host to listen on")
	port = flag.Int("port", 8000, "port number to listen on")
)

func main() {
	out := flag.CommandLine.Output()
	flag.Usage = func() {
		fmt.Fprintf(out, "Usage: %s [-port <port] [-host <host>]\n\n", os.Args[0])
		fmt.Fprintf(out, "Kapua MCP Server for Eclipse Kapua IoT Device Management.\n")
		fmt.Fprintf(out, "Options:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if *host != "localhost" {
		cfg.MCP.Host = *host
	}
	if *port != 8000 {
		cfg.MCP.Port = *port
	}
	if err := runServer(cfg, fmt.Sprintf("%s:%d", cfg.MCP.Host, cfg.MCP.Port)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func runServer(cfg *config.Config, url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	srv, err := newServer(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialise MCP server: %w", err)
	}

	baseHandler := srv.Handler()
	handlerWithLogging := LoggingHandler(baseHandler)

	if err := srv.ListenAndServe(url, handlerWithLogging); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}
	return nil
}
