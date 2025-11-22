package services

import (
	"context"
	"encoding/json"
	"net/http"
)

// Commands-related device APIs

// ExecuteDeviceCommand executes a command on the device
// command payload should follow Kapua spec for commands/_execute
func (c *KapuaClient) ExecuteDeviceCommand(ctx context.Context, deviceID string, command any) (json.RawMessage, error) {
	endpoint := c.scopedEndpoint("/devices/%s/commands/_execute", deviceID)
	var out json.RawMessage
	if err := c.doKapuaRequest(ctx, http.MethodPost, endpoint, "execute device command", command, &out); err != nil {
		return nil, err
	}
	return out, nil
}
