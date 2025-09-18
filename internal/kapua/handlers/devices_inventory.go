package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"kapua-mcp-server/internal/kapua/models"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DeviceInventoryParams struct {
	DeviceID string `json:"deviceId" jsonschema:"The device ID"`
}

func (h *KapuaHandler) HandleDeviceInventoryRead(ctx context.Context, req *mcp.CallToolRequest, params *DeviceInventoryParams) (*mcp.CallToolResult, any, error) {
	if params == nil || params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required")
	}
	h.logger.Info("Reading inventory for device %s", params.DeviceID)
	inv, err := h.client.ReadDeviceInventory(ctx, params.DeviceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read device inventory: %w", err)
	}
	bytes, _ := json.Marshal(inv)
	summary := fmt.Sprintf("Retrieved %d inventory items", len(inv.InventoryItems))
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: summary}, &mcp.TextContent{Text: string(bytes)}}}, inv, nil
}

func (h *KapuaHandler) HandleDeviceInventoryBundles(ctx context.Context, req *mcp.CallToolRequest, params *DeviceInventoryParams) (*mcp.CallToolResult, any, error) {
	if params == nil || params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required")
	}
	h.logger.Info("Listing inventory bundles for device %s", params.DeviceID)
	inv, err := h.client.ListDeviceInventoryBundles(ctx, params.DeviceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list device inventory bundles: %w", err)
	}
	bytes, _ := json.Marshal(inv)
	summary := fmt.Sprintf("Retrieved %d inventory bundles", len(inv.InventoryBundles))
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: summary}, &mcp.TextContent{Text: string(bytes)}}}, inv, nil
}

type DeviceInventoryBundleActionParams struct {
	DeviceID string                       `json:"deviceId" jsonschema:"The device ID"`
	Bundle   models.DeviceInventoryBundle `json:"bundle" jsonschema:"Bundle descriptor"`
}

func (h *KapuaHandler) HandleDeviceInventoryBundleStart(ctx context.Context, req *mcp.CallToolRequest, params *DeviceInventoryBundleActionParams) (*mcp.CallToolResult, any, error) {
	if params == nil || params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required")
	}
	h.logger.Info("Starting inventory bundle scan for device %s", params.DeviceID)
	if err := h.client.StartDeviceInventoryBundle(ctx, params.DeviceID, params.Bundle); err != nil {
		return nil, nil, fmt.Errorf("failed to start device inventory bundle: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Bundle inventory start requested"}}}, map[string]string{"status": "requested"}, nil
}

func (h *KapuaHandler) HandleDeviceInventoryBundleStop(ctx context.Context, req *mcp.CallToolRequest, params *DeviceInventoryBundleActionParams) (*mcp.CallToolResult, any, error) {
	if params == nil || params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required")
	}
	h.logger.Info("Stopping inventory bundle scan for device %s", params.DeviceID)
	if err := h.client.StopDeviceInventoryBundle(ctx, params.DeviceID, params.Bundle); err != nil {
		return nil, nil, fmt.Errorf("failed to stop device inventory bundle: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Bundle inventory stop requested"}}}, map[string]string{"status": "requested"}, nil
}

func (h *KapuaHandler) HandleDeviceInventoryContainers(ctx context.Context, req *mcp.CallToolRequest, params *DeviceInventoryParams) (*mcp.CallToolResult, any, error) {
	if params == nil || params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required")
	}
	h.logger.Info("Listing inventory containers for device %s", params.DeviceID)
	inv, err := h.client.ListDeviceInventoryContainers(ctx, params.DeviceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list device inventory containers: %w", err)
	}
	bytes, _ := json.Marshal(inv)
	summary := fmt.Sprintf("Retrieved %d inventory containers", len(inv.InventoryContainers))
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: summary}, &mcp.TextContent{Text: string(bytes)}}}, inv, nil
}

type DeviceInventoryContainerActionParams struct {
	DeviceID  string                          `json:"deviceId" jsonschema:"The device ID"`
	Container models.DeviceInventoryContainer `json:"container" jsonschema:"Container descriptor"`
}

func (h *KapuaHandler) HandleDeviceInventoryContainerStart(ctx context.Context, req *mcp.CallToolRequest, params *DeviceInventoryContainerActionParams) (*mcp.CallToolResult, any, error) {
	if params == nil || params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required")
	}
	h.logger.Info("Starting inventory container scan for device %s", params.DeviceID)
	if err := h.client.StartDeviceInventoryContainer(ctx, params.DeviceID, params.Container); err != nil {
		return nil, nil, fmt.Errorf("failed to start device inventory container: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Container inventory start requested"}}}, map[string]string{"status": "requested"}, nil
}

func (h *KapuaHandler) HandleDeviceInventoryContainerStop(ctx context.Context, req *mcp.CallToolRequest, params *DeviceInventoryContainerActionParams) (*mcp.CallToolResult, any, error) {
	if params == nil || params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required")
	}
	h.logger.Info("Stopping inventory container scan for device %s", params.DeviceID)
	if err := h.client.StopDeviceInventoryContainer(ctx, params.DeviceID, params.Container); err != nil {
		return nil, nil, fmt.Errorf("failed to stop device inventory container: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Container inventory stop requested"}}}, map[string]string{"status": "requested"}, nil
}

func (h *KapuaHandler) HandleDeviceInventorySystemPackages(ctx context.Context, req *mcp.CallToolRequest, params *DeviceInventoryParams) (*mcp.CallToolResult, any, error) {
	if params == nil || params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required")
	}
	h.logger.Info("Listing system packages for device %s", params.DeviceID)
	inv, err := h.client.ListDeviceInventorySystemPackages(ctx, params.DeviceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list device system packages: %w", err)
	}
	bytes, _ := json.Marshal(inv)
	summary := fmt.Sprintf("Retrieved %d system packages", len(inv.SystemPackages))
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: summary}, &mcp.TextContent{Text: string(bytes)}}}, inv, nil
}

func (h *KapuaHandler) HandleDeviceInventoryDeploymentPackages(ctx context.Context, req *mcp.CallToolRequest, params *DeviceInventoryParams) (*mcp.CallToolResult, any, error) {
	if params == nil || params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required")
	}
	h.logger.Info("Listing deployment packages for device %s", params.DeviceID)
	inv, err := h.client.ListDeviceInventoryDeploymentPackages(ctx, params.DeviceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list device deployment packages: %w", err)
	}
	bytes, _ := json.Marshal(inv)
	count := len(inv.DeploymentPackages)
	if count == 0 {
		count = len(inv.SystemPackages)
	}
	summary := fmt.Sprintf("Retrieved %d deployment packages", count)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: summary}, &mcp.TextContent{Text: string(bytes)}}}, inv, nil
}
