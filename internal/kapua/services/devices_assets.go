package services

import (
	"context"
	"encoding/json"
	"net/http"
)

// Assets-related device APIs

// ListDeviceAssets lists asset definitions for a device
func (c *KapuaClient) ListDeviceAssets(ctx context.Context, deviceID string) (json.RawMessage, error) {
	endpoint := c.scopedEndpoint("/devices/%s/assets", deviceID)
	var out json.RawMessage
	if err := c.doKapuaRequest(ctx, http.MethodGet, endpoint, "list device assets", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ReadDeviceAssets reads current values for one or more assets
// request payload should follow Kapua spec for assets/_read
func (c *KapuaClient) ReadDeviceAssets(ctx context.Context, deviceID string, request any) (json.RawMessage, error) {
	endpoint := c.scopedEndpoint("/devices/%s/assets/_read", deviceID)
	var out json.RawMessage
	if err := c.doKapuaRequest(ctx, http.MethodPost, endpoint, "read device assets", request, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// WriteDeviceAssets writes values for one or more assets
// values payload should follow Kapua spec for assets/_write
func (c *KapuaClient) WriteDeviceAssets(ctx context.Context, deviceID string, values any) (json.RawMessage, error) {
	endpoint := c.scopedEndpoint("/devices/%s/assets/_write", deviceID)
	var out json.RawMessage
	if err := c.doKapuaRequest(ctx, http.MethodPost, endpoint, "write device assets", values, &out); err != nil {
		return nil, err
	}
	return out, nil
}
