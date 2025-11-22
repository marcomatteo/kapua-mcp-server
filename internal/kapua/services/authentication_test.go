package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"kapua-mcp-server/internal/kapua/config"
	"kapua-mcp-server/internal/kapua/models"
	"kapua-mcp-server/pkg/utils"
)

func TestSetTokenAndGetToken(t *testing.T) {
	client := &KapuaClient{logger: utils.NewDefaultLogger("test")}
	client.SetToken("abc123")

	if got := client.getToken(); got != "abc123" {
		t.Fatalf("expected token abc123, got %q", got)
	}
}

func TestSetTokenInfo(t *testing.T) {
	client := &KapuaClient{logger: utils.NewDefaultLogger("test")}
	now := time.Now().UTC().Truncate(time.Second)
	accessToken := &models.AccessToken{
		KapuaEntity:      models.KapuaEntity{ScopeID: models.KapuaID("root")},
		TokenID:          "token-123",
		RefreshToken:     "refresh-456",
		ExpiresOn:        now.Add(30 * time.Minute),
		RefreshExpiresOn: now.Add(2 * time.Hour),
	}

	client.SetTokenInfo(accessToken)

	client.tokenMutex.RLock()
	defer client.tokenMutex.RUnlock()
	if client.token != "token-123" {
		t.Fatalf("expected token to be set, got %q", client.token)
	}
	if client.refreshToken != "refresh-456" {
		t.Fatalf("expected refresh token to be set, got %q", client.refreshToken)
	}
	if !client.tokenExpiry.Equal(accessToken.ExpiresOn) {
		t.Fatalf("expected token expiry %v, got %v", accessToken.ExpiresOn, client.tokenExpiry)
	}
	if !client.refreshExpiry.Equal(accessToken.RefreshExpiresOn) {
		t.Fatalf("expected refresh expiry %v, got %v", accessToken.RefreshExpiresOn, client.refreshExpiry)
	}
	if client.scopeId != "root" {
		t.Fatalf("expected scopeId root, got %q", client.scopeId)
	}
}

func TestRefreshTokenIfNeededAutoRefreshDisabled(t *testing.T) {
	client := &KapuaClient{
		logger:      utils.NewDefaultLogger("test"),
		autoRefresh: false,
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("unexpected HTTP call to %s", req.URL.Path)
			return nil, nil
		})},
	}
	client.tokenExpiry = time.Now().Add(30 * time.Second)

	if err := client.refreshTokenIfNeeded(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRefreshTokenIfNeededNoExpiryKnown(t *testing.T) {
	client := &KapuaClient{
		logger:      utils.NewDefaultLogger("test"),
		autoRefresh: true,
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("unexpected HTTP call to %s", req.URL.Path)
			return nil, nil
		})},
	}

	if err := client.refreshTokenIfNeeded(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRefreshTokenIfNeededNotExpiring(t *testing.T) {
	client := &KapuaClient{
		logger:      utils.NewDefaultLogger("test"),
		autoRefresh: true,
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("unexpected HTTP call to %s", req.URL.Path)
			return nil, nil
		})},
	}
	client.tokenExpiry = time.Now().Add(10 * time.Minute)
	client.refreshToken = "refresh"

	if err := client.refreshTokenIfNeeded(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRefreshTokenIfNeededExpiringSoonRefreshSuccess(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	client := &KapuaClient{
		config:      &config.KapuaConfig{},
		logger:      utils.NewDefaultLogger("test"),
		baseURL:     "http://example.com/v1",
		autoRefresh: true,
	}

	refreshCalled := 0
	newTokenExpiry := now.Add(30 * time.Minute)
	newRefreshExpiry := now.Add(2 * time.Hour)
	responseBody := fmt.Sprintf(`{"tokenId":"new","refreshToken":"new-refresh","expiresOn":"%s","refreshExpiresOn":"%s"}`,
		newTokenExpiry.Format(time.RFC3339), newRefreshExpiry.Format(time.RFC3339))

	client.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if !strings.HasSuffix(req.URL.Path, "/authentication/refresh") {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		refreshCalled++
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(responseBody)),
			Header:     make(http.Header),
		}, nil
	})}

	client.tokenMutex.Lock()
	client.token = "old"
	client.tokenExpiry = now.Add(2 * time.Minute)
	client.refreshToken = "refresh"
	client.refreshExpiry = now.Add(time.Hour)
	client.tokenMutex.Unlock()

	if err := client.refreshTokenIfNeeded(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if refreshCalled != 1 {
		t.Fatalf("expected RefreshToken to be called once, got %d", refreshCalled)
	}
	client.tokenMutex.RLock()
	if client.token != "new" {
		client.tokenMutex.RUnlock()
		t.Fatalf("expected token to be updated, got %q", client.token)
	}
	if !client.tokenExpiry.Equal(newTokenExpiry) {
		client.tokenMutex.RUnlock()
		t.Fatalf("expected expiry %v, got %v", newTokenExpiry, client.tokenExpiry)
	}
	if !client.refreshExpiry.Equal(newRefreshExpiry) {
		client.tokenMutex.RUnlock()
		t.Fatalf("expected refresh expiry %v, got %v", newRefreshExpiry, client.refreshExpiry)
	}
	client.tokenMutex.RUnlock()
}

