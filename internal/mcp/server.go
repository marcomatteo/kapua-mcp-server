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
	registerKapuaPrompts(sdkServer)

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

	mcpHandler := newOriginMiddleware(httpCfg, logger, streamHandler)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle("/", mcpHandler)

	return mux
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
		Description: "Request an OSGi bundle inventory start operation on a Kapua device. Requires deviceId and a bundle descriptor object. This is an asynchronous remote operation that triggers an inventory scan for the specified bundle.",
	}, kapuaHandler.HandleDeviceInventoryBundleStart)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-inventory-bundle-stop",
		Description: "Request an OSGi bundle inventory stop operation on a Kapua device. Requires deviceId and a bundle descriptor object. This is an asynchronous remote operation that stops an inventory scan for the specified bundle.",
	}, kapuaHandler.HandleDeviceInventoryBundleStop)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-inventory-containers-list",
		Description: "List container inventory entries for a Kapua device. Requires deviceId. Returns container name, version, type, and state (ACTIVE/INSTALLED/UNINSTALLED/UNKNOWN).",
	}, kapuaHandler.HandleDeviceInventoryContainers)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-inventory-container-start",
		Description: "Request a container inventory start operation on a Kapua device. Requires deviceId and a container descriptor object. This is an asynchronous remote operation that triggers an inventory scan for the specified container.",
	}, kapuaHandler.HandleDeviceInventoryContainerStart)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-inventory-container-stop",
		Description: "Request a container inventory stop operation on a Kapua device. Requires deviceId and a container descriptor object. This is an asynchronous remote operation that stops an inventory scan for the specified container.",
	}, kapuaHandler.HandleDeviceInventoryContainerStop)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-inventory-system-packages-list",
		Description: "List system packages installed on a Kapua device. Requires deviceId. Returns package name, version, and type from the device OS inventory.",
	}, kapuaHandler.HandleDeviceInventorySystemPackages)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-inventory-deployment-packages-list",
		Description: "List deployment packages installed on a Kapua device. Requires deviceId. Returns deployment package metadata including name, version, and contained bundles.",
	}, kapuaHandler.HandleDeviceInventoryDeploymentPackages)

	// Device command execution
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-command-execute",
		Description: "Execute a command on a Kapua device. Requires deviceId and a command payload object following the Kapua command spec. Returns the command execution result.",
	}, kapuaHandler.HandleDeviceCommandExecute)

	// Device assets
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-assets-list",
		Description: "List asset channel definitions for a Kapua device. Requires deviceId. Returns the device's asset model with channels and their types.",
	}, kapuaHandler.HandleDeviceAssetsList)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-assets-read",
		Description: "Read current asset channel values from a Kapua device. Requires deviceId and a request payload specifying which assets/channels to read.",
	}, kapuaHandler.HandleDeviceAssetsRead)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-assets-write",
		Description: "Write values to asset channels on a Kapua device. Requires deviceId and a values payload specifying which assets/channels to write.",
	}, kapuaHandler.HandleDeviceAssetsWrite)

	// Legacy bundle management (by bundle ID)
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-bundles-list",
		Description: "List OSGi bundles installed on a Kapua device. Requires deviceId. Returns bundle details including ID, name, version, and state.",
	}, kapuaHandler.HandleDeviceBundlesList)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-bundle-start",
		Description: "Start an OSGi bundle on a Kapua device by bundle ID. Requires deviceId and bundleId.",
	}, kapuaHandler.HandleDeviceBundleStart)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-bundle-stop",
		Description: "Stop an OSGi bundle on a Kapua device by bundle ID. Requires deviceId and bundleId.",
	}, kapuaHandler.HandleDeviceBundleStop)

	// Configuration write operations
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-configurations-write",
		Description: "Write all configuration components to a Kapua device. Requires a device reference and a configurations payload object. This is a mutating operation.",
	}, kapuaHandler.HandleDeviceConfigurationsWrite)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-component-configuration-read",
		Description: "Read a single OSGi component configuration from a Kapua device. Requires a device reference and componentId.",
	}, kapuaHandler.HandleDeviceComponentConfigurationRead)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kapua-device-component-configuration-write",
		Description: "Write a single OSGi component configuration on a Kapua device. Requires a device reference, componentId, and a configuration payload. This is a mutating operation.",
	}, kapuaHandler.HandleDeviceComponentConfigurationWrite)
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

