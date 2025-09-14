package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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

// DeviceManagementParams defines parameters for device management operations
type DeviceManagementParams struct {
	Operation   string `json:"operation" jsonschema:"The operation to perform (list/get/create/update/delete)"`
	ScopeID     string `json:"scopeId" jsonschema:"The scope ID to operate on"`
	DeviceID    string `json:"deviceId,omitempty" jsonschema:"The device ID (required for get/update/delete operations)"`
	ClientID    string `json:"clientId,omitempty" jsonschema:"The client ID filter for list or required for create"`
	Status      string `json:"status,omitempty" jsonschema:"The device status filter (ENABLED/DISABLED) for list or value for create/update"`
	DisplayName string `json:"displayName,omitempty" jsonschema:"The display name for create/update operations"`
	MatchTerm   string `json:"matchTerm,omitempty" jsonschema:"Search term to match against device fields for list operation"`
	Limit       int    `json:"limit,omitempty" jsonschema:"Maximum number of devices to return for list operation (default: 50)"`
	Offset      int    `json:"offset,omitempty" jsonschema:"Number of devices to skip for list operation (default: 0)"`
}

// MCP Tool Handlers

// HandleDeviceManagement handles device management operations
func (h *KapuaHandler) HandleDeviceManagement(ctx context.Context, req *mcp.CallToolRequest, params *DeviceManagementParams) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Processing device management operation: %s for scope: %s", params.Operation, params.ScopeID)

	switch strings.ToLower(params.Operation) {
	case "list":
		return h.handleListDevices(ctx, params)
	case "get":
		return h.handleGetDevice(ctx, params)
	case "create":
		return h.handleCreateDevice(ctx, params)
	case "update":
		return h.handleUpdateDevice(ctx, params)
	case "delete":
		return h.handleDeleteDevice(ctx, params)
	default:
		return nil, nil, fmt.Errorf("unsupported operation: %s. Supported operations: list, get, create, update, delete", params.Operation)
	}
}

// Device Management Sub-handlers

func (h *KapuaHandler) handleListDevices(ctx context.Context, params *DeviceManagementParams) (*mcp.CallToolResult, any, error) {
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
	} else {
		queryParams["limit"] = "50" // Default limit
	}
	if params.Offset > 0 {
		queryParams["offset"] = strconv.Itoa(params.Offset)
	}

	result, err := h.client.ListDevices(ctx, params.ScopeID, queryParams)
	if err != nil {
		h.logger.Error("List devices failed: %v", err)
		return nil, nil, fmt.Errorf("failed to list devices: %w", err)
	}

	response := fmt.Sprintf("Found %d devices in scope %s:\n\n", len(result.Items), params.ScopeID)
	for i, device := range result.Items {
		if i >= 20 { // Limit output to avoid overwhelming response
			response += fmt.Sprintf("... and %d more devices\n", len(result.Items)-20)
			break
		}
		response += fmt.Sprintf("Device %d:\n", i+1)
		response += fmt.Sprintf("  ID: %s\n", device.ID)
		response += fmt.Sprintf("  Client ID: %s\n", device.ClientID)
		response += fmt.Sprintf("  Status: %s\n", device.Status)
		if device.DisplayName != "" {
			response += fmt.Sprintf("  Display Name: %s\n", device.DisplayName)
		}
		if device.SerialNumber != "" {
			response += fmt.Sprintf("  Serial Number: %s\n", device.SerialNumber)
		}
		if device.ModelName != "" {
			response += fmt.Sprintf("  Model: %s\n", device.ModelName)
		}
		response += fmt.Sprintf("  Created: %s\n\n", device.CreatedOn.Format("2006-01-02 15:04:05 MST"))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: response},
		},
	}, nil, nil
}

