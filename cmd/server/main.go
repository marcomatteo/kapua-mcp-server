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

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"kapua-mcp-server/internal/config"
	"kapua-mcp-server/internal/handlers"
	"kapua-mcp-server/internal/services"
	"kapua-mcp-server/pkg/utils"
)

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
	runServer(cfg, fmt.Sprintf("%s:%d", cfg.MCP.Host, cfg.MCP.Port))
}

func runServer(cfg *config.Config, url string) {
	logger := utils.NewDefaultLogger("MCPServer")
	logger.Info("Starting Kapua MCP Server")

	// Create Kapua client
	kapuaClient := services.NewKapuaClient(&cfg.Kapua)

	// Authenticate to Kapua on startup
	logger.Info("Authenticating to Kapua on startup...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := kapuaClient.QuickAuthenticate(ctx)
	if err != nil {
		log.Fatalf("Failed to authenticate to Kapua on startup: %v", err)
	}
	logger.Info("Successfully authenticated to Kapua")

	kapuaHandler := handlers.NewKapuaHandler(kapuaClient)

	// Create an MCP server.
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "kapua-mcp-server",
		Version: "1.0.0",
	}, nil)

	// Add Kapua device management tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kapua-list-devices",
		Description: "List Kapua IoT devices",
	}, kapuaHandler.HandleListDevices)

	// Add Kapua device create/update/delete tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "kapua-create-device",
		Description: "Create a new Kapua device",
	}, kapuaHandler.HandleCreateDevice)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "kapua-update-device",
		Description: "Update an existing Kapua device",
	}, kapuaHandler.HandleUpdateDevice)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "kapua-delete-device",
		Description: "Delete a Kapua device",
	}, kapuaHandler.HandleDeleteDevice)

	// Register Kapua resources so clients can discover and read them
	server.AddResource(&mcp.Resource{
		URI:         "kapua://devices",
		Name:        "Kapua Devices",
		Description: "Live list of Kapua IoT devices",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return kapuaHandler.ReadResource(ctx, req.Params.URI)
	})

	// Create the streamable HTTP handler.
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return server
	}, nil)

	handlerWithLogging := LoggingHandler(handler)

	logger.Info("MCP server listening on %s", url)
	logger.Info("Kapua API endpoint: %s", cfg.Kapua.APIEndpoint)
	logger.Info("Available Kapua device management tools:")
	logger.Info("  - kapua-list-devices: List IoT devices in Kapua with filtering options.")
	logger.Info("  - kapua-create-device: Create a new device")
	logger.Info("  - kapua-update-device: Update an existing device")
	logger.Info("  - kapua-delete-device: Delete a device")
    logger.Info("Available Kapua resources:")
    logger.Info("  - kapua://devices (application/json)")

	// Start the HTTP server with logging handler.
	if err := http.ListenAndServe(url, handlerWithLogging); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
