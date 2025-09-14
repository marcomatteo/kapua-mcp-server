package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"kapua-mcp-server/internal/models"
	"kapua-mcp-server/internal/services"
	"kapua-mcp-server/pkg/utils"
)

// KapuaHandler provides MCP tool handlers for Kapua operations
type KapuaHandler struct {
	client *services.KapuaClient
	logger *utils.Logger
}

// NewKapuaHandler creates a new Kapua handler
func NewKapuaHandler(client *services.KapuaClient) *KapuaHandler {
	return &KapuaHandler{
		client: client,
		logger: utils.NewDefaultLogger("KapuaHandler"),
	}
}

// Device Management Tool Parameters

// ListDevicesParams defines parameters for listing devices
type ListDevicesParams struct {
	ClientID  string `json:"clientId,omitempty" jsonschema:"Filter devices by client ID"`
	Status    string `json:"status,omitempty" jsonschema:"Filter devices by status (ENABLED/DISABLED)"`
	MatchTerm string `json:"matchTerm,omitempty" jsonschema:"Search term to match against device fields"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of devices to return (default: 50)"`
	Offset    int    `json:"offset,omitempty" jsonschema:"Number of devices to skip (default: 0)"`
}

// CreateDeviceParams defines parameters for creating a device
type CreateDeviceParams struct {
    // Device payload as a generic object to avoid schema tag constraints
    Device map[string]any `json:"device" jsonschema:"Device creation payload as object"`
}

// UpdateDeviceParams defines parameters for updating a device
type UpdateDeviceParams struct {
    DeviceID string                 `json:"deviceId" jsonschema:"The device ID to update"`
    // Device payload as a generic object to avoid schema tag constraints
    Device   map[string]any         `json:"device" jsonschema:"Updated device payload as object"`
}

// DeleteDeviceParams defines parameters for deleting a device
type DeleteDeviceParams struct {
	DeviceID string `json:"deviceId" jsonschema:"The device ID to delete"`
}

// MCP Tool Handlers

// HandleListDevices handles listing Kapua devices with structured JSON response
func (h *KapuaHandler) HandleListDevices(ctx context.Context, req *mcp.CallToolRequest, params *ListDevicesParams) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Listing devices:")

	// Build query parameters
	queryParams := make(map[string]string)
	if params.ClientID != "" {
		queryParams["clientId"] = params.ClientID
	}
	if params.Status != "" {
		queryParams["status"] = params.Status
	}
	if params.MatchTerm != "" {
		queryParams["matchTerm"] = params.MatchTerm
	}
	if params.Limit > 0 {
		queryParams["limit"] = strconv.Itoa(params.Limit)
	}

	if params.Offset > 0 {
		queryParams["offset"] = strconv.Itoa(params.Offset)
	}

	result, err := h.client.ListDevices(ctx, queryParams)
	if err != nil {
		h.logger.Error("List devices failed: %v", err)
		return nil, nil, fmt.Errorf("failed to list devices: %w", err)
	}

	// Return structured JSON data that LLMs can interpret
	jsonData, err := json.Marshal(result)
	if err != nil {
		h.logger.Error("Failed to marshal device list: %v", err)
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	summary := fmt.Sprintf("Found %d devices.", len(result.Items))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{
				Text: string(jsonData),
			},
		},
	}, nil, nil
}

// HandleCreateDevice handles creating a new Kapua device
func (h *KapuaHandler) HandleCreateDevice(ctx context.Context, req *mcp.CallToolRequest, params *CreateDeviceParams) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Creating device")

	var creator models.DeviceCreator
	// Convert generic map to strong typed creator
	if b, err := json.Marshal(params.Device); err != nil {
		return nil, nil, fmt.Errorf("failed to encode device payload: %w", err)
	} else if err := json.Unmarshal(b, &creator); err != nil {
        return nil, nil, fmt.Errorf("invalid device payload: %w", err)
    }

	device, err := h.client.CreateDevice(ctx, creator)
	if err != nil {
		h.logger.Error("Create device failed: %v", err)
		return nil, nil, fmt.Errorf("failed to create device: %w", err)
	}

	jsonData, err := json.Marshal(device)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal created device: %w", err)
	}

	summary := fmt.Sprintf("Created device %s (ID: %s)", device.ClientID, device.ID)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(jsonData)},
		},
	}, device, nil
}