func registerKapuaPrompts(server *mcpsdk.Server) {
	server.AddPrompt(&mcpsdk.Prompt{
		Name:        "diagnose-device",
		Title:       "Diagnose Device",
		Description: "Diagnose the health and status of a specific IoT device",
		Arguments: []*mcpsdk.PromptArgument{
			{Name: "deviceId", Title: "Device ID", Description: "The Kapua device ID to diagnose", Required: true},
		},
	}, func(ctx context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
		deviceID := req.Params.Arguments["deviceId"]
		return &mcpsdk.GetPromptResult{
			Description: "Diagnose device " + deviceID,
			Messages: []*mcpsdk.PromptMessage{
				{
					Role: "user",
					Content: &mcpsdk.TextContent{
						Text: fmt.Sprintf(`Diagnose the health and status of Kapua device %s. Follow these steps:

1. Use kapua-devices-list to find the device and check its connection status, firmware version, and last event time.
2. Use kapua-device-events-list to review recent lifecycle events (connection changes, errors, application updates).
3. Use kapua-device-configurations-read to inspect the current device configuration for anomalies.
4. Use kapua-device-inventory-read to check installed software (bundles, packages, containers).
5. Use kapua-device-logs-list to look for recent error or warning log entries.

Summarize your findings with:
- Overall device health status (healthy, degraded, or critical)
- Key issues found (if any)
- Actionable recommendations for remediation`, deviceID),
					},
				},
			},
		}, nil
	})

	server.AddPrompt(&mcpsdk.Prompt{
		Name:        "fleet-overview",
		Title:       "Fleet Overview",
		Description: "Get a comprehensive overview of the IoT device fleet",
	}, func(ctx context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
		return &mcpsdk.GetPromptResult{
			Description: "IoT fleet overview",
			Messages: []*mcpsdk.PromptMessage{
				{
					Role: "user",
					Content: &mcpsdk.TextContent{
						Text: `Provide a comprehensive overview of the IoT device fleet. Follow these steps:

1. Use kapua-devices-list to retrieve all devices. Note the total count and pagination if needed.
2. Categorize devices by connection status (CONNECTED, DISCONNECTED, MISSING).
3. Identify firmware version distribution across the fleet.
4. Flag stale devices (DISCONNECTED or MISSING for extended periods).
5. If concerning devices are found, use kapua-device-events-list to check their recent activity.

Produce a fleet summary including:
- Total device count and connection status breakdown
- Firmware version distribution
- Devices needing attention (with reasons)
- Overall fleet health assessment`,
					},
				},
			},
		}, nil
	})

	server.AddPrompt(&mcpsdk.Prompt{
		Name:        "security-audit",
		Title:       "Security Audit",
		Description: "Perform a security audit on IoT devices",
		Arguments: []*mcpsdk.PromptArgument{
			{Name: "deviceId", Title: "Device ID", Description: "Optional device ID to audit a single device; omit for fleet-wide audit", Required: false},
		},
	}, func(ctx context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
		deviceID := req.Params.Arguments["deviceId"]
		scope := "all devices in the fleet"
		if deviceID != "" {
			scope = "device " + deviceID
		}
		return &mcpsdk.GetPromptResult{
			Description: "Security audit for " + scope,
			Messages: []*mcpsdk.PromptMessage{
				{
					Role: "user",
					Content: &mcpsdk.TextContent{
						Text: fmt.Sprintf(`Perform a security audit on %s. Follow these steps:

1. Use kapua-devices-list to identify target device(s) and their connection details.
2. Use kapua-device-configurations-read to review security-relevant configuration settings (authentication, encryption, firewall rules, exposed services).
3. Use kapua-device-inventory-bundles-list and kapua-device-inventory-deployment-packages-list to check for outdated or vulnerable software versions.
4. Use kapua-device-snapshots-list to verify configuration backup practices.
5. Use kapua-device-events-list to look for suspicious activity patterns (unexpected restarts, configuration changes, failed auth attempts).

Produce a security assessment with:
- Risk level (low, medium, high, critical) for each finding
- Specific vulnerabilities or misconfigurations found
- Prioritized remediation recommendations
- Compliance observations (if applicable)`, scope),
					},
				},
			},
		}, nil
	})

	server.AddPrompt(&mcpsdk.Prompt{
		Name:        "troubleshoot-connectivity",
		Title:       "Troubleshoot Connectivity",
		Description: "Troubleshoot connectivity issues for a specific device",
		Arguments: []*mcpsdk.PromptArgument{
			{Name: "deviceId", Title: "Device ID", Description: "The Kapua device ID experiencing connectivity issues", Required: true},
		},
	}, func(ctx context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
		deviceID := req.Params.Arguments["deviceId"]
		return &mcpsdk.GetPromptResult{
			Description: "Troubleshoot connectivity for device " + deviceID,
			Messages: []*mcpsdk.PromptMessage{
				{
					Role: "user",
					Content: &mcpsdk.TextContent{
						Text: fmt.Sprintf(`Troubleshoot connectivity issues for Kapua device %s. Follow these steps:

1. Use kapua-devices-list with a filter for this device to check its current connection status, last event time, and connection IP.
2. Use kapua-device-events-list to trace recent connection/disconnection events and identify patterns (intermittent drops, clean disconnects, timeouts).
3. Use kapua-device-logs-list to search for network-related errors, timeout messages, or authentication failures.
4. Use kapua-device-configurations-read to examine network and connection configuration (broker endpoints, keep-alive intervals, retry settings).
5. If the device is connected, use kapua-device-inventory-read to verify the communication stack is intact.

Provide a connectivity diagnosis with:
- Current connection state and history
- Root cause analysis (network, configuration, authentication, or device-side issue)
- Step-by-step remediation plan
- Preventive measures to avoid recurrence`, deviceID),
					},
				},
			},
		}, nil
	})
}
