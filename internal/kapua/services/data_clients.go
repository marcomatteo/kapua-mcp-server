package services

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"kapua-mcp-server/internal/kapua/models"
)

// DataMessagesQuery captures filter options supported by Kapua's /data/messages endpoint.
type DataMessagesQuery struct {
	ClientIDs     []string
	Channel       string
	StrictChannel *bool
	StartDate     string
	EndDate       string
	SortDir       string
	Limit         *int
	Offset        *int
}

func (q *DataMessagesQuery) toValues() url.Values {
	values := url.Values{}
	if q == nil {
		return values
	}

	for _, clientID := range q.ClientIDs {
		if clientID != "" {
			values.Add("clientId", clientID)
		}
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

// ListDataMessages queries the Kapua Data Message API within the current scope.
func (c *KapuaClient) ListDataMessages(ctx context.Context, query *DataMessagesQuery) (*models.DataMessageListResult, error) {
	c.logger.Info("Listing data messages for scope: %s", c.scopeId)

	params := query.toValues()

	endpoint := fmt.Sprintf("/%s/data/messages", c.scopeId)
	if encoded := params.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("list data messages request failed: %w", err)
	}

	var result models.DataMessageListResult
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to list data messages: %w", err)
	}

	c.logger.Info("Listed %d data messages successfully", len(result.Items))
	return &result, nil
}
