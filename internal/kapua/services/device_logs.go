package services

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"kapua-mcp-server/internal/kapua/models"
)

// DeviceLogsQuery captures filter options for Kapua's /deviceLogs endpoint.
type DeviceLogsQuery struct {
	ClientID        string
	Channel         string
	StrictChannel   *bool
	StartDate       string
	EndDate         string
	LogPropertyName string
	LogPropertyType string
	LogPropertyMin  string
	LogPropertyMax  string
	SortDir         string
	Limit           *int
	Offset          *int
}

func (q *DeviceLogsQuery) toValues() url.Values {
	values := url.Values{}
	if q == nil {
		return values
	}

	if q.ClientID != "" {
		values.Set("clientId", q.ClientID)
	}
	if q.Channel != "" {
		values.Set("channel", q.Channel)
	}
	if q.StrictChannel != nil {
		values.Set("strictChannel", strconv.FormatBool(*q.StrictChannel))
	}
	if q.StartDate != "" {
		values.Set("startDate", q.StartDate)
	}
	if q.EndDate != "" {
		values.Set("endDate", q.EndDate)
	}
	if q.LogPropertyName != "" {
		values.Set("logPropertyName", q.LogPropertyName)
	}
	if q.LogPropertyType != "" {
		values.Set("logPropertyType", q.LogPropertyType)
	}
	if q.LogPropertyMin != "" {
		values.Set("logPropertyMin", q.LogPropertyMin)
	}
	if q.LogPropertyMax != "" {
		values.Set("logPropertyMax", q.LogPropertyMax)
	}
	if q.SortDir != "" {
		values.Set("sortDir", q.SortDir)
	}
	if q.Limit != nil {
		values.Set("limit", strconv.Itoa(*q.Limit))
	}
	if q.Offset != nil {
		values.Set("offset", strconv.Itoa(*q.Offset))
	}

	return values
}

// ListDeviceLogs queries the Kapua Device Logs endpoint within the current scope.
func (c *KapuaClient) ListDeviceLogs(ctx context.Context, query *DeviceLogsQuery) (*models.DeviceLogListResult, error) {
	c.logger.Info("Listing device logs for scope: %s", c.scopeId)

	params := query.toValues()

	endpoint := fmt.Sprintf("/%s/deviceLogs", c.scopeId)
	if encoded := params.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("list device logs request failed: %w", err)
	}

	var result models.DeviceLogListResult
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to list device logs: %w", err)
	}

	c.logger.Info("Retrieved %d device logs", len(result.Items))
	return &result, nil
}
