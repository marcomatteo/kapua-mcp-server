package services

import (
	"context"
	"fmt"
	"net/http"

	"kapua-mcp-server/internal/kapua/models"
)

// Configurations-related device APIs

func (c *KapuaClient) ReadDeviceConfigurations(ctx context.Context, deviceId string) (*models.DeviceConfiguration, error) {
	endpoint := fmt.Sprintf("/%s/devices/%s/configurations", c.scopeId, deviceId)
	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("read device configurations request failed: %w", err)
	}
	var out models.DeviceConfiguration
	if err := c.handleResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to read device configurations: %w", err)
	}
	return &out, nil
}

func (c *KapuaClient) WriteDeviceConfigurations(ctx context.Context, device models.Device, payload any) error {
	endpoint := fmt.Sprintf("/%s/devices/%s/configurations", c.scopeId, device.ID)
	resp, err := c.makeRequest(ctx, http.MethodPut, endpoint, payload)
	if err != nil {
		return fmt.Errorf("write device configurations request failed: %w", err)
	}
	if err := c.handleResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to write device configurations: %w", err)
	}
	return nil
}

func (c *KapuaClient) ReadDeviceComponentConfiguration(ctx context.Context, device models.Device, componentID string) (*models.DeviceConfiguration, error) {
	endpoint := fmt.Sprintf("/%s/devices/%s/configurations/%s", c.scopeId, device.ID, componentID)
	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("read device component configuration request failed: %w", err)
	}
	var out models.DeviceConfiguration
	if err := c.handleResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to read device component configuration: %w", err)
	}
	return &out, nil
}

func (c *KapuaClient) WriteDeviceComponentConfiguration(ctx context.Context, device models.Device, componentID string, payload any) error {
	endpoint := fmt.Sprintf("/%s/devices/%s/configurations/%s", c.scopeId, device.ID, componentID)
	resp, err := c.makeRequest(ctx, http.MethodPut, endpoint, payload)
	if err != nil {
		return fmt.Errorf("write device component configuration request failed: %w", err)
	}
	if err := c.handleResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to write device component configuration: %w", err)
	}
	return nil
}