func TestRefreshTokenIfNeededExpiredNoRefreshToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	client := &KapuaClient{
		config:      &config.KapuaConfig{Username: "kapua-sys", Password: "kapua-password"},
		logger:      utils.NewDefaultLogger("test"),
		baseURL:     "http://example.com/v1",
		autoRefresh: true,
	}

	authCalled := 0
	newTokenExpiry := now.Add(45 * time.Minute)
	newRefreshExpiry := now.Add(3 * time.Hour)
	responseBody := fmt.Sprintf(`{"tokenId":"fresh","refreshToken":"fresh-refresh","expiresOn":"%s","refreshExpiresOn":"%s"}`,
		newTokenExpiry.Format(time.RFC3339), newRefreshExpiry.Format(time.RFC3339))

	client.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if !strings.HasSuffix(req.URL.Path, "/authentication/user") {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		authCalled++
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(responseBody)),
			Header:     make(http.Header),
		}, nil
	})}

	client.tokenMutex.Lock()
	client.token = "expired"
	client.tokenExpiry = now.Add(-1 * time.Minute)
	client.refreshToken = ""
	client.refreshExpiry = time.Time{}
	client.tokenMutex.Unlock()

	if err := client.refreshTokenIfNeeded(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if authCalled != 1 {
		t.Fatalf("expected QuickAuthenticate to be called once, got %d", authCalled)
	}
	client.tokenMutex.RLock()
	if client.token != "fresh" {
		client.tokenMutex.RUnlock()
		t.Fatalf("expected token to be updated, got %q", client.token)
	}
	if !client.tokenExpiry.Equal(newTokenExpiry) {
		client.tokenMutex.RUnlock()
		t.Fatalf("expected expiry %v, got %v", newTokenExpiry, client.tokenExpiry)
	}
	if !client.refreshExpiry.Equal(newRefreshExpiry) {
		client.tokenMutex.RUnlock()
		t.Fatalf("expected refresh expiry %v, got %v", newRefreshExpiry, client.refreshExpiry)
	}
	client.tokenMutex.RUnlock()
}

func TestRefreshTokenIfNeededRefreshFails(t *testing.T) {
	client := &KapuaClient{
		config:      &config.KapuaConfig{},
		logger:      utils.NewDefaultLogger("test"),
		baseURL:     "http://example.com/v1",
		autoRefresh: true,
	}

	client.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       io.NopCloser(strings.NewReader("bad")),
			Header:     make(http.Header),
		}, nil
	})}

	client.tokenMutex.Lock()
	client.token = "old"
	client.tokenExpiry = time.Now().Add(30 * time.Second)
	client.refreshToken = "refresh"
	client.refreshExpiry = time.Now().Add(time.Hour)
	client.tokenMutex.Unlock()

	err := client.refreshTokenIfNeeded(context.Background())
	if err == nil {
		t.Fatal("expected error from refresh failure")
	}
	if !strings.Contains(err.Error(), "bad") {
		t.Fatalf("expected error to include response body, got %v", err)
	}
}

func TestAuthenticateUserSuccess(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	tokenJSON := fmt.Sprintf(`{"tokenId":"new-token","refreshToken":"refresh-token","expiresOn":"%s","refreshExpiresOn":"%s","scopeId":"root"}`,
		now.Add(time.Hour).Format(time.RFC3339), now.Add(2*time.Hour).Format(time.RFC3339))

	client := &KapuaClient{
		config:  &config.KapuaConfig{},
		logger:  utils.NewDefaultLogger("test"),
		baseURL: "http://example.com/v1",
	}

	client.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.String() != "http://example.com/v1/authentication/user" {
			t.Fatalf("unexpected URL %s", req.URL.String())
		}
		body, _ := io.ReadAll(req.Body)
		expectedBody := `{"username":"kapua-sys","password":"kapua-password"}`
		if strings.TrimSpace(string(body)) != expectedBody {
			t.Fatalf("unexpected body: %s", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(tokenJSON)),
			Header:     make(http.Header),
		}, nil
	})}

	token, err := client.AuthenticateUser(context.Background(), models.UsernamePasswordCredentials{
		Username: "kapua-sys",
		Password: "kapua-password",
	})
	if err != nil {
		t.Fatalf("AuthenticateUser returned error: %v", err)
	}
	if token == nil || token.TokenID != "new-token" {
		t.Fatalf("expected token to be populated, got %+v", token)
	}
	client.tokenMutex.RLock()
	if client.token != "new-token" {
		client.tokenMutex.RUnlock()
		t.Fatalf("expected client token to be stored")
	}
	if client.scopeId != "root" {
		client.tokenMutex.RUnlock()
		t.Fatalf("expected scopeId to be set, got %q", client.scopeId)
	}
	client.tokenMutex.RUnlock()
}

