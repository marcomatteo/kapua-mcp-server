package services

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

// Bundles-related device APIs

// ListDeviceBundles lists bundles installed on a device
func (c *KapuaClient) ListDeviceBundles(ctx context.Context, deviceID string) (json.RawMessage, error) {
    endpoint := fmt.Sprintf("/%s/devices/%s/bundles", c.scopeId, deviceID)
    resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
    if err != nil {
        return nil, fmt.Errorf("list device bundles request failed: %w", err)
    }
    var out json.RawMessage
    if err := c.handleResponse(resp, &out); err != nil {
        return nil, fmt.Errorf("failed to list device bundles: %w", err)
    }
    return out, nil
}

// StartDeviceBundle starts a bundle by ID
func (c *KapuaClient) StartDeviceBundle(ctx context.Context, deviceID, bundleID string) error {
    endpoint := fmt.Sprintf("/%s/devices/%s/bundles/%s/_start", c.scopeId, deviceID, bundleID)
    resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, nil)
    if err != nil {
        return fmt.Errorf("start device bundle request failed: %w", err)
    }
    return c.handleResponse(resp, nil)
}

// StopDeviceBundle stops a bundle by ID
func (c *KapuaClient) StopDeviceBundle(ctx context.Context, deviceID, bundleID string) error {
    endpoint := fmt.Sprintf("/%s/devices/%s/bundles/%s/_stop", c.scopeId, deviceID, bundleID)
    resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, nil)
    if err != nil {
        return fmt.Errorf("stop device bundle request failed: %w", err)
    }
    return c.handleResponse(resp, nil)
}

