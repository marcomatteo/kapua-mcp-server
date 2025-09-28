package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListDeviceEventsParams defines parameters for listing device events (logs)
// for a Kapua device.
type ListDeviceEventsParams struct {
	DeviceID      string `json:"deviceId" jsonschema:"The Kapua device ID to read events for"`
	Resource      string `json:"resource,omitempty" jsonschema:"Filter events by resource (e.g. LOG)"`
	StartDate     string `json:"startDate,omitempty" jsonschema:"Filter events created on or after this RFC3339 timestamp"`
	EndDate       string `json:"endDate,omitempty" jsonschema:"Filter events created on or before this RFC3339 timestamp"`
	SortParam     string `json:"sortParam,omitempty" jsonschema:"Device event field to sort by (e.g. sentOn)"`
	SortDir       string `json:"sortDir,omitempty" jsonschema:"Sort direction ASCENDING or DESCENDING"`
	AskTotalCount *bool  `json:"askTotalCount,omitempty" jsonschema:"Set true to request totalCount in the response"`
	Limit         int    `json:"limit,omitempty" jsonschema:"Maximum number of events to return"`
	Offset        int    `json:"offset,omitempty" jsonschema:"Number of events to skip before returning results"`
}

// HandleListDeviceEvents retrieves device events (logs) for a Kapua device.
func (h *KapuaHandler) HandleListDeviceEvents(ctx context.Context, req *mcp.CallToolRequest, params *ListDeviceEventsParams) (*mcp.CallToolResult, any, error) {
	if params == nil {
		return nil, nil, fmt.Errorf("device event parameters are required")
	}

	if params.DeviceID == "" {
		return nil, nil, fmt.Errorf("deviceId is required")
	}

	h.logger.Info("Listing device events for device %s", params.DeviceID)

	queryParams := make(map[string]string)
	if params.Resource != "" {
		queryParams["resource"] = params.Resource
	}
	if params.StartDate != "" {
		queryParams["startDate"] = params.StartDate
	}
	if params.EndDate != "" {
		queryParams["endDate"] = params.EndDate
	}
	if params.SortParam != "" {
		queryParams["sortParam"] = params.SortParam
	}
	if params.SortDir != "" {
		queryParams["sortDir"] = params.SortDir
	}
	if params.AskTotalCount != nil {
		queryParams["askTotalCount"] = strconv.FormatBool(*params.AskTotalCount)
	}
	if params.Limit > 0 {
		queryParams["limit"] = strconv.Itoa(params.Limit)
	}
	if params.Offset > 0 {
		queryParams["offset"] = strconv.Itoa(params.Offset)
	}

	result, err := h.client.ListDeviceEvents(ctx, params.DeviceID, queryParams)
	if err != nil {
		h.logger.Error("List device events failed: %v", err)
		return nil, nil, fmt.Errorf("failed to list device events: %w", err)
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal device events: %w", err)
	}

	summary := fmt.Sprintf("Found %d device events.", len(result.Items))
	if params.AskTotalCount != nil && *params.AskTotalCount {
		summary = fmt.Sprintf("Found %d device events (total count: %d).", len(result.Items), result.TotalCount)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(jsonData)},
		},
	}, result, nil
}
