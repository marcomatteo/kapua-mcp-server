package services

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

// Assets-related device APIs

// ListDeviceAssets lists asset definitions for a device
func (c *KapuaClient) ListDeviceAssets(ctx context.Context, deviceID string) (json.RawMessage, error) {
    endpoint := fmt.Sprintf("/%s/devices/%s/assets", c.scopeId, deviceID)
    resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
    if err != nil {
        return nil, fmt.Errorf("list device assets request failed: %w", err)
    }
    var out json.RawMessage
    if err := c.handleResponse(resp, &out); err != nil {
        return nil, fmt.Errorf("failed to list device assets: %w", err)
    }
    return out, nil
}

// ReadDeviceAssets reads current values for one or more assets
// request payload should follow Kapua spec for assets/_read
func (c *KapuaClient) ReadDeviceAssets(ctx context.Context, deviceID string, request any) (json.RawMessage, error) {
    endpoint := fmt.Sprintf("/%s/devices/%s/assets/_read", c.scopeId, deviceID)
    resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, request)
    if err != nil {
        return nil, fmt.Errorf("read device assets request failed: %w", err)
    }
    var out json.RawMessage
    if err := c.handleResponse(resp, &out); err != nil {
        return nil, fmt.Errorf("failed to read device assets: %w", err)
    }
    return out, nil
}

// WriteDeviceAssets writes values for one or more assets
// values payload should follow Kapua spec for assets/_write
func (c *KapuaClient) WriteDeviceAssets(ctx context.Context, deviceID string, values any) (json.RawMessage, error) {
    endpoint := fmt.Sprintf("/%s/devices/%s/assets/_write", c.scopeId, deviceID)
    resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, values)
    if err != nil {
        return nil, fmt.Errorf("write device assets request failed: %w", err)
    }
    var out json.RawMessage
    if err := c.handleResponse(resp, &out); err != nil {
        return nil, fmt.Errorf("failed to write device assets: %w", err)
    }
    return out, nil
}

