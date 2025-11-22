package services

import (
	"context"
	"net/http"

	"kapua-mcp-server/internal/kapua/models"
)

// Inventory-related device APIs

// ReadDeviceInventory retrieves the general inventory information for a device.
func (c *KapuaClient) ReadDeviceInventory(ctx context.Context, deviceID string) (*models.DeviceInventory, error) {
	var out models.DeviceInventory
	endpoint := c.scopedEndpoint("/devices/%s/inventory", deviceID)
	if err := c.doKapuaRequest(ctx, http.MethodGet, endpoint, "read device inventory", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListDeviceInventoryBundles retrieves bundle-specific inventory information for a device.
func (c *KapuaClient) ListDeviceInventoryBundles(ctx context.Context, deviceID string) (*models.DeviceInventoryBundles, error) {
	var out models.DeviceInventoryBundles
	endpoint := c.scopedEndpoint("/devices/%s/inventory/bundles", deviceID)
	if err := c.doKapuaRequest(ctx, http.MethodGet, endpoint, "list device inventory bundles", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// StartDeviceInventoryBundle triggers an inventory refresh for bundles.
func (c *KapuaClient) StartDeviceInventoryBundle(ctx context.Context, deviceID string, bundle models.DeviceInventoryBundle) error {
	endpoint := c.scopedEndpoint("/devices/%s/inventory/bundles/_start", deviceID)
	return c.doKapuaRequest(ctx, http.MethodPost, endpoint, "start device inventory bundle", bundle, nil)
}

// StopDeviceInventoryBundle stops an ongoing bundle inventory operation.
func (c *KapuaClient) StopDeviceInventoryBundle(ctx context.Context, deviceID string, bundle models.DeviceInventoryBundle) error {
	endpoint := c.scopedEndpoint("/devices/%s/inventory/bundles/_stop", deviceID)
	return c.doKapuaRequest(ctx, http.MethodPost, endpoint, "stop device inventory bundle", bundle, nil)
}

// ListDeviceInventoryContainers retrieves container inventory information for a device.
func (c *KapuaClient) ListDeviceInventoryContainers(ctx context.Context, deviceID string) (*models.DeviceInventoryContainers, error) {
	var out models.DeviceInventoryContainers
	endpoint := c.scopedEndpoint("/devices/%s/inventory/containers", deviceID)
	if err := c.doKapuaRequest(ctx, http.MethodGet, endpoint, "list device inventory containers", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// StartDeviceInventoryContainer triggers an inventory refresh for containers.
func (c *KapuaClient) StartDeviceInventoryContainer(ctx context.Context, deviceID string, container models.DeviceInventoryContainer) error {
	endpoint := c.scopedEndpoint("/devices/%s/inventory/containers/_start", deviceID)
	return c.doKapuaRequest(ctx, http.MethodPost, endpoint, "start device inventory container", container, nil)
}

// StopDeviceInventoryContainer stops an ongoing container inventory operation.
func (c *KapuaClient) StopDeviceInventoryContainer(ctx context.Context, deviceID string, container models.DeviceInventoryContainer) error {
	endpoint := c.scopedEndpoint("/devices/%s/inventory/containers/_stop", deviceID)
	return c.doKapuaRequest(ctx, http.MethodPost, endpoint, "stop device inventory container", container, nil)
}

// ListDeviceInventorySystemPackages retrieves system package inventory for a device.
func (c *KapuaClient) ListDeviceInventorySystemPackages(ctx context.Context, deviceID string) (*models.DeviceInventoryPackages, error) {
	var out models.DeviceInventoryPackages
	endpoint := c.scopedEndpoint("/devices/%s/inventory/system", deviceID)
	if err := c.doKapuaRequest(ctx, http.MethodGet, endpoint, "list device inventory system packages", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListDeviceInventoryDeploymentPackages retrieves deployment package inventory for a device.
func (c *KapuaClient) ListDeviceInventoryDeploymentPackages(ctx context.Context, deviceID string) (*models.DeviceInventoryDeploymentPackages, error) {
	var out models.DeviceInventoryDeploymentPackages
	endpoint := c.scopedEndpoint("/devices/%s/inventory/packages", deviceID)
	if err := c.doKapuaRequest(ctx, http.MethodGet, endpoint, "list device inventory deployment packages", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
