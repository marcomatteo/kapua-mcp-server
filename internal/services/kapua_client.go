package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"kapua-mcp-server/internal/config"
	"kapua-mcp-server/internal/models"
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

// SetToken sets the authentication token for subsequent requests
func (c *KapuaClient) SetToken(token string) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.token = token
	c.logger.Debug("Authentication token updated")
}

// SetTokenInfo sets comprehensive token information including expiry and refresh token
func (c *KapuaClient) SetTokenInfo(accessToken *models.AccessToken) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.token = accessToken.TokenID
	c.tokenExpiry = accessToken.ExpiresOn
	c.refreshToken = accessToken.RefreshToken
	c.refreshExpiry = accessToken.RefreshExpiresOn
	c.scopeId = accessToken.ScopeID.String()
	c.logger.Debug("Token information updated - expires: %v, refresh expires: %v",
		c.tokenExpiry.Format(time.RFC3339), c.refreshExpiry.Format(time.RFC3339))
}

// getToken safely retrieves the current token
func (c *KapuaClient) getToken() string {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	return c.token
}

// isTokenExpiringSoon checks if the token expires within the next 5 minutes
func (c *KapuaClient) isTokenExpiringSoon() bool {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	if c.tokenExpiry.IsZero() {
		return false
	}
	return time.Until(c.tokenExpiry) < 5*time.Minute
}

