package services

import (
	"context"
	"kapua-mcp-server/internal/kapua/models"
	"net/http"
	"net/url"
)

// Device Management Methods

// ListDevices retrieves a list of devices from a scope
func (c *KapuaClient) ListDevices(ctx context.Context, params map[string]string) (*models.DeviceListResult, error) {
	c.logger.Info("Listing devices for scope: %s", c.scopeId)

	// Build query parameters
	queryParams := url.Values{}
	for key, value := range params {
		if value != "" {
			queryParams.Set(key, value)
		}
	}

	endpoint := c.scopedEndpoint("/devices")
	if len(queryParams) > 0 {
		endpoint += "?" + queryParams.Encode()
	}

	var result models.DeviceListResult
	if err := c.doKapuaRequest(ctx, http.MethodGet, endpoint, "list devices", nil, &result); err != nil {
		return nil, err
	}

	c.logger.Info("Listed %d devices successfully", len(result.Items))
	return &result, nil
}

// GetDevice retrieves a specific device by ID
func (c *KapuaClient) GetDevice(ctx context.Context, deviceID string) (*models.Device, error) {
	c.logger.Info("Getting device %s from scope: %s", deviceID, c.scopeId)

	var device models.Device
	endpoint := c.scopedEndpoint("/devices/%s", deviceID)
	if err := c.doKapuaRequest(ctx, http.MethodGet, endpoint, "get device", nil, &device); err != nil {
		return nil, err
	}

	c.logger.Info("Device retrieved successfully: %s", device.ClientID)
	return &device, nil
}

// UpdateDevice updates an existing device
func (c *KapuaClient) UpdateDevice(ctx context.Context, deviceID string, device models.Device) (*models.Device, error) {
	c.logger.Info("Updating device %s in scope: %s", deviceID, c.scopeId)

	var updatedDevice models.Device
	endpoint := c.scopedEndpoint("/devices/%s", deviceID)
	if err := c.doKapuaRequest(ctx, http.MethodPut, endpoint, "update device", device, &updatedDevice); err != nil {
		return nil, err
	}

	c.logger.Info("Device updated successfully: %s", updatedDevice.ClientID)
	return &updatedDevice, nil
}

// DeleteDevice deletes a device
func (c *KapuaClient) DeleteDevice(ctx context.Context, deviceID string) error {
	c.logger.Info("Deleting device %s from scope: %s", deviceID, c.scopeId)

	endpoint := c.scopedEndpoint("/devices/%s", deviceID)
	if err := c.doKapuaRequest(ctx, http.MethodDelete, endpoint, "delete device", nil, nil); err != nil {
		return err
	}

	c.logger.Info("Device deleted successfully")
	return nil
}
