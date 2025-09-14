package services

import (
	"context"
	"fmt"
	"kapua-mcp-server/internal/kapua/models"
	"net/http"
	"time"
)

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

// refreshTokenIfNeeded refreshes an access token when expiring soon or already expired.
// If already expired and the refresh token is also expired, it falls back to QuickAuthenticate.
func (c *KapuaClient) refreshTokenIfNeeded(ctx context.Context) error {
	if !c.autoRefresh {
		return nil
	}

	// Snapshot token timings and values without holding the lock during I/O
	c.tokenMutex.RLock()
	tokenExpiry := c.tokenExpiry
	refreshExpiry := c.refreshExpiry
	refreshToken := c.refreshToken
	tokenID := c.token
	c.tokenMutex.RUnlock()

	if tokenExpiry.IsZero() {
		// No expiry known; nothing to do
		return nil
	}

	now := time.Now()
	expiringSoon := time.Until(tokenExpiry) < 5*time.Minute
	expired := now.After(tokenExpiry)
	if !expiringSoon && !expired {
		return nil
	}

	// If already expired, ensure refresh token is still valid; otherwise re-authenticate
	if expired {
		if refreshToken == "" {
			c.logger.Warn("Access token expired and no refresh token available; performing full re-authentication")
			_, err := c.QuickAuthenticate(ctx)
			return err
		}
		if refreshExpiry.IsZero() || now.After(refreshExpiry) {
			c.logger.Info("Refresh token expired; performing full re-authentication")
			_, err := c.QuickAuthenticate(ctx)
			return err
		}
		// else: refresh token still valid; proceed to refresh below
		c.logger.Info("Access token expired; attempting automatic refresh")
	} else {
		c.logger.Info("Token expiring soon, attempting automatic refresh")
	}

	// Attempt refresh with current refresh token
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
