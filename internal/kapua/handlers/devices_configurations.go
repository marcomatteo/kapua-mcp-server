package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"kapua-mcp-server/internal/kapua/models"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DeviceIDArg is a minimal object wrapper to satisfy MCP input schema (must be type object)
type DeviceID struct {
	DeviceID string `json:"deviceId" jsonschema:"The device ID"`
}

// HandleDeviceConfigurationsRead reads all configurations for a device by deviceId only
func (h *KapuaHandler) HandleDeviceConfigurationsRead(ctx context.Context, req *mcp.CallToolRequest, args *DeviceID) (*mcp.CallToolResult, any, error) {
	if args == nil || args.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required")
	}
	h.logger.Info("Reading configurations for device %s", args.DeviceID)
	//dev := models.Device{KapuaEntity: models.KapuaEntity{ID: models.KapuaID(args.DeviceID)}}
	conf, err := h.client.ReadDeviceConfigurations(ctx, args.DeviceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read device configurations: %w", err)
	}
	bytes, _ := json.Marshal(conf)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(bytes)}}}, conf, nil
}

type DeviceConfigurationsWriteParams struct {
	Device  models.Device  `json:"device" jsonschema:"Device reference"`
	Payload map[string]any `json:"payload" jsonschema:"Configurations write payload as object"`
}

func (h *KapuaHandler) HandleDeviceConfigurationsWrite(ctx context.Context, req *mcp.CallToolRequest, params *DeviceConfigurationsWriteParams) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Writing configurations for device %s", params.Device.ID)
	if err := h.client.WriteDeviceConfigurations(ctx, params.Device, params.Payload); err != nil {
		return nil, nil, fmt.Errorf("failed to write device configurations: %w", err)
	}
	summary := fmt.Sprintf("Updated configurations for device %s", params.Device.ID)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: summary}}}, map[string]string{"status": "updated"}, nil
}

type DeviceComponentConfigurationReadParams struct {
	Device      models.Device `json:"device" jsonschema:"Device reference"`
	ComponentID string        `json:"componentId" jsonschema:"The component ID"`
}

func (h *KapuaHandler) HandleDeviceComponentConfigurationRead(ctx context.Context, req *mcp.CallToolRequest, params *DeviceComponentConfigurationReadParams) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Reading component configuration %s for device %s", params.ComponentID, params.Device.ID)
	conf, err := h.client.ReadDeviceComponentConfiguration(ctx, params.Device, params.ComponentID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read device component configuration: %w", err)
	}
	bytes, _ := json.Marshal(conf)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(bytes)}}}, conf, nil
}

type DeviceComponentConfigurationWriteParams struct {
	Device      models.Device  `json:"device" jsonschema:"Device reference"`
	ComponentID string         `json:"componentId" jsonschema:"The component ID"`
	Payload     map[string]any `json:"payload" jsonschema:"Component configuration write payload as object"`
}

func (h *KapuaHandler) HandleDeviceComponentConfigurationWrite(ctx context.Context, req *mcp.CallToolRequest, params *DeviceComponentConfigurationWriteParams) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Writing component configuration %s for device %s", params.ComponentID, params.Device.ID)
	if err := h.client.WriteDeviceComponentConfiguration(ctx, params.Device, params.ComponentID, params.Payload); err != nil {
		return nil, nil, fmt.Errorf("failed to write device component configuration: %w", err)
	}
	summary := fmt.Sprintf("Updated component %s configuration for device %s", params.ComponentID, params.Device.ID)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: summary}}}, map[string]string{"status": "updated"}, nil
}
