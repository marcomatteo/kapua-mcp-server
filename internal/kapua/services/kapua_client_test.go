package services

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"kapua-mcp-server/internal/kapua/config"
	"kapua-mcp-server/internal/kapua/models"
	"kapua-mcp-server/pkg/utils"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestNewKapuaClientBaseURL(t *testing.T) {
	cfg := &config.KapuaConfig{APIEndpoint: "http://example.com/api", Timeout: 10}
	client := NewKapuaClient(cfg)
	expected := "http://example.com/api/v1"
	if client.baseURL != expected {
		t.Fatalf("expected baseURL %q, got %q", expected, client.baseURL)
	}
	if client.httpClient.Timeout != 10*time.Second {
		t.Fatalf("expected timeout 10s, got %v", client.httpClient.Timeout)
	}
	if !client.autoRefresh {
		t.Fatalf("expected autoRefresh to be true")
	}
}

func TestNewKapuaClientBaseURLAlreadyV1(t *testing.T) {
	cfg := &config.KapuaConfig{APIEndpoint: "http://example.com/api/v1", Timeout: 5}
	client := NewKapuaClient(cfg)
	expected := "http://example.com/api/v1"
	if client.baseURL != expected {
		t.Fatalf("expected baseURL %q, got %q", expected, client.baseURL)
	}
	if client.httpClient.Timeout != 5*time.Second {
		t.Fatalf("expected timeout 5s, got %v", client.httpClient.Timeout)
	}
}

func TestKapuaClientMakeRequestSuccess(t *testing.T) {
	client := &KapuaClient{
		baseURL:     "http://example.com/v1",
		httpClient:  &http.Client{},
		logger:      utils.NewDefaultLogger("test"),
		autoRefresh: false,
	}

	var capturedReq *http.Request
	var capturedBody string
	client.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			capturedReq = req
			if req.Body != nil {
				bytes, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("failed to read request body: %v", err)
				}
				capturedBody = string(bytes)
			}
			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(strings.NewReader("{\"ok\":true}")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	client.SetToken("abc123")
	type payload struct {
		Name string `json:"name"`
	}

	resp, err := client.makeRequest(context.Background(), http.MethodPost, "/resources", payload{Name: "demo"})
	if err != nil {
		t.Fatalf("makeRequest returned error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	if capturedReq == nil {
		t.Fatalf("expected request to be captured")
	}
	if got := capturedReq.URL.String(); got != "http://example.com/v1/resources" {
		t.Fatalf("expected URL %q, got %q", "http://example.com/v1/resources", got)
	}
	if capturedReq.Method != http.MethodPost {
		t.Fatalf("expected method POST, got %s", capturedReq.Method)
	}
	if h := capturedReq.Header.Get("Content-Type"); h != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", h)
	}
	if h := capturedReq.Header.Get("Accept"); h != "application/json" {
		t.Fatalf("expected Accept application/json, got %q", h)
	}
	if h := capturedReq.Header.Get("Authorization"); h != "Bearer abc123" {
		t.Fatalf("expected Authorization header, got %q", h)
	}
	if capturedBody != "{\"name\":\"demo\"}" {
		t.Fatalf("expected body %q, got %q", "{\"name\":\"demo\"}", capturedBody)
	}
}

func TestKapuaClientMakeRequestMarshalError(t *testing.T) {
	client := &KapuaClient{
		baseURL:    "http://example.com/v1",
		httpClient: &http.Client{},
		logger:     utils.NewDefaultLogger("test"),
	}

	_, err := client.makeRequest(context.Background(), http.MethodPost, "/fail", make(chan struct{}))
	if err == nil {
		t.Fatal("expected error from unmarshalable body")
	}
	if !strings.Contains(err.Error(), "failed to marshal request body") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKapuaClientMakeRequestDoError(t *testing.T) {
	client := &KapuaClient{
		baseURL:     "http://example.com/v1",
		logger:      utils.NewDefaultLogger("test"),
		autoRefresh: false,
	}

	client.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("boom")
		}),
	}

	_, err := client.makeRequest(context.Background(), http.MethodGet, "/broken", nil)
	if err == nil {
		t.Fatal("expected error from transport")
	}
	if !strings.Contains(err.Error(), "failed to execute request") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKapuaClientHandleResponseSuccess(t *testing.T) {
	client := &KapuaClient{logger: utils.NewDefaultLogger("test")}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("{\"value\":\"ok\"}")),
		Header:     make(http.Header),
	}
	var result struct {
		Value string `json:"value"`
	}

	if err := client.handleResponse(resp, &result); err != nil {
		t.Fatalf("handleResponse returned error: %v", err)
	}
	if result.Value != "ok" {
		t.Fatalf("expected value 'ok', got %q", result.Value)
	}
}

func TestKapuaClientHandleResponseKapuaError(t *testing.T) {
	client := &KapuaClient{logger: utils.NewDefaultLogger("test")}
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader("{\"code\":\"ERR\",\"message\":\"bad\",\"details\":\"oops\"}")),
		Header:     make(http.Header),
	}

	err := client.handleResponse(resp, nil)
	if err == nil {
		t.Fatal("expected KapuaError")
	}
	var kapuaErr models.KapuaError
	if !errors.As(err, &kapuaErr) {
		t.Fatalf("expected KapuaError, got %T", err)
	}
	if kapuaErr.Message != "bad" || kapuaErr.Details != "oops" {
		t.Fatalf("unexpected KapuaError contents: %+v", kapuaErr)
	}
}

func TestKapuaClientHandleResponseKapuaErrorInvalidJSON(t *testing.T) {
	client := &KapuaClient{logger: utils.NewDefaultLogger("test")}
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader("not-json")),
		Header:     make(http.Header),
	}

	err := client.handleResponse(resp, nil)
	if err == nil {
		t.Fatal("expected error for invalid Kapua error JSON")
	}
	if !strings.Contains(err.Error(), "API request failed with status 500") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKapuaClientHandleResponseSuccessInvalidJSON(t *testing.T) {
	client := &KapuaClient{logger: utils.NewDefaultLogger("test")}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("not-json")),
		Header:     make(http.Header),
	}

	var out struct{}
	err := client.handleResponse(resp, &out)
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
	if !strings.Contains(err.Error(), "failed to unmarshal response") {
		t.Fatalf("unexpected error: %v", err)
	}
}
