package handlers

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"kapua-mcp-server/internal/kapua/services"
	"kapua-mcp-server/pkg/utils"
)

// KapuaHandler provides MCP tool handlers for Kapua operations
type KapuaHandler struct {
	client *services.KapuaClient
	logger *utils.Logger
}

// NewKapuaHandler creates a new Kapua handler
func NewKapuaHandler(client *services.KapuaClient) *KapuaHandler {
	return &KapuaHandler{
		client: client,
		logger: utils.NewDefaultLogger("KapuaHandler"),
	}
}

// MCP Resource Handlers

// ListResources returns a list of available Kapua resources
func (h *KapuaHandler) ListResources(ctx context.Context) ([]mcp.Resource, error) {
	h.logger.Debug("Listing available Kapua resources")

	resources := []mcp.Resource{
		{
			URI:         "kapua://devices",
			Name:        "Kapua Devices",
			Description: "Live list of all Kapua IoT devices with current status and metadata",
			MIMEType:    "application/json",
		},
	}

	return resources, nil
}

// ReadResource returns the content of a specific Kapua resource
func (h *KapuaHandler) ReadResource(ctx context.Context, uri string) (*mcp.ReadResourceResult, error) {
	h.logger.Debug("Reading Kapua resource: %s", uri)

	switch uri {
	case "kapua://devices":
		return h.readDevicesResource(ctx)
	default:
		return nil, fmt.Errorf("unknown resource URI: %s", uri)
	}
}
