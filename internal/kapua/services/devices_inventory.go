package services

import (
	"context"
	"fmt"
	"net/http"

	"kapua-mcp-server/internal/kapua/models"
)

// Inventory-related device APIs

// ReadDeviceInventory retrieves the general inventory information for a device.
func (c *KapuaClient) ReadDeviceInventory(ctx context.Context, deviceID string) (*models.DeviceInventory, error) {
	endpoint := fmt.Sprintf("/%s/devices/%s/inventory", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("read device inventory request failed: %w", err)
	}
	var out models.DeviceInventory
	if err := c.handleResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to read device inventory: %w", err)
	}
	return &out, nil
}

// ListDeviceInventoryBundles retrieves bundle-specific inventory information for a device.
func (c *KapuaClient) ListDeviceInventoryBundles(ctx context.Context, deviceID string) (*models.DeviceInventoryBundles, error) {
	endpoint := fmt.Sprintf("/%s/devices/%s/inventory/bundles", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("list device inventory bundles request failed: %w", err)
	}
	var out models.DeviceInventoryBundles
	if err := c.handleResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to list device inventory bundles: %w", err)
	}
	return &out, nil
}

// StartDeviceInventoryBundle triggers an inventory refresh for bundles.
func (c *KapuaClient) StartDeviceInventoryBundle(ctx context.Context, deviceID string, bundle models.DeviceInventoryBundle) error {
	endpoint := fmt.Sprintf("/%s/devices/%s/inventory/bundles/_start", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, bundle)
	if err != nil {
		return fmt.Errorf("start device inventory bundle request failed: %w", err)
	}
	if err := c.handleResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to start device inventory bundle: %w", err)
	}
	return nil
}

// StopDeviceInventoryBundle stops an ongoing bundle inventory operation.
func (c *KapuaClient) StopDeviceInventoryBundle(ctx context.Context, deviceID string, bundle models.DeviceInventoryBundle) error {
	endpoint := fmt.Sprintf("/%s/devices/%s/inventory/bundles/_stop", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, bundle)
	if err != nil {
		return fmt.Errorf("stop device inventory bundle request failed: %w", err)
	}
	if err := c.handleResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to stop device inventory bundle: %w", err)
	}
	return nil
}

// ListDeviceInventoryContainers retrieves container inventory information for a device.
func (c *KapuaClient) ListDeviceInventoryContainers(ctx context.Context, deviceID string) (*models.DeviceInventoryContainers, error) {
	endpoint := fmt.Sprintf("/%s/devices/%s/inventory/containers", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("list device inventory containers request failed: %w", err)
	}
	var out models.DeviceInventoryContainers
	if err := c.handleResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to list device inventory containers: %w", err)
	}
	return &out, nil
}

// StartDeviceInventoryContainer triggers an inventory refresh for containers.
func (c *KapuaClient) StartDeviceInventoryContainer(ctx context.Context, deviceID string, container models.DeviceInventoryContainer) error {
	endpoint := fmt.Sprintf("/%s/devices/%s/inventory/containers/_start", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, container)
	if err != nil {
		return fmt.Errorf("start device inventory container request failed: %w", err)
	}
	if err := c.handleResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to start device inventory container: %w", err)
	}
	return nil
}

// StopDeviceInventoryContainer stops an ongoing container inventory operation.
func (c *KapuaClient) StopDeviceInventoryContainer(ctx context.Context, deviceID string, container models.DeviceInventoryContainer) error {
	endpoint := fmt.Sprintf("/%s/devices/%s/inventory/containers/_stop", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, container)
	if err != nil {
		return fmt.Errorf("stop device inventory container request failed: %w", err)
	}
	if err := c.handleResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to stop device inventory container: %w", err)
	}
	return nil
}

// ListDeviceInventorySystemPackages retrieves system package inventory for a device.
func (c *KapuaClient) ListDeviceInventorySystemPackages(ctx context.Context, deviceID string) (*models.DeviceInventoryPackages, error) {
	endpoint := fmt.Sprintf("/%s/devices/%s/inventory/system", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("list device inventory system packages request failed: %w", err)
	}
	var out models.DeviceInventoryPackages
	if err := c.handleResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to list device inventory system packages: %w", err)
	}
	return &out, nil
}

// ListDeviceInventoryDeploymentPackages retrieves deployment package inventory for a device.
func (c *KapuaClient) ListDeviceInventoryDeploymentPackages(ctx context.Context, deviceID string) (*models.DeviceInventoryDeploymentPackages, error) {
	endpoint := fmt.Sprintf("/%s/devices/%s/inventory/packages", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("list device inventory deployment packages request failed: %w", err)
	}
	var out models.DeviceInventoryDeploymentPackages
	if err := c.handleResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to list device inventory deployment packages: %w", err)
	}
	return &out, nil
}
