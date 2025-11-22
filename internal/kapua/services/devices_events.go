package services

import (
	"context"
	"net/http"
	"net/url"

	"kapua-mcp-server/internal/kapua/models"
)

// Events-related device APIs

// ListDeviceEvents retrieves device log messages (events) for the specified device.
// Optional query parameters should match the Kapua API specification, e.g. resource, startDate, endDate, etc.
func (c *KapuaClient) ListDeviceEvents(ctx context.Context, deviceID string, params map[string]string) (*models.DeviceEventListResult, error) {
	c.logger.Info("Listing device events for device %s in scope: %s", deviceID, c.scopeId)

	query := url.Values{}
	for key, value := range params {
		if value != "" {
			query.Set(key, value)
		}
	}

	endpoint := c.scopedEndpoint("/devices/%s/events", deviceID)
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	var result models.DeviceEventListResult
	if err := c.doKapuaRequest(ctx, http.MethodGet, endpoint, "list device events", nil, &result); err != nil {
		return nil, err
	}

	c.logger.Info("Retrieved %d device events", len(result.Items))
	return &result, nil
}
