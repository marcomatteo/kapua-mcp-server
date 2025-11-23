package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Asset tools

type DeviceAssetsListParams struct {
	DeviceID string `json:"deviceId" jsonschema:"The device ID to inspect assets"`
}

func (h *KapuaHandler) HandleDeviceAssetsList(ctx context.Context, req *mcp.CallToolRequest, params *DeviceAssetsListParams) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Listing assets for device %s", params.DeviceID)
	out, err := h.client.ListDeviceAssets(ctx, params.DeviceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list device assets: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(out)}}}, json.RawMessage(out), nil
}

type DeviceAssetsReadParams struct {
	DeviceID string         `json:"deviceId" jsonschema:"The device ID to read assets from"`
	Request  map[string]any `json:"request" jsonschema:"Assets read request as object"`
}

func (h *KapuaHandler) HandleDeviceAssetsRead(ctx context.Context, req *mcp.CallToolRequest, params *DeviceAssetsReadParams) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Reading assets for device %s", params.DeviceID)
	out, err := h.client.ReadDeviceAssets(ctx, params.DeviceID, params.Request)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read device assets: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(out)}}}, json.RawMessage(out), nil
}

type DeviceAssetsWriteParams struct {
	DeviceID string         `json:"deviceId" jsonschema:"The device ID to write assets to"`
	Values   map[string]any `json:"values" jsonschema:"Assets write values as object"`
}

func (h *KapuaHandler) HandleDeviceAssetsWrite(ctx context.Context, req *mcp.CallToolRequest, params *DeviceAssetsWriteParams) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Writing assets for device %s", params.DeviceID)
	out, err := h.client.WriteDeviceAssets(ctx, params.DeviceID, params.Values)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write device assets: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(out)}}}, json.RawMessage(out), nil
}
