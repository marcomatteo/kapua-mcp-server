package services

import (
	"context"
	"net/http"

	"kapua-mcp-server/internal/kapua/models"
)

// Configurations-related device APIs

func (c *KapuaClient) ReadDeviceConfigurations(ctx context.Context, deviceId string) (*models.DeviceConfiguration, error) {
	var out models.DeviceConfiguration
	endpoint := c.scopedEndpoint("/devices/%s/configurations", deviceId)
	if err := c.doKapuaRequest(ctx, http.MethodGet, endpoint, "read device configurations", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *KapuaClient) WriteDeviceConfigurations(ctx context.Context, device models.Device, payload any) error {
	endpoint := c.scopedEndpoint("/devices/%s/configurations", device.ID)
	return c.doKapuaRequest(ctx, http.MethodPut, endpoint, "write device configurations", payload, nil)
}

func (c *KapuaClient) ReadDeviceComponentConfiguration(ctx context.Context, device models.Device, componentID string) (*models.DeviceConfiguration, error) {
	var out models.DeviceConfiguration
	endpoint := c.scopedEndpoint("/devices/%s/configurations/%s", device.ID, componentID)
	if err := c.doKapuaRequest(ctx, http.MethodGet, endpoint, "read device component configuration", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *KapuaClient) WriteDeviceComponentConfiguration(ctx context.Context, device models.Device, componentID string, payload any) error {
	endpoint := c.scopedEndpoint("/devices/%s/configurations/%s", device.ID, componentID)
	return c.doKapuaRequest(ctx, http.MethodPut, endpoint, "write device component configuration", payload, nil)
}
