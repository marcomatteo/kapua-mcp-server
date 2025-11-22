package services

import (
	"context"
	"encoding/json"
	"net/http"
)

// Bundles-related device APIs

// ListDeviceBundles lists bundles installed on a device
func (c *KapuaClient) ListDeviceBundles(ctx context.Context, deviceID string) (json.RawMessage, error) {
	endpoint := c.scopedEndpoint("/devices/%s/bundles", deviceID)
	var out json.RawMessage
	if err := c.doKapuaRequest(ctx, http.MethodGet, endpoint, "list device bundles", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// StartDeviceBundle starts a bundle by ID
func (c *KapuaClient) StartDeviceBundle(ctx context.Context, deviceID, bundleID string) error {
	endpoint := c.scopedEndpoint("/devices/%s/bundles/%s/_start", deviceID, bundleID)
	return c.doKapuaRequest(ctx, http.MethodPost, endpoint, "start device bundle", nil, nil)
}

// StopDeviceBundle stops a bundle by ID
func (c *KapuaClient) StopDeviceBundle(ctx context.Context, deviceID, bundleID string) error {
	endpoint := c.scopedEndpoint("/devices/%s/bundles/%s/_stop", deviceID, bundleID)
	return c.doKapuaRequest(ctx, http.MethodPost, endpoint, "stop device bundle", nil, nil)
}
