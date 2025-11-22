package services

import (
	"context"
	"fmt"
	"net/http"

	"kapua-mcp-server/internal/kapua/models"
)

// ListDeviceSnapshots retrieves the available snapshots for the given device.
func (c *KapuaClient) ListDeviceSnapshots(ctx context.Context, deviceID string) (*models.DeviceSnapshots, error) {
	endpoint := fmt.Sprintf("/%s/devices/%s/snapshots", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("list device snapshots request failed: %w", err)
	}
	var out models.DeviceSnapshots
	if err := c.handleResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to list device snapshots: %w", err)
	}
	return &out, nil
}

// ReadDeviceSnapshotConfigurations retrieves the configuration snapshot identified by snapshotID.
func (c *KapuaClient) ReadDeviceSnapshotConfigurations(ctx context.Context, deviceID, snapshotID string) (*models.DeviceConfiguration, error) {
	endpoint := fmt.Sprintf("/%s/devices/%s/snapshots/%s", c.scopeId, deviceID, snapshotID)
	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("read device snapshot configurations request failed: %w", err)
	}
	var out models.DeviceConfiguration
	if err := c.handleResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to read device snapshot configurations: %w", err)
	}
	return &out, nil
}

// RollbackDeviceSnapshot applies the given snapshot on the target device.
func (c *KapuaClient) RollbackDeviceSnapshot(ctx context.Context, deviceID, snapshotID string) error {
	endpoint := fmt.Sprintf("/%s/devices/%s/snapshots/%s/_rollback", c.scopeId, deviceID, snapshotID)
	resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return fmt.Errorf("rollback device snapshot request failed: %w", err)
	}
	if err := c.handleResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to rollback device snapshot: %w", err)
	}
	return nil
}
