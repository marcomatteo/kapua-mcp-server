package mcp

import (
	"context"
	"fmt"
	"net/http"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"kapua-mcp-server/internal/kapua/config"
	"kapua-mcp-server/internal/kapua/handlers"
	"kapua-mcp-server/internal/kapua/services"
	"kapua-mcp-server/pkg/utils"
)

type Server struct {
	logger    *utils.Logger
	kapuaCfg  *config.Config
	mcpServer *mcpsdk.Server
}

var kapuaClientFactory = services.NewKapuaClient

func NewServer(ctx context.Context, kapuaCfg *config.Config) (*Server, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if kapuaCfg == nil {
		return nil, fmt.Errorf("kapua configuration cannot be nil")
	}

	logger := utils.NewDefaultLogger("MCPServer")
	logger.Info("Starting Kapua MCP Server")

	kapuaClient := kapuaClientFactory(&kapuaCfg.Kapua)

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

	return &Server{
		logger:    logger,
		kapuaCfg:  kapuaCfg,
		mcpServer: sdkServer,
	}, nil
}

func (s *Server) Handler(httpCfg *HTTPConfig) http.Handler {

	streamHandler := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server {
		return s.mcpServer
	}, nil)

	logger := s.logger
	if logger == nil {
		logger = utils.NewDefaultLogger("MCPServer")
	}

	return newOriginMiddleware(httpCfg, logger, streamHandler)
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
	if s.kapuaCfg != nil {
		s.logger.Info("Kapua API endpoint: %s", s.kapuaCfg.Kapua.APIEndpoint)
	}
}

func registerKapuaTools(server *mcpsdk.Server, kapuaHandler *handlers.KapuaHandler) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-devices-list",
		Description: "List Kapua IoT devices with optional filters for client ID, connection status (CONNECTED/DISCONNECTED/MISSING/NULL), and free-text search. Supports pagination via limit and offset. Returns device metadata including connection state, firmware, and OS info.",
	}, kapuaHandler.HandleListDevices)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-events-list",
		Description: "List lifecycle events for a Kapua device (requires deviceId). Filter by resource type, date range, and sort order. Returns timestamped events such as connection changes, command executions, and application updates.",
	}, kapuaHandler.HandleListDeviceEvents)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-logs-list",
		Description: "List device log entries stored in the Kapua datastore. Filter by clientId, channel, date range, and log property values. Returns structured log records with timestamps and metric payloads. Supports pagination.",
	}, kapuaHandler.HandleListDeviceLogs)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-data-messages-list",
		Description: "List telemetry data messages stored in the Kapua datastore. Filter by one or more clientIds, channel, and date range. Returns message payloads with channel, timestamp, and metric values. Supports pagination.",
	}, kapuaHandler.HandleListDataMessages)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-configurations-read",
		Description: "Read all OSGi configuration components currently active on a Kapua device. Requires deviceId. Returns the full set of component configurations with their properties and values.",
	}, kapuaHandler.HandleDeviceConfigurationsRead)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-snapshots-list",
		Description: "List available configuration snapshots for a Kapua device. Requires deviceId. Returns snapshot IDs that can be used with kapua-device-snapshot-configurations-read or kapua-device-snapshot-rollback.",
	}, kapuaHandler.HandleDeviceSnapshotsList)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-snapshot-configurations-read",
		Description: "Read the component configurations stored in a specific device snapshot. Requires deviceId and snapshotId. Use kapua-device-snapshots-list first to discover available snapshot IDs.",
	}, kapuaHandler.HandleDeviceSnapshotConfigurationsRead)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-snapshot-rollback",
		Description: "Trigger a configuration rollback on a Kapua device to a previously saved snapshot. Requires deviceId and snapshotId. This is a mutating operation that restores the device configuration to the snapshot state.",
	}, kapuaHandler.HandleDeviceSnapshotRollback)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-inventory-read",
		Description: "Read the general software inventory for a Kapua device. Requires deviceId. Returns all inventory items (bundles, packages, containers) with name, version, and type.",
	}, kapuaHandler.HandleDeviceInventoryRead)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-inventory-bundles-list",
		Description: "List OSGi bundle inventory entries for a Kapua device. Requires deviceId. Returns bundle ID, name, version, status (ACTIVE/RESOLVED/INSTALLED/etc.), and signed flag.",
	}, kapuaHandler.HandleDeviceInventoryBundles)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-inventory-bundle-start",
		Description: "Request an OSGi bundle to be started on a Kapua device. Requires deviceId and a bundle descriptor object. This is an asynchronous remote operation.",
	}, kapuaHandler.HandleDeviceInventoryBundleStart)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-inventory-bundle-stop",
		Description: "Request an OSGi bundle to be stopped on a Kapua device. Requires deviceId and a bundle descriptor object. This is an asynchronous remote operation.",
	}, kapuaHandler.HandleDeviceInventoryBundleStop)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-inventory-containers-list",
		Description: "List container inventory entries for a Kapua device. Requires deviceId. Returns container name, version, type, and state (ACTIVE/INSTALLED/UNINSTALLED/UNKNOWN).",
	}, kapuaHandler.HandleDeviceInventoryContainers)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-inventory-container-start",
		Description: "Request a container to be started on a Kapua device. Requires deviceId and a container descriptor object. This is an asynchronous remote operation.",
	}, kapuaHandler.HandleDeviceInventoryContainerStart)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-inventory-container-stop",
		Description: "Request a container to be stopped on a Kapua device. Requires deviceId and a container descriptor object. This is an asynchronous remote operation.",
	}, kapuaHandler.HandleDeviceInventoryContainerStop)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-inventory-system-packages-list",
		Description: "List system packages installed on a Kapua device. Requires deviceId. Returns package name, version, and type from the device OS inventory.",
	}, kapuaHandler.HandleDeviceInventorySystemPackages)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-inventory-deployment-packages-list",
		Description: "List deployment packages installed on a Kapua device. Requires deviceId. Returns deployment package metadata including name, version, and contained bundles.",
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

	server.AddResource(&mcpsdk.Resource{
		URI:         "kapua://fleet-health",
		Name:        "Kapua Fleet Health",
		Description: "Aggregated fleet health snapshot",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcpsdk.ReadResourceRequest) (*mcpsdk.ReadResourceResult, error) {
		return kapuaHandler.ReadResource(ctx, req.Params.URI)
	})
}
