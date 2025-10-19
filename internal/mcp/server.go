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

	streamHandler := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server {
		return sdkServer
	}, nil)

	httpHandler := newOriginMiddleware(cfg.MCP, logger, streamHandler)

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
	s.logStartup("streamable-http", addr)
	return http.ListenAndServe(addr, handler)
}

func (s *Server) RunTransport(ctx context.Context, transportName string, transport mcpsdk.Transport) error {
	if transport == nil {
		return fmt.Errorf("transport cannot be nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if transportName == "" {
		transportName = "custom"
	}

	s.logStartup(transportName, "")

	return s.mcpServer.Run(ctx, transport)
}

func (s *Server) logStartup(transportName, endpoint string) {
	if endpoint != "" {
		s.logger.Info("MCP server listening on %s via %s transport", endpoint, transportName)
	} else {
		s.logger.Info("MCP server using %s transport", transportName)
	}
	s.logger.Info("Kapua API endpoint: %s", s.cfg.Kapua.APIEndpoint)
	s.logAvailableTools()
}

func (s *Server) logAvailableTools() {
	s.logger.Info("Available Kapua tools:")
	s.logger.Info("  - kapua-list-devices: List IoT devices in Kapua with filtering options.")
	s.logger.Info("  - kapua-list-device-events: List device log events for a Kapua device.")
	s.logger.Info("  - kapua-list-device-logs: List logs across Kapua devices.")
	s.logger.Info("  - kapua-list-data-messages: List Kapua data messages with filtering options.")
	// s.logger.Info("  - kapua-update-device: Update an existing Kapua device")
	// s.logger.Info("  - kapua-delete-device: Delete a device")
	s.logger.Info("  - kapua-configurations-read: Read all configurations for a device")
	s.logger.Info("  - kapua-inventory-read: Read general inventory for a device")
	s.logger.Info("  - kapua-inventory-bundles: List bundle inventory for a device")
	s.logger.Info("  - kapua-inventory-bundle-start: Trigger bundle inventory start")
	s.logger.Info("  - kapua-inventory-bundle-stop: Trigger bundle inventory stop")
	s.logger.Info("  - kapua-inventory-containers: List container inventory for a device")
	s.logger.Info("  - kapua-inventory-container-start: Trigger container inventory start")
	s.logger.Info("  - kapua-inventory-container-stop: Trigger container inventory stop")
	s.logger.Info("  - kapua-inventory-system-packages: List system packages for a device")
	s.logger.Info("  - kapua-inventory-deployment-packages: List deployment packages for a device")
}

func registerKapuaTools(server *mcpsdk.Server, kapuaHandler *handlers.KapuaHandler) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-list-devices",
		Description: "List Kapua IoT devices",
	}, kapuaHandler.HandleListDevices)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-list-device-events",
		Description: "Read Kapua device events",
	}, kapuaHandler.HandleListDeviceEvents)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-list-device-logs",
		Description: "List Kapua device logs",
	}, kapuaHandler.HandleListDeviceLogs)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-list-data-messages",
		Description: "List Kapua data messages",
	}, kapuaHandler.HandleListDataMessages)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-configurations-read",
		Description: "Read all configurations for a Kapua device (input: {id})",
	}, kapuaHandler.HandleDeviceConfigurationsRead)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-inventory-read",
		Description: "Read general inventory for a Kapua device",
	}, kapuaHandler.HandleDeviceInventoryRead)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-inventory-bundles",
		Description: "List bundle inventory entries for a Kapua device",
	}, kapuaHandler.HandleDeviceInventoryBundles)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-inventory-bundle-start",
		Description: "Request a bundle inventory start for a Kapua device",
	}, kapuaHandler.HandleDeviceInventoryBundleStart)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-inventory-bundle-stop",
		Description: "Request a bundle inventory stop for a Kapua device",
	}, kapuaHandler.HandleDeviceInventoryBundleStop)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-inventory-containers",
		Description: "List container inventory entries for a Kapua device",
	}, kapuaHandler.HandleDeviceInventoryContainers)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-inventory-container-start",
		Description: "Request a container inventory start for a Kapua device",
	}, kapuaHandler.HandleDeviceInventoryContainerStart)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-inventory-container-stop",
		Description: "Request a container inventory stop for a Kapua device",
	}, kapuaHandler.HandleDeviceInventoryContainerStop)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-inventory-system-packages",
		Description: "List system packages inventory for a Kapua device",
	}, kapuaHandler.HandleDeviceInventorySystemPackages)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-inventory-deployment-packages",
		Description: "List deployment packages inventory for a Kapua device",
	}, kapuaHandler.HandleDeviceInventoryDeploymentPackages)
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