// refreshTokenIfNeeded automatically refreshes the token if it's expiring soon
func (c *KapuaClient) refreshTokenIfNeeded(ctx context.Context) error {
	if !c.autoRefresh || !c.isTokenExpiringSoon() {
		return nil
	}

	c.tokenMutex.RLock()
	refreshToken := c.refreshToken
	tokenID := c.token
	c.tokenMutex.RUnlock()

	if refreshToken == "" {
		c.logger.Warn("Token expiring soon but no refresh token available")
		return fmt.Errorf("no refresh token available for automatic refresh")
	}

	c.logger.Info("Token expiring soon, attempting automatic refresh")
	request := models.RefreshTokenRequest{
		RefreshToken: refreshToken,
		TokenID:      tokenID,
	}

	_, err := c.RefreshToken(ctx, request)
	if err != nil {
		c.logger.Error("Automatic token refresh failed: %v", err)
		return err
	}

	c.logger.Info("Token automatically refreshed successfully")
	return nil
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

	// Check if token needs refresh and add authentication header
	if err := c.refreshTokenIfNeeded(ctx); err != nil {
		c.logger.Warn("Token refresh failed, continuing with current token: %v", err)
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

// Authentication Methods

// AuthenticateUser authenticates using username and password
func (c *KapuaClient) AuthenticateUser(ctx context.Context, credentials models.UsernamePasswordCredentials) (*models.AccessToken, error) {
	c.logger.Info("Authenticating user: %s", credentials.Username)

	resp, err := c.makeRequest(ctx, http.MethodPost, "/authentication/user", credentials)
	if err != nil {
		return nil, fmt.Errorf("authentication request failed: %w", err)
	}

	var token models.AccessToken
	if err := c.handleResponse(resp, &token); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Store the token information for subsequent requests
	c.SetTokenInfo(&token)
	c.logger.Info("User authentication successful")

	return &token, nil
}

// AuthenticateAPIKey authenticates using an API key
func (c *KapuaClient) AuthenticateAPIKey(ctx context.Context, credentials models.APIKeyCredentials) (*models.AccessToken, error) {
	c.logger.Info("Authenticating with API key")

	resp, err := c.makeRequest(ctx, http.MethodPost, "/authentication/apikey", credentials)
	if err != nil {
		return nil, fmt.Errorf("API key authentication request failed: %w", err)
	}

	var token models.AccessToken
	if err := c.handleResponse(resp, &token); err != nil {
		return nil, fmt.Errorf("API key authentication failed: %w", err)
	}

	// Store the token information for subsequent requests
	c.SetTokenInfo(&token)
	c.logger.Info("API key authentication successful")

	return &token, nil
}

// AuthenticateJWT authenticates using a JWT token
func (c *KapuaClient) AuthenticateJWT(ctx context.Context, credentials models.JWTCredentials) (*models.AccessToken, error) {
	c.logger.Info("Authenticating with JWT")

	resp, err := c.makeRequest(ctx, http.MethodPost, "/authentication/jwt", credentials)
	if err != nil {
		return nil, fmt.Errorf("JWT authentication request failed: %w", err)
	}

	var token models.AccessToken
	if err := c.handleResponse(resp, &token); err != nil {
		return nil, fmt.Errorf("JWT authentication failed: %w", err)
	}

	// Store the token information for subsequent requests
	c.SetTokenInfo(&token)
	c.logger.Info("JWT authentication successful")

	return &token, nil
}

// RefreshToken refreshes an existing access token
func (c *KapuaClient) RefreshToken(ctx context.Context, request models.RefreshTokenRequest) (*models.AccessToken, error) {
	c.logger.Info("Refreshing access token")

	resp, err := c.makeRequest(ctx, http.MethodPost, "/authentication/refresh", request)
	if err != nil {
		return nil, fmt.Errorf("token refresh request failed: %w", err)
	}

	var token models.AccessToken
	if err := c.handleResponse(resp, &token); err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	// Store the new token information for subsequent requests
	c.SetTokenInfo(&token)
	c.logger.Info("Token refresh successful")

	return &token, nil
}

// GetLoginInfo retrieves comprehensive authentication and authorization information
func (c *KapuaClient) GetLoginInfo(ctx context.Context) (*models.LoginInfo, error) {
	c.logger.Info("Retrieving login information")

	if c.token == "" {
		return nil, fmt.Errorf("no authentication token available")
	}

	resp, err := c.makeRequest(ctx, http.MethodGet, "/authentication/info", nil)
	if err != nil {
		return nil, fmt.Errorf("login info request failed: %w", err)
	}

	var loginInfo models.LoginInfo
	if err := c.handleResponse(resp, &loginInfo); err != nil {
		return nil, fmt.Errorf("failed to retrieve login info: %w", err)
	}

	c.logger.Info("Login information retrieved successfully")
	return &loginInfo, nil
}

// Logout invalidates the current session
func (c *KapuaClient) Logout(ctx context.Context) error {
	c.logger.Info("Logging out")

	if c.token == "" {
		return fmt.Errorf("no authentication token available")
	}

	resp, err := c.makeRequest(ctx, http.MethodPost, "/authentication/logout", nil)
	if err != nil {
		return fmt.Errorf("logout request failed: %w", err)
	}

	if err := c.handleResponse(resp, nil); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	// Clear the stored token
	c.tokenMutex.Lock()
	c.token = ""
	c.tokenExpiry = time.Time{}
	c.refreshToken = ""
	c.refreshExpiry = time.Time{}
	c.tokenMutex.Unlock()
	c.logger.Info("Logout successful")

	return nil
}

// QuickAuthenticate performs a quick authentication using configured credentials
func (c *KapuaClient) QuickAuthenticate(ctx context.Context) (*models.AccessToken, error) {
	credentials := models.UsernamePasswordCredentials{
		Username: c.config.Username,
		Password: c.config.Password,
	}

	return c.AuthenticateUser(ctx, credentials)
}

// Device Management Methods

// ListDevices retrieves a list of devices from a scope
func (c *KapuaClient) ListDevices(ctx context.Context, params map[string]string) (*models.DeviceListResult, error) {
	c.logger.Info("Listing devices for scope: %s", c.scopeId)

	// Build query parameters
	queryParams := url.Values{}
	for key, value := range params {
		if value != "" {
			queryParams.Set(key, value)
		}
	}

	endpoint := fmt.Sprintf("/%s/devices", c.scopeId)
	if len(queryParams) > 0 {
		endpoint += "?" + queryParams.Encode()
	}

	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("list devices request failed: %w", err)
	}

	var result models.DeviceListResult
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	c.logger.Info("Listed %d devices successfully", len(result.Items))
	return &result, nil
}

// GetDevice retrieves a specific device by ID
func (c *KapuaClient) GetDevice(ctx context.Context, deviceID string) (*models.Device, error) {
	c.logger.Info("Getting device %s from scope: %s", deviceID, c.scopeId)

	endpoint := fmt.Sprintf("/%s/devices/%s", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("get device request failed: %w", err)
	}

	var device models.Device
	if err := c.handleResponse(resp, &device); err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	c.logger.Info("Device retrieved successfully: %s", device.ClientID)
	return &device, nil
}

// CreateDevice creates a new device
func (c *KapuaClient) CreateDevice(ctx context.Context, creator models.DeviceCreator) (*models.Device, error) {
	c.logger.Info("Creating device %s in scope: %s", creator.ClientID, c.scopeId)

	endpoint := fmt.Sprintf("/%s/devices", c.scopeId)
	resp, err := c.makeRequest(ctx, http.MethodPost, endpoint, creator)
	if err != nil {
		return nil, fmt.Errorf("create device request failed: %w", err)
	}

	var device models.Device
	if err := c.handleResponse(resp, &device); err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	c.logger.Info("Device created successfully: %s (ID: %s)", device.ClientID, device.ID)
	return &device, nil
}

// UpdateDevice updates an existing device
func (c *KapuaClient) UpdateDevice(ctx context.Context, deviceID string, device models.Device) (*models.Device, error) {
	c.logger.Info("Updating device %s in scope: %s", deviceID, c.scopeId)

	endpoint := fmt.Sprintf("/%s/devices/%s", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodPut, endpoint, device)
	if err != nil {
		return nil, fmt.Errorf("update device request failed: %w", err)
	}

	var updatedDevice models.Device
	if err := c.handleResponse(resp, &updatedDevice); err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	c.logger.Info("Device updated successfully: %s", updatedDevice.ClientID)
	return &updatedDevice, nil
}

// DeleteDevice deletes a device
func (c *KapuaClient) DeleteDevice(ctx context.Context, deviceID string) error {
	c.logger.Info("Deleting device %s from scope: %s", deviceID, c.scopeId)

	endpoint := fmt.Sprintf("/%s/devices/%s", c.scopeId, deviceID)
	resp, err := c.makeRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("delete device request failed: %w", err)
	}

	if err := c.handleResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	c.logger.Info("Device deleted successfully")
	return nil
}
