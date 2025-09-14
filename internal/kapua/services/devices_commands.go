package services

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

// Commands-related device APIs

// ExecuteDeviceCommand executes a command on the device
// command payload should follow Kapua spec for commands/_execute
func (c *KapuaClient) ExecuteDeviceCommand(ctx context.Context, deviceID string, command any) (json.RawMessage, error) {
    endpoint := fmt.Sprintf("/%s/devices/%s/commands/_execute", c.scopeId, deviceID)
    resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, command)
    if err != nil {
        return nil, fmt.Errorf("execute device command request failed: %w", err)
    }
    var out json.RawMessage
    if err := c.handleResponse(resp, &out); err != nil {
        return nil, fmt.Errorf("failed to execute device command: %w", err)
    }
    return out, nil
}

