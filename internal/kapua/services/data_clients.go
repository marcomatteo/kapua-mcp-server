package services

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"kapua-mcp-server/internal/kapua/models"
)

// ListDataClients queries the Kapua Data Client registry within the current scope.
func (c *KapuaClient) ListDataClients(ctx context.Context, params map[string]string) (*models.ClientInfoListResult, error) {
	c.logger.Info("Listing data clients for scope: %s", c.scopeId)

	query := url.Values{}
	for key, value := range params {
		if value != "" {
			query.Set(key, value)
		}
	}

	endpoint := fmt.Sprintf("/%s/data/clients", c.scopeId)
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("list data clients request failed: %w", err)
	}

	var result models.ClientInfoListResult
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to list data clients: %w", err)
	}

	c.logger.Info("Listed %d data clients successfully", len(result.Items))
	return &result, nil
}

// CountDataClients returns how many data clients exist for the provided query.
func (c *KapuaClient) CountDataClients(ctx context.Context, query *models.KapuaQuery) (*models.KapuaCountResult, error) {
	c.logger.Info("Counting data clients for scope: %s", c.scopeId)

	payload := query
	if payload == nil {
		payload = &models.KapuaQuery{}
	}

	endpoint := fmt.Sprintf("/%s/data/clients/_count", c.scopeId)
	resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("count data clients request failed: %w", err)
	}

	var result models.KapuaCountResult
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to count data clients: %w", err)
	}

	c.logger.Info("Counted %d data clients", result.Count)
	return &result, nil
}

// GetDataClient retrieves a single data client by its clientInfo identifier.
func (c *KapuaClient) GetDataClient(ctx context.Context, clientInfoID string) (*models.ClientInfo, error) {
	c.logger.Info("Getting data client %s in scope: %s", clientInfoID, c.scopeId)

	endpoint := fmt.Sprintf("/%s/data/clients/%s", c.scopeId, clientInfoID)
	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("get data client request failed: %w", err)
	}

	var info models.ClientInfo
	if err := c.handleResponse(resp, &info); err != nil {
		return nil, fmt.Errorf("failed to get data client: %w", err)
	}

	c.logger.Info("Retrieved data client %s", info.ClientID)
	return &info, nil
}
