package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DeviceCommandExecuteParams struct {
	DeviceID string         `json:"deviceId" jsonschema:"The device ID to execute command on"`
	Command  map[string]any `json:"command" jsonschema:"Command payload as object"`
}

func (h *KapuaHandler) HandleDeviceCommandExecute(ctx context.Context, req *mcp.CallToolRequest, params *DeviceCommandExecuteParams) (*mcp.CallToolResult, any, error) {
	h.logger.Info("Executing command on device %s", params.DeviceID)
	out, err := h.client.ExecuteDeviceCommand(ctx, params.DeviceID, params.Command)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute device command: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(out)}}}, json.RawMessage(out), nil
}
