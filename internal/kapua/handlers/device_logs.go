package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"kapua-mcp-server/internal/kapua/services"
)

// Device log tooling parameters

// ListDeviceLogsParams captures filters for retrieving Kapua device logs.
type ListDeviceLogsParams struct {
	ClientID        string `json:"clientId,omitempty" jsonschema:"Filter device logs by clientId"`
	Channel         string `json:"channel,omitempty" jsonschema:"Filter device logs by channel"`
	StrictChannel   *bool  `json:"strictChannel,omitempty" jsonschema:"Restrict search to the provided channel only"`
	StartDate       string `json:"startDate,omitempty" jsonschema:"Filter logs captured on or after this timestamp"`
	EndDate         string `json:"endDate,omitempty" jsonschema:"Filter logs captured on or before this timestamp"`
	LogPropertyName string `json:"logPropertyName,omitempty" jsonschema:"Filter logs by property name"`
	LogPropertyType string `json:"logPropertyType,omitempty" jsonschema:"Filter logs by property type"`
	LogPropertyMin  string `json:"logPropertyMin,omitempty" jsonschema:"Filter logs by minimum property value"`
	LogPropertyMax  string `json:"logPropertyMax,omitempty" jsonschema:"Filter logs by maximum property value"`
	SortDir         string `json:"sortDir,omitempty" jsonschema:"Sort direction ASCENDING or DESCENDING"`
	Limit           *int   `json:"limit,omitempty" jsonschema:"Maximum number of logs to return"`
	Offset          *int   `json:"offset,omitempty" jsonschema:"Number of logs to skip before returning results"`
}

// HandleListDeviceLogs lists device logs registered in Kapua for the current scope.
func (h *KapuaHandler) HandleListDeviceLogs(ctx context.Context, req *mcp.CallToolRequest, params *ListDeviceLogsParams) (*mcp.CallToolResult, any, error) {
	if params == nil {
		params = &ListDeviceLogsParams{}
	}

	h.logger.Info("Listing device logs")

	query := &services.DeviceLogsQuery{
		ClientID:        params.ClientID,
		Channel:         params.Channel,
		StrictChannel:   params.StrictChannel,
		StartDate:       params.StartDate,
		EndDate:         params.EndDate,
		LogPropertyName: params.LogPropertyName,
		LogPropertyType: params.LogPropertyType,
		LogPropertyMin:  params.LogPropertyMin,
		LogPropertyMax:  params.LogPropertyMax,
		SortDir:         params.SortDir,
		Limit:           params.Limit,
		Offset:          params.Offset,
	}

	// Avoid passing an empty query to the service.
	if params.ClientID == "" && params.Channel == "" && params.StrictChannel == nil &&
		params.StartDate == "" && params.EndDate == "" && params.LogPropertyName == "" &&
		params.LogPropertyType == "" && params.LogPropertyMin == "" && params.LogPropertyMax == "" &&
		params.SortDir == "" && params.Limit == nil && params.Offset == nil {
		query = nil
	}

	result, err := h.client.ListDeviceLogs(ctx, query)
	if err != nil {
		h.logger.Error("List device logs failed: %v", err)
		if errors.Is(err, services.ErrDeviceLogsNotSupported) {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("failed to list device logs: %w", err)
	}

	payload, err := json.Marshal(result)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal device logs: %w", err)
	}

	summary := fmt.Sprintf("Found %d device logs.", len(result.Items))
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(payload)},
		},
	}, result, nil
}
