package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"kapua-mcp-server/internal/kapua/config"
	"kapua-mcp-server/internal/kapua/models"
	"kapua-mcp-server/pkg/utils"
)

// KapuaClient provides access to Kapua REST API
type KapuaClient struct {
	config        *config.KapuaConfig
	httpClient    *http.Client
	logger        *utils.Logger
	baseURL       string
	token         string       // Current JWT token for authenticated requests
	tokenExpiry   time.Time    // Token expiration time
	refreshToken  string       // Refresh token
	refreshExpiry time.Time    // Refresh token expiration time
	tokenMutex    sync.RWMutex // Protects token-related fields
	autoRefresh   bool         // Enable automatic token refresh
	scopeId       string       // Default scope ID for account operations
}

// NewKapuaClient creates a new Kapua API client
func NewKapuaClient(cfg *config.KapuaConfig) *KapuaClient {
	// Ensure the base URL has the correct format
	baseURL := strings.TrimSuffix(cfg.APIEndpoint, "/")
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL = baseURL + "/v1"
	}

	return &KapuaClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
		logger:      utils.NewDefaultLogger("KapuaClient"),
		baseURL:     baseURL,
		autoRefresh: true, // Enable automatic token refresh by default
	}
}

// SetHTTPClient overrides the underlying HTTP client (used by tests to avoid network calls).
func (c *KapuaClient) SetHTTPClient(client *http.Client) {
	if client != nil {
		c.httpClient = client
	}
}

// makeRequest performs an HTTP request to the Kapua API
func (c *KapuaClient) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	url := c.baseURL + endpoint
	c.logger.Debug("Making %s request to %s", method, url)

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
		c.logger.Debug("Request body: %s", string(jsonData))
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Skip refresh handling for authentication endpoints to avoid recursion
	if !strings.HasPrefix(endpoint, "/authentication/") {
		if err := c.refreshTokenIfNeeded(ctx); err != nil {
			c.logger.Warn("Token refresh failed, continuing with current token: %v", err)
		}
	}

	token := c.getToken()
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	c.logger.Debug("Request info: %s %s", method, url)
	c.logger.Debug("Request body: %v", body)
	c.logger.Debug("Response status: %d", resp.StatusCode)
	return resp, nil
}

// handleResponse processes the HTTP response and unmarshals the result
func (c *KapuaClient) handleResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Debug("Response body: %s", string(body))

	// Check for error responses
	if resp.StatusCode >= 400 {
		var kapuaErr models.KapuaError
		if err := json.Unmarshal(body, &kapuaErr); err != nil {
			// If we can't unmarshal as KapuaError, return a generic error
			return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return kapuaErr
	}

	// Unmarshal successful response
	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// scopedEndpoint builds an endpoint path that automatically prefixes the client's scope ID.
// pathTemplate should start with "/" and may include additional formatting verbs for args.
func (c *KapuaClient) scopedEndpoint(pathTemplate string, args ...interface{}) string {
	return fmt.Sprintf("/%s"+pathTemplate, append([]interface{}{c.scopeId}, args...)...)
}

// doKapuaRequest wraps makeRequest and handleResponse, applying consistent error wrapping.
// action should describe the operation, e.g., "list devices" or "authenticate user".
func (c *KapuaClient) doKapuaRequest(ctx context.Context, method, endpoint, action string, body interface{}, out interface{}) error {
	resp, err := c.makeRequest(ctx, method, endpoint, body)
	if err != nil {
		return fmt.Errorf("%s request failed: %w", action, err)
	}

	if err := c.handleResponse(resp, out); err != nil {
		return fmt.Errorf("failed to %s: %w", action, err)
	}

	return nil
}
