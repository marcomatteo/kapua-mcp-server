package services

import (
	"context"
	"fmt"
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

	endpoint := fmt.Sprintf("/%s/devices/%s/events", c.scopeId, deviceID)
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("list device events request failed: %w", err)
	}

	var result models.DeviceEventListResult
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to list device events: %w", err)
	}

	c.logger.Info("Retrieved %d device events", len(result.Items))
	return &result, nil
}
