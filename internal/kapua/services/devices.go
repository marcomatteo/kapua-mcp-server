package services

import (
	"context"
	"fmt"
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

	endpoint := fmt.Sprintf("/%s/devices", c.scopeId)
	if len(queryParams) > 0 {
		endpoint += "?" + queryParams.Encode()
	}

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("list devices request failed: %w", err)
	}

	var result models.DeviceListResult
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	c.logger.Info("Listed %d devices successfully", len(result.Items))
	return &result, nil
}

// GetDevice retrieves a specific device by ID
func (c *KapuaClient) GetDevice(ctx context.Context, deviceID string) (*models.Device, error) {
	c.logger.Info("Getting device %s from scope: %s", deviceID, c.scopeId)

	endpoint := fmt.Sprintf("/%s/devices/%s", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("get device request failed: %w", err)
	}

	var device models.Device
	if err := c.handleResponse(resp, &device); err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	c.logger.Info("Device retrieved successfully: %s", device.ClientID)
	return &device, nil
}

// UpdateDevice updates an existing device
func (c *KapuaClient) UpdateDevice(ctx context.Context, deviceID string, device models.Device) (*models.Device, error) {
	c.logger.Info("Updating device %s in scope: %s", deviceID, c.scopeId)

	endpoint := fmt.Sprintf("/%s/devices/%s", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodPut, endpoint, device)
	if err != nil {
		return nil, fmt.Errorf("update device request failed: %w", err)
	}

	var updatedDevice models.Device
	if err := c.handleResponse(resp, &updatedDevice); err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	c.logger.Info("Device updated successfully: %s", updatedDevice.ClientID)
	return &updatedDevice, nil
}

// DeleteDevice deletes a device
func (c *KapuaClient) DeleteDevice(ctx context.Context, deviceID string) error {
	c.logger.Info("Deleting device %s from scope: %s", deviceID, c.scopeId)

	endpoint := fmt.Sprintf("/%s/devices/%s", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("delete device request failed: %w", err)
	}

	if err := c.handleResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	c.logger.Info("Device deleted successfully")
	return nil
}
