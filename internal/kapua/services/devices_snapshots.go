package services

import (
	"context"
	"net/http"

	"kapua-mcp-server/internal/kapua/models"
)

// ListDeviceSnapshots retrieves the available snapshots for the given device.
func (c *KapuaClient) ListDeviceSnapshots(ctx context.Context, deviceID string) (*models.DeviceSnapshots, error) {
	var out models.DeviceSnapshots
	endpoint := c.scopedEndpoint("/devices/%s/snapshots", deviceID)
	if err := c.doKapuaRequest(ctx, http.MethodGet, endpoint, "list device snapshots", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ReadDeviceSnapshotConfigurations retrieves the configuration snapshot identified by snapshotID.
func (c *KapuaClient) ReadDeviceSnapshotConfigurations(ctx context.Context, deviceID, snapshotID string) (*models.DeviceConfiguration, error) {
	var out models.DeviceConfiguration
	endpoint := c.scopedEndpoint("/devices/%s/snapshots/%s", deviceID, snapshotID)
	if err := c.doKapuaRequest(ctx, http.MethodGet, endpoint, "read device snapshot configurations", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RollbackDeviceSnapshot applies the given snapshot on the target device.
func (c *KapuaClient) RollbackDeviceSnapshot(ctx context.Context, deviceID, snapshotID string) error {
	endpoint := c.scopedEndpoint("/devices/%s/snapshots/%s/_rollback", deviceID, snapshotID)
	return c.doKapuaRequest(ctx, http.MethodPost, endpoint, "rollback device snapshot", nil, nil)
}
