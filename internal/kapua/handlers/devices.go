package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"kapua-mcp-server/internal/kapua/models"
)

const deviceResourcePageSize = 200

// Device Management Tool Parameters

// ListDevicesParams defines parameters for listing devices
type ListDevicesParams struct {
	ClientID         string                  `json:"clientId,omitempty" jsonschema:"Filter devices by client ID"`
	ConnectionStatus models.ConnectionStatus `json:"status,omitempty" jsonschema:"Filter devices by connection status (CONNECTED/DISCONNECTED/MISSING/NULL)"`
	MatchTerm        string                  `json:"matchTerm,omitempty" jsonschema:"Search term to match against device fields"`
	Limit            int                     `json:"limit,omitempty" jsonschema:"Maximum number of devices to return (default: 50)"`
	Offset           int                     `json:"offset,omitempty" jsonschema:"Number of devices to skip (default: 0)"`
}

// CreateDeviceParams defines parameters for creating a device
type CreateDeviceParams struct {
	// Device payload as a generic object to avoid schema tag constraints
	Device map[string]any `json:"device" jsonschema:"Device creation payload as object"`
}

// UpdateDeviceParams defines parameters for updating a device
type UpdateDeviceParams struct {
	DeviceID string `json:"deviceId" jsonschema:"The device ID to update"`
	// Device payload as a generic object to avoid schema tag constraints
	Device map[string]any `json:"device" jsonschema:"Updated device payload as object"`
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
	if params.ConnectionStatus != "" {
		queryParams["status"] = string(params.ConnectionStatus)
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

// readDevicesResource returns all devices as a JSON resource
func (h *KapuaHandler) readDevicesResource(ctx context.Context, uri *url.URL) (*mcp.ReadResourceResult, error) {
	limitParam := 0
	if uri != nil {
		if parsedLimit, err := strconv.Atoi(uri.Query().Get("limit")); err == nil && parsedLimit > 0 {
			limitParam = parsedLimit
		}
	}

	var devices []models.Device
	offset := 0
	totalCount := 0
	targetCount := limitParam // <=0 means fetch all

	for {

		pageSize := deviceResourcePageSize
		if targetCount > 0 {
			remaining := targetCount - len(devices)
			if remaining <= 0 {
				break
			}
			if remaining < pageSize {
				pageSize = remaining
			}
		}

		queryParams := map[string]string{
			"limit":         strconv.Itoa(pageSize),
			"offset":        strconv.Itoa(offset),
			"askTotalCount": "true",
		}

		result, err := h.client.ListDevices(ctx, queryParams)
		if err != nil {
			h.logger.Error("Failed to read devices resource: %v", err)
			return nil, fmt.Errorf("failed to read devices resource: %w", err)
		}

		if totalCount == 0 {
			totalCount = result.TotalCount
		}

		devices = append(devices, result.Items...)

		if len(result.Items) == 0 || len(result.Items) < pageSize {
			break
		}
		offset += len(result.Items)
	}

	if totalCount == 0 {
		totalCount = len(devices)
	}

	resourceData := map[string]interface{}{
		"total_count":     totalCount,
		"processed_count": len(devices),
		"devices":         devices,
		"last_updated":    fmt.Sprintf("%d", timeNow().Unix()),
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
