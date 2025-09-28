package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"kapua-mcp-server/internal/kapua/services"
)

// Data message tooling parameters

// ListDataMessagesParams captures the filters for retrieving Kapua data messages.
type ListDataMessagesParams struct {
	ClientIDs     []string `json:"clientIds,omitempty" jsonschema:"Filter data messages by one or more clientIds"`
	Channel       string   `json:"channel,omitempty" jsonschema:"Filter data messages by channel"`
	StrictChannel *bool    `json:"strictChannel,omitempty" jsonschema:"Restrict search to the provided channel only"`
	StartDate     string   `json:"startDate,omitempty" jsonschema:"Filter messages captured on or after this timestamp"`
	EndDate       string   `json:"endDate,omitempty" jsonschema:"Filter messages captured on or before this timestamp"`
	SortDir       string   `json:"sortDir,omitempty" jsonschema:"Sort direction ASC or DESC"`
	Limit         *int     `json:"limit,omitempty" jsonschema:"Maximum number of messages to return"`
	Offset        *int     `json:"offset,omitempty" jsonschema:"Number of messages to skip before returning results"`
}

// GetDataMessageParams identifies a specific Kapua data message entry.
type GetDataMessageParams struct {
	DatastoreMessageID string `json:"datastoreMessageId" jsonschema:"Kapua datastore message identifier"`
}

// HandleListDataMessages lists data messages registered in Kapua for the current scope.
func (h *KapuaHandler) HandleListDataMessages(ctx context.Context, req *mcp.CallToolRequest, params *ListDataMessagesParams) (*mcp.CallToolResult, any, error) {
	if params == nil {
		params = &ListDataMessagesParams{}
	}

	h.logger.Info("Listing data messages")

	query := &services.DataMessagesQuery{
		ClientIDs:     params.ClientIDs,
		Channel:       params.Channel,
		StrictChannel: params.StrictChannel,
		StartDate:     params.StartDate,
		EndDate:       params.EndDate,
		SortDir:       params.SortDir,
		Limit:         params.Limit,
		Offset:        params.Offset,
	}

	// Avoid passing an empty query object to keep the request clean.
	if len(params.ClientIDs) == 0 && params.Channel == "" && params.StrictChannel == nil &&
		params.StartDate == "" && params.EndDate == "" && params.SortDir == "" &&
		params.Limit == nil && params.Offset == nil {
		query = nil
	}

	result, err := h.client.ListDataMessages(ctx, query)
	if err != nil {
		h.logger.Error("List data messages failed: %v", err)
		return nil, nil, fmt.Errorf("failed to list data messages: %w", err)
	}

	payload, err := json.Marshal(result)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal data messages: %w", err)
	}

	summary := fmt.Sprintf("Found %d data messages.", len(result.Items))
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
			&mcp.TextContent{Text: string(payload)},
		},
	}, result, nil
}