// HandleUpdateDevice handles updating an existing Kapua device
func (h *KapuaHandler) HandleUpdateDevice(ctx context.Context, req *mcp.CallToolRequest, params *UpdateDeviceParams) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Updating device %s", params.DeviceID)

	var payload models.Device
	if b, err := json.Marshal(params.Device); err != nil {
		return nil, nil, fmt.Errorf("failed to encode device payload: %w", err)
	} else if err := json.Unmarshal(b, &payload); err != nil {
		return nil, nil, fmt.Errorf("invalid device payload: %w", err)
	}

	updated, err := h.client.UpdateDevice(ctx, params.DeviceID, payload)
	if err != nil {
		h.logger.Error("Update device failed: %v", err)
		return nil, nil, fmt.Errorf("failed to update device: %w", err)
	}

	jsonData, err := json.Marshal(updated)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal updated device: %w", err)
	}

	summary := fmt.Sprintf("Updated device %s (ID: %s)", updated.ClientID, updated.ID)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(jsonData)},
		},
	}, updated, nil
}

// HandleDeleteDevice handles deleting a Kapua device
func (h *KapuaHandler) HandleDeleteDevice(ctx context.Context, req *mcp.CallToolRequest, params *DeleteDeviceParams) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Deleting device %s", params.DeviceID)

	if err := h.client.DeleteDevice(ctx, params.DeviceID); err != nil {
		h.logger.Error("Delete device failed: %v", err)
		return nil, nil, fmt.Errorf("failed to delete device: %w", err)
	}

	summary := fmt.Sprintf("Deleted device ID %s", params.DeviceID)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
		},
	}, map[string]string{"status": "deleted", "deviceId": params.DeviceID}, nil
}

// MCP Resource Handlers

// ListResources returns a list of available Kapua resources
func (h *KapuaHandler) ListResources(ctx context.Context) ([]mcp.Resource, error) {
	h.logger.Debug("Listing available Kapua resources")

	resources := []mcp.Resource{
		{
			URI:         "kapua://devices",
			Name:        "Kapua Devices",
			Description: "Live list of all Kapua IoT devices with current status and metadata",
			MIMEType:    "application/json",
		},
	}

	return resources, nil
}

// ReadResource returns the content of a specific Kapua resource
func (h *KapuaHandler) ReadResource(ctx context.Context, uri string) (*mcp.ReadResourceResult, error) {
	h.logger.Debug("Reading Kapua resource: %s", uri)

	switch uri {
	case "kapua://devices":
		return h.readDevicesResource(ctx)
	default:
		return nil, fmt.Errorf("unknown resource URI: %s", uri)
	}
}

// readDevicesResource returns all devices as a JSON resource
func (h *KapuaHandler) readDevicesResource(ctx context.Context) (*mcp.ReadResourceResult, error) {
	// Get all devices with reasonable defaults
	queryParams := map[string]string{
		"limit": "100", // Reasonable limit for resource view
	}

	result, err := h.client.ListDevices(ctx, queryParams)
	if err != nil {
		h.logger.Error("Failed to read devices resource: %v", err)
		return nil, fmt.Errorf("failed to read devices resource: %w", err)
	}

	// Create a structured resource response
	resourceData := map[string]interface{}{
		"total_count":  len(result.Items),
		"devices":      result.Items,
		"last_updated": fmt.Sprintf("%d", time.Now().Unix()),
	}

	jsonData, err := json.MarshalIndent(resourceData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal devices resource: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      "kapua://devices",
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}
