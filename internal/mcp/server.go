package mcp

import (
	"context"
	"fmt"
	"net/http"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"kapua-mcp-server/internal/config"
	"kapua-mcp-server/internal/kapua/handlers"
	"kapua-mcp-server/internal/kapua/services"
	"kapua-mcp-server/pkg/utils"
)

type Server struct {
	logger    *utils.Logger
	handler   http.Handler
	cfg       *config.Config
	mcpServer *mcpsdk.Server
}

func NewServer(ctx context.Context, cfg *config.Config) (*Server, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	logger := utils.NewDefaultLogger("MCPServer")
	logger.Info("Starting Kapua MCP Server")

	kapuaClient := services.NewKapuaClient(&cfg.Kapua)

	logger.Info("Authenticating to Kapua on startup...")
	if _, err := kapuaClient.QuickAuthenticate(ctx); err != nil {
		return nil, fmt.Errorf("failed to authenticate to Kapua on startup: %w", err)
	}
	logger.Info("Successfully authenticated to Kapua")

	kapuaHandler := handlers.NewKapuaHandler(kapuaClient)

	sdkServer := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "kapua-mcp-server",
		Version: "1.0.0",
	}, nil)

	registerKapuaTools(sdkServer, kapuaHandler)
	registerKapuaResources(sdkServer, kapuaHandler)

	httpHandler := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server {
		return sdkServer
	}, nil)

	return &Server{
		logger:    logger,
		handler:   httpHandler,
		cfg:       cfg,
		mcpServer: sdkServer,
	}, nil
}

func (s *Server) Handler() http.Handler {
	return s.handler
}

func (s *Server) ListenAndServe(addr string, handler http.Handler) error {
	s.logger.Info("MCP server listening on %s", addr)
	s.logger.Info("Kapua API endpoint: %s", s.cfg.Kapua.APIEndpoint)
	s.logger.Info("Available Kapua tools:")
	s.logger.Info("  - kapua-list-devices: List IoT devices in Kapua with filtering options.")
	s.logger.Info("  - kapua-list-device-events: List device log events for a Kapua device.")
	s.logger.Info("  - kapua-update-device: Update an existing Kapua device")
	s.logger.Info("  - kapua-delete-device: Delete a device")
	s.logger.Info("  - kapua-configurations-read: Read all configurations for a device")

	return http.ListenAndServe(addr, handler)
}

func registerKapuaTools(server *mcpsdk.Server, kapuaHandler *handlers.KapuaHandler) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-list-devices",
		Description: "List Kapua IoT devices",
	}, kapuaHandler.HandleListDevices)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-list-device-events",
		Description: "List Kapua device events (logs)",
	}, kapuaHandler.HandleListDeviceEvents)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-update-device",
		Description: "Update an existing Kapua device",
	}, kapuaHandler.HandleUpdateDevice)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-delete-device",
		Description: "Delete a Kapua device",
	}, kapuaHandler.HandleDeleteDevice)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-configurations-read",
		Description: "Read all configurations for a Kapua device (input: {id})",
	}, kapuaHandler.HandleDeviceConfigurationsRead)
}

func registerKapuaResources(server *mcpsdk.Server, kapuaHandler *handlers.KapuaHandler) {
	server.AddResource(&mcpsdk.Resource{
		URI:         "kapua://devices",
		Name:        "Kapua Devices",
		Description: "Live list of Kapua IoT devices",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
		return kapuaHandler.ReadResource(ctx, req.Params.URI)
	})
}
