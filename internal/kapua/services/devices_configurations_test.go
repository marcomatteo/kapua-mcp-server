package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"kapua-mcp-server/internal/kapua/models"
	"kapua-mcp-server/pkg/utils"
)

type configRoundTripFunc func(*http.Request) (*http.Response, error)

func (f configRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestKapuaClient() *KapuaClient {
	return &KapuaClient{
		baseURL:     "http://example.com/v1",
		scopeId:     "tenant",
		logger:      utils.NewDefaultLogger("test"),
		autoRefresh: false,
	}
}

func TestReadDeviceConfigurationsSuccess(t *testing.T) {
	client := newTestKapuaClient()

	sampleResp := `{"configuration":[{"id":"component-1","definition":{"id":"org.eclipse.kura.sample","name":"Sample Component"},"properties":{"property":[{"name":"enabled","type":"BOOLEAN","value":["true"]}]}}]}`

	client.httpClient = &http.Client{Transport: configRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/devices/device-123/configurations" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(sampleResp)),
			Header:     make(http.Header),
		}, nil
	})}

	cfg, err := client.ReadDeviceConfigurations(context.Background(), "device-123")
	if err != nil {
		t.Fatalf("ReadDeviceConfigurations returned error: %v", err)
	}
	if cfg == nil || len(cfg.Configuration) != 1 || cfg.Configuration[0].ID != "component-1" {
		t.Fatalf("unexpected configuration payload: %+v", cfg)
	}
}

func TestReadDeviceConfigurationsRequestError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: configRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})}

	_, err := client.ReadDeviceConfigurations(context.Background(), "device-123")
	if err == nil || !strings.Contains(err.Error(), "read device configurations request failed") {
		t.Fatalf("expected wrapped request error, got %v", err)
	}
}

func TestReadDeviceConfigurationsHandleError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: configRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("kapua boom")),
			Header:     make(http.Header),
		}, nil
	})}

	_, err := client.ReadDeviceConfigurations(context.Background(), "device-123")
	if err == nil || !strings.Contains(err.Error(), "failed to read device configurations") {
		t.Fatalf("expected response error, got %v", err)
	}
}

func TestWriteDeviceConfigurationsSuccess(t *testing.T) {
	client := newTestKapuaClient()
	device := models.Device{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-123")}}
	payload := map[string]interface{}{
		"configuration": []map[string]interface{}{
			{
				"id": "component-1",
				"properties": map[string]interface{}{
					"property": []map[string]interface{}{{
						"name":  "poll.interval",
						"type":  "INTEGER",
						"value": []string{"60"},
					}},
				},
			},
		},
	}

	client.httpClient = &http.Client{Transport: configRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/devices/device-123/configurations" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		configs, ok := body["configuration"].([]interface{})
		if !ok || len(configs) != 1 {
			t.Fatalf("unexpected configuration body: %+v", body)
		}
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})}

	if err := client.WriteDeviceConfigurations(context.Background(), device, payload); err != nil {
		t.Fatalf("WriteDeviceConfigurations returned error: %v", err)
	}
}

func TestWriteDeviceConfigurationsRequestError(t *testing.T) {
	client := newTestKapuaClient()
	device := models.Device{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-123")}}

	client.httpClient = &http.Client{Transport: configRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("timeout")
	})}

	err := client.WriteDeviceConfigurations(context.Background(), device, map[string]string{})
	if err == nil || !strings.Contains(err.Error(), "write device configurations request failed") {
		t.Fatalf("expected wrapped request error, got %v", err)
	}
}

func TestWriteDeviceConfigurationsHandleError(t *testing.T) {
	client := newTestKapuaClient()
	device := models.Device{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-123")}}

	client.httpClient = &http.Client{Transport: configRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader("kapua error")),
			Header:     make(http.Header),
		}, nil
	})}

	err := client.WriteDeviceConfigurations(context.Background(), device, map[string]string{})
	if err == nil || !strings.Contains(err.Error(), "failed to write device configurations") {
		t.Fatalf("expected response error, got %v", err)
	}
}

func TestReadDeviceComponentConfigurationSuccess(t *testing.T) {
	client := newTestKapuaClient()
	device := models.Device{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-123")}}

	sampleResp := `{"configuration":[{"id":"component-1"}]}`

	client.httpClient = &http.Client{Transport: configRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/devices/device-123/configurations/service-1" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(sampleResp)),
			Header:     make(http.Header),
		}, nil
	})}

	cfg, err := client.ReadDeviceComponentConfiguration(context.Background(), device, "service-1")
	if err != nil {
		t.Fatalf("ReadDeviceComponentConfiguration returned error: %v", err)
	}
	if cfg == nil || len(cfg.Configuration) != 1 {
		t.Fatalf("unexpected configuration: %+v", cfg)
	}
}

func TestWriteDeviceComponentConfigurationSuccess(t *testing.T) {
	client := newTestKapuaClient()
	device := models.Device{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-123")}}
	payload := map[string]interface{}{"configuration": []string{"value"}}

	client.httpClient = &http.Client{Transport: configRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/devices/device-123/configurations/service-1" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})}

	if err := client.WriteDeviceComponentConfiguration(context.Background(), device, "service-1", payload); err != nil {
		t.Fatalf("WriteDeviceComponentConfiguration returned error: %v", err)
	}
}

func TestWriteDeviceComponentConfigurationHandleError(t *testing.T) {
	client := newTestKapuaClient()
	device := models.Device{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-123")}}

	client.httpClient = &http.Client{Transport: configRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Body:       io.NopCloser(strings.NewReader("nope")),
			Header:     make(http.Header),
		}, nil
	})}

	err := client.WriteDeviceComponentConfiguration(context.Background(), device, "service-1", map[string]string{})
	if err == nil || !strings.Contains(err.Error(), "failed to write device component configuration") {
		t.Fatalf("expected response error, got %v", err)
	}
}

func TestReadDeviceComponentConfigurationRequestError(t *testing.T) {
	client := newTestKapuaClient()
	device := models.Device{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-123")}}

	client.httpClient = &http.Client{Transport: configRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("dns lookup failed")
	})}

	_, err := client.ReadDeviceComponentConfiguration(context.Background(), device, "service-1")
	if err == nil || !strings.Contains(err.Error(), "read device component configuration request failed") {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestWriteDeviceComponentConfigurationRequestError(t *testing.T) {
	client := newTestKapuaClient()
	device := models.Device{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-123")}}

	client.httpClient = &http.Client{Transport: configRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("socket timeout")
	})}

	err := client.WriteDeviceComponentConfiguration(context.Background(), device, "service-1", map[string]string{})
	if err == nil || !strings.Contains(err.Error(), "write device component configuration request failed") {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}