func TestAuthenticateUserRequestFailure(t *testing.T) {
	client := &KapuaClient{
		config:  &config.KapuaConfig{},
		logger:  utils.NewDefaultLogger("test"),
		baseURL: "http://example.com/v1",
	}
	client.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})}

	_, err := client.AuthenticateUser(context.Background(), models.UsernamePasswordCredentials{Username: "kapua-sys", Password: "kapua-password"})
	if err == nil || !strings.Contains(err.Error(), "authenticate user request failed") {
		t.Fatalf("expected wrapped request error, got %v", err)
	}
}

func TestAuthenticateUserHandleResponseError(t *testing.T) {
	client := &KapuaClient{
		config:  &config.KapuaConfig{},
		logger:  utils.NewDefaultLogger("test"),
		baseURL: "http://example.com/v1",
	}

	client.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(strings.NewReader(`{"code":"ERR","message":"Invalid","details":"bad creds"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	_, err := client.AuthenticateUser(context.Background(), models.UsernamePasswordCredentials{Username: "kapua-sys", Password: "kapua-password"})
	if err == nil {
		t.Fatal("expected error")
	}
	var kapuaErr models.KapuaError
	if !errors.As(err, &kapuaErr) {
		t.Fatalf("expected KapuaError, got %T", err)
	}
}

func TestAuthenticateAPIKeySuccess(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	tokenJSON := fmt.Sprintf(`{"tokenId":"api-token","refreshToken":"refresh-token","expiresOn":"%s","refreshExpiresOn":"%s"}`,
		now.Add(time.Hour).Format(time.RFC3339), now.Add(2*time.Hour).Format(time.RFC3339))

	client := &KapuaClient{
		config:  &config.KapuaConfig{},
		logger:  utils.NewDefaultLogger("test"),
		baseURL: "http://example.com/v1",
	}

	client.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "http://example.com/v1/authentication/apikey" {
			t.Fatalf("unexpected URL %s", req.URL.String())
		}
		body, _ := io.ReadAll(req.Body)
		expectedBody := `{"apiKey":"z8PEVr4XdBS/KKEKbVG9tJzj6DNNpSVCDpW53CWm"}`
		if strings.TrimSpace(string(body)) != expectedBody {
			t.Fatalf("unexpected body: %s", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(tokenJSON)),
			Header:     make(http.Header),
		}, nil
	})}

	token, err := client.AuthenticateAPIKey(context.Background(), models.APIKeyCredentials{APIKey: "z8PEVr4XdBS/KKEKbVG9tJzj6DNNpSVCDpW53CWm"})
	if err != nil {
		t.Fatalf("AuthenticateAPIKey returned error: %v", err)
	}
	if token == nil || token.TokenID != "api-token" {
		t.Fatalf("expected token to be populated, got %+v", token)
	}
}

func TestAuthenticateJWTSuccess(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	tokenJSON := fmt.Sprintf(`{"tokenId":"jwt-token","refreshToken":"refresh-token","expiresOn":"%s","refreshExpiresOn":"%s"}`,
		now.Add(time.Hour).Format(time.RFC3339), now.Add(2*time.Hour).Format(time.RFC3339))

	client := &KapuaClient{
		config:  &config.KapuaConfig{},
		logger:  utils.NewDefaultLogger("test"),
		baseURL: "http://example.com/v1",
	}

	client.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "http://example.com/v1/authentication/jwt" {
			t.Fatalf("unexpected URL %s", req.URL.String())
		}
		body, _ := io.ReadAll(req.Body)
		expectedBody := `{"jwt":"jwt-token"}`
		if strings.TrimSpace(string(body)) != expectedBody {
			t.Fatalf("unexpected body: %s", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(tokenJSON)),
			Header:     make(http.Header),
		}, nil
	})}

	token, err := client.AuthenticateJWT(context.Background(), models.JWTCredentials{JWT: "jwt-token"})
	if err != nil {
		t.Fatalf("AuthenticateJWT returned error: %v", err)
	}
	if token == nil || token.TokenID != "jwt-token" {
		t.Fatalf("expected token to be populated, got %+v", token)
	}
}

func TestRefreshTokenSuccess(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	tokenJSON := fmt.Sprintf(`{"tokenId":"new-token","refreshToken":"refresh-token","expiresOn":"%s","refreshExpiresOn":"%s"}`,
		now.Add(time.Hour).Format(time.RFC3339), now.Add(2*time.Hour).Format(time.RFC3339))

	client := &KapuaClient{
		config:  &config.KapuaConfig{},
		logger:  utils.NewDefaultLogger("test"),
		baseURL: "http://example.com/v1",
	}

	client.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(req.Body)
		expectedBody := `{"refreshToken":"refresh-token","tokenId":"old-token"}`
		if strings.TrimSpace(string(body)) != expectedBody {
			t.Fatalf("unexpected body: %s", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(tokenJSON)),
			Header:     make(http.Header),
		}, nil
	})}

	token, err := client.RefreshToken(context.Background(), models.RefreshTokenRequest{RefreshToken: "refresh-token", TokenID: "old-token"})
	if err != nil {
		t.Fatalf("RefreshToken returned error: %v", err)
	}
	if token.TokenID != "new-token" {
		t.Fatalf("expected new token, got %s", token.TokenID)
	}
}

func TestGetLoginInfoRequiresToken(t *testing.T) {
	client := &KapuaClient{logger: utils.NewDefaultLogger("test")}

	if _, err := client.GetLoginInfo(context.Background()); err == nil || !strings.Contains(err.Error(), "no authentication token") {
		t.Fatalf("expected missing token error, got %v", err)
	}
}

func TestGetLoginInfoSuccess(t *testing.T) {
	client := &KapuaClient{
		logger:  utils.NewDefaultLogger("test"),
		baseURL: "http://example.com/v1",
	}
	client.SetToken("bearer-token")

	loginJSON := `{"accessToken":{"tokenId":"current"}}`

	client.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("Authorization") != "Bearer bearer-token" {
			t.Fatalf("expected Authorization header, got %q", req.Header.Get("Authorization"))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(loginJSON)),
			Header:     make(http.Header),
		}, nil
	})}

	info, err := client.GetLoginInfo(context.Background())
	if err != nil {
		t.Fatalf("GetLoginInfo returned error: %v", err)
	}
	if info.AccessToken.TokenID != "current" {
		t.Fatalf("expected access token id current, got %s", info.AccessToken.TokenID)
	}
}

func TestLogoutRequiresToken(t *testing.T) {
	client := &KapuaClient{logger: utils.NewDefaultLogger("test")}

	if err := client.Logout(context.Background()); err == nil || !strings.Contains(err.Error(), "no authentication token") {
		t.Fatalf("expected missing token error, got %v", err)
	}
}

func TestLogoutSuccessClearsToken(t *testing.T) {
	client := &KapuaClient{
		logger:  utils.NewDefaultLogger("test"),
		baseURL: "http://example.com/v1",
	}
	client.SetToken("bearer-token")
	client.tokenMutex.Lock()
	client.tokenExpiry = time.Now().Add(time.Hour)
	client.refreshToken = "refresh"
	client.refreshExpiry = time.Now().Add(2 * time.Hour)
	client.tokenMutex.Unlock()

	client.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})}

	if err := client.Logout(context.Background()); err != nil {
		t.Fatalf("Logout returned error: %v", err)
	}
	client.tokenMutex.RLock()
	defer client.tokenMutex.RUnlock()
	if client.token != "" || !client.tokenExpiry.IsZero() || client.refreshToken != "" || !client.refreshExpiry.IsZero() {
		t.Fatalf("expected token state to be cleared")
	}
}

func TestQuickAuthenticateUsesConfig(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	tokenJSON := fmt.Sprintf(`{"tokenId":"quick-token","refreshToken":"refresh-token","expiresOn":"%s","refreshExpiresOn":"%s"}`,
		now.Add(time.Hour).Format(time.RFC3339), now.Add(2*time.Hour).Format(time.RFC3339))

	client := &KapuaClient{
		config:  &config.KapuaConfig{Username: "kapua-sys", Password: "kapua-password"},
		logger:  utils.NewDefaultLogger("test"),
		baseURL: "http://example.com/v1",
	}

	client.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(req.Body)
		expectedBody := `{"username":"kapua-sys","password":"kapua-password"}`
		if strings.TrimSpace(string(body)) != expectedBody {
			t.Fatalf("expected credentials from config, got %s", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(tokenJSON)),
			Header:     make(http.Header),
		}, nil
	})}

	token, err := client.QuickAuthenticate(context.Background())
	if err != nil {
		t.Fatalf("QuickAuthenticate returned error: %v", err)
	}
	if token.TokenID != "quick-token" {
		t.Fatalf("expected quick token, got %s", token.TokenID)
	}
}
