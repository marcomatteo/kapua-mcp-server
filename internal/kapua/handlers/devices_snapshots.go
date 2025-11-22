package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DeviceSnapshotsParams struct {
	DeviceID string `json:"deviceId" jsonschema:"The device ID"`
}

type DeviceSnapshotLookupParams struct {
	DeviceID   string `json:"deviceId" jsonschema:"The device ID"`
	SnapshotID string `json:"snapshotId" jsonschema:"The snapshot ID"`
}

// HandleDeviceSnapshotsList lists available snapshots for a device and returns both
// a quick summary and the raw Kapua payload to the MCP client.
func (h *KapuaHandler) HandleDeviceSnapshotsList(ctx context.Context, req *mcp.CallToolRequest, params *DeviceSnapshotsParams) (*mcp.CallToolResult, any, error) {
	if params == nil || params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required")
	}
	h.logger.Info("Listing snapshots for device %s", params.DeviceID)
	snapshots, err := h.client.ListDeviceSnapshots(ctx, params.DeviceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list device snapshots: %w", err)
	}
	bytes, _ := json.Marshal(snapshots)
	snapshotCount := 0
	if snapshots != nil {
		snapshotCount = len(snapshots.SnapshotID)
	}
	summary := fmt.Sprintf("Retrieved %d snapshots", snapshotCount)
	return &mcp.CallToolResult{Content: []mcp.Content{
		&mcp.TextContent{Text: summary},
		&mcp.TextContent{Text: string(bytes)},
	}}, snapshots, nil
}

// HandleDeviceSnapshotConfigurationsRead returns the configuration payload for a given snapshot.
func (h *KapuaHandler) HandleDeviceSnapshotConfigurationsRead(ctx context.Context, req *mcp.CallToolRequest, params *DeviceSnapshotLookupParams) (*mcp.CallToolResult, any, error) {
	if params == nil || params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required")
	}
	if params.SnapshotID == "" {
		return nil, nil, fmt.Errorf("snapshotId is required")
	}
	h.logger.Info("Reading snapshot %s for device %s", params.SnapshotID, params.DeviceID)
	conf, err := h.client.ReadDeviceSnapshotConfigurations(ctx, params.DeviceID, params.SnapshotID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read device snapshot configurations: %w", err)
	}
	bytes, _ := json.Marshal(conf)
	configCount := 0
	if conf != nil {
		configCount = len(conf.Configuration)
	}
	summary := fmt.Sprintf("Retrieved %d component configurations from snapshot %s", configCount, params.SnapshotID)
	return &mcp.CallToolResult{Content: []mcp.Content{
		&mcp.TextContent{Text: summary},
		&mcp.TextContent{Text: string(bytes)},
	}}, conf, nil
}

// HandleDeviceSnapshotRollback triggers a rollback to the provided snapshot on the device.
func (h *KapuaHandler) HandleDeviceSnapshotRollback(ctx context.Context, req *mcp.CallToolRequest, params *DeviceSnapshotLookupParams) (*mcp.CallToolResult, any, error) {
	if params == nil || params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required")
	}
	if params.SnapshotID == "" {
		return nil, nil, fmt.Errorf("snapshotId is required")
	}
	h.logger.Info("Requesting rollback of device %s to snapshot %s", params.DeviceID, params.SnapshotID)
	if err := h.client.RollbackDeviceSnapshot(ctx, params.DeviceID, params.SnapshotID); err != nil {
		return nil, nil, fmt.Errorf("failed to rollback device snapshot: %w", err)
	}
	summary := fmt.Sprintf("Rollback to snapshot %s requested for device %s", params.SnapshotID, params.DeviceID)
	return &mcp.CallToolResult{Content: []mcp.Content{
		&mcp.TextContent{Text: summary},
	}}, map[string]string{"status": "requested"}, nil
}
