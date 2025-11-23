package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DeviceBundlesListParams struct {
	DeviceID string `json:"deviceId" jsonschema:"The device ID to list bundles"`
}

func (h *KapuaHandler) HandleDeviceBundlesList(ctx context.Context, req *mcp.CallToolRequest, params *DeviceBundlesListParams) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Listing bundles for device %s", params.DeviceID)
	out, err := h.client.ListDeviceBundles(ctx, params.DeviceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list device bundles: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(out)}}}, string(out), nil
}

type DeviceBundleActionParams struct {
	DeviceID string `json:"deviceId" jsonschema:"The device ID"`
	BundleID string `json:"bundleId" jsonschema:"The bundle ID"`
}

func (h *KapuaHandler) HandleDeviceBundleStart(ctx context.Context, req *mcp.CallToolRequest, params *DeviceBundleActionParams) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Starting bundle %s on device %s", params.BundleID, params.DeviceID)
	if err := h.client.StartDeviceBundle(ctx, params.DeviceID, params.BundleID); err != nil {
		return nil, nil, fmt.Errorf("failed to start device bundle: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Bundle start requested"}}}, map[string]string{"status": "requested"}, nil
}

func (h *KapuaHandler) HandleDeviceBundleStop(ctx context.Context, req *mcp.CallToolRequest, params *DeviceBundleActionParams) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Stopping bundle %s on device %s", params.BundleID, params.DeviceID)
	if err := h.client.StopDeviceBundle(ctx, params.DeviceID, params.BundleID); err != nil {
		return nil, nil, fmt.Errorf("failed to stop device bundle: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Bundle stop requested"}}}, map[string]string{"status": "requested"}, nil
}