func (h *KapuaHandler) handleGetDevice(ctx context.Context, params *DeviceManagementParams) (*mcp.CallToolResult, any, error) {
	if params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required for get operation")
	}

	device, err := h.client.GetDevice(ctx, params.ScopeID, params.DeviceID)
	if err != nil {
		h.logger.Error("Get device failed: %v", err)
		return nil, nil, fmt.Errorf("failed to get device: %w", err)
	}

	response := fmt.Sprintf("Device Details:\n")
	response += fmt.Sprintf("ID: %s\n", device.ID)
	response += fmt.Sprintf("Client ID: %s\n", device.ClientID)
	response += fmt.Sprintf("Status: %s\n", device.Status)
	response += fmt.Sprintf("Display Name: %s\n", device.DisplayName)
	response += fmt.Sprintf("Serial Number: %s\n", device.SerialNumber)
	response += fmt.Sprintf("Model ID: %s\n", device.ModelID)
	response += fmt.Sprintf("Model Name: %s\n", device.ModelName)
	response += fmt.Sprintf("BIOS Version: %s\n", device.BiosVersion)
	response += fmt.Sprintf("Firmware Version: %s\n", device.FirmwareVersion)
	response += fmt.Sprintf("OS Version: %s\n", device.OsVersion)
	response += fmt.Sprintf("JVM Version: %s\n", device.JvmVersion)
	response += fmt.Sprintf("Connection Interface: %s\n", device.ConnectionInterface)
	response += fmt.Sprintf("Connection IP: %s\n", device.ConnectionIP)
	response += fmt.Sprintf("Created: %s\n", device.CreatedOn.Format("2006-01-02 15:04:05 MST"))
	response += fmt.Sprintf("Modified: %s\n", device.ModifiedOn.Format("2006-01-02 15:04:05 MST"))

	if len(device.ExtendedProperties) > 0 {
		response += "\nExtended Properties:\n"
		for _, prop := range device.ExtendedProperties {
			response += fmt.Sprintf("  %s - %s: %s\n", prop.GroupName, prop.Name, prop.Value)
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: response},
		},
	}, nil, nil
}

func (h *KapuaHandler) handleCreateDevice(ctx context.Context, params *DeviceManagementParams) (*mcp.CallToolResult, any, error) {
	if params.ClientID == "" {
		return nil, nil, fmt.Errorf("clientId is required for create operation")
	}

	creator := models.DeviceCreator{
		ClientID:    params.ClientID,
		DisplayName: params.DisplayName,
	}

	if params.Status != "" {
		creator.Status = models.DeviceStatus(params.Status)
	} else {
		creator.Status = models.DeviceStatusEnabled // Default to enabled
	}

	device, err := h.client.CreateDevice(ctx, params.ScopeID, creator)
	if err != nil {
		h.logger.Error("Create device failed: %v", err)
		return nil, nil, fmt.Errorf("failed to create device: %w", err)
	}

	response := fmt.Sprintf("Device created successfully!\n")
	response += fmt.Sprintf("ID: %s\n", device.ID)
	response += fmt.Sprintf("Client ID: %s\n", device.ClientID)
	response += fmt.Sprintf("Status: %s\n", device.Status)
	response += fmt.Sprintf("Display Name: %s\n", device.DisplayName)
	response += fmt.Sprintf("Created: %s\n", device.CreatedOn.Format("2006-01-02 15:04:05 MST"))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: response},
		},
	}, nil, nil
}

func (h *KapuaHandler) handleUpdateDevice(ctx context.Context, params *DeviceManagementParams) (*mcp.CallToolResult, any, error) {
	if params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required for update operation")
	}

	// First get the existing device
	existingDevice, err := h.client.GetDevice(ctx, params.ScopeID, params.DeviceID)
	if err != nil {
		h.logger.Error("Get device for update failed: %v", err)
		return nil, nil, fmt.Errorf("failed to get device for update: %w", err)
	}

	// Update fields if provided
	if params.DisplayName != "" {
		existingDevice.DisplayName = params.DisplayName
	}
	if params.Status != "" {
		existingDevice.Status = models.DeviceStatus(params.Status)
	}

	updatedDevice, err := h.client.UpdateDevice(ctx, params.ScopeID, params.DeviceID, *existingDevice)
	if err != nil {
		h.logger.Error("Update device failed: %v", err)
		return nil, nil, fmt.Errorf("failed to update device: %w", err)
	}

	response := fmt.Sprintf("Device updated successfully!\n")
	response += fmt.Sprintf("ID: %s\n", updatedDevice.ID)
	response += fmt.Sprintf("Client ID: %s\n", updatedDevice.ClientID)
	response += fmt.Sprintf("Status: %s\n", updatedDevice.Status)
	response += fmt.Sprintf("Display Name: %s\n", updatedDevice.DisplayName)
	response += fmt.Sprintf("Modified: %s\n", updatedDevice.ModifiedOn.Format("2006-01-02 15:04:05 MST"))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: response},
		},
	}, nil, nil
}

func (h *KapuaHandler) handleDeleteDevice(ctx context.Context, params *DeviceManagementParams) (*mcp.CallToolResult, any, error) {
	if params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required for delete operation")
	}

	err := h.client.DeleteDevice(ctx, params.ScopeID, params.DeviceID)
	if err != nil {
		h.logger.Error("Delete device failed: %v", err)
		return nil, nil, fmt.Errorf("failed to delete device: %w", err)
	}

	response := fmt.Sprintf("Device %s deleted successfully from scope %s", params.DeviceID, params.ScopeID)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: response},
		},
	}, nil, nil
}
