package handlers

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"kapua-mcp-server/internal/config"
	"kapua-mcp-server/internal/kapua/models"
	"kapua-mcp-server/internal/kapua/services"
	"kapua-mcp-server/pkg/utils"
)

func newDeviceHandler(t *testing.T, fn http.HandlerFunc) *KapuaHandler {
	t.Helper()
	ts := httptest.NewServer(fn)
	t.Cleanup(ts.Close)
	client := services.NewKapuaClient(&config.KapuaConfig{APIEndpoint: ts.URL, Timeout: 5})
	client.SetTokenInfo(&models.AccessToken{KapuaEntity: models.KapuaEntity{ScopeID: models.KapuaID("tenant")}})
	return &KapuaHandler{client: client, logger: utils.NewDefaultLogger("KapuaDeviceHandlerTest")}
}

func textContent(t *testing.T, content mcp.Content) string {
	t.Helper()
	txt, ok := content.(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", content)
	}
	return txt.Text
}

func TestHandleListDevicesSuccess(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tenant/devices" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("clientId") != "acme" {
			t.Fatalf("expected clientId acme, got %s", r.URL.Query().Get("clientId"))
		}
		if r.URL.Query().Get("status") != "ENABLED" {
			t.Fatalf("expected status ENABLED, got %s", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("matchTerm") != "sensor" {
			t.Fatalf("expected matchTerm sensor, got %s", r.URL.Query().Get("matchTerm"))
		}
		if r.URL.Query().Get("limit") != "25" {
			t.Fatalf("expected limit 25, got %s", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("offset") != "5" {
			t.Fatalf("expected offset 5, got %s", r.URL.Query().Get("offset"))
		}

		payload := models.DeviceListResult{
			Items: []models.Device{{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-1")}, ClientID: "client-1"}},
			Size:  1,
		}
		data, _ := json.Marshal(payload)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})

	params := &ListDevicesParams{ClientID: "acme", Status: "ENABLED", MatchTerm: "sensor", Limit: 25, Offset: 5}
	result, _, err := handler.HandleListDevices(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleListDevices returned error: %v", err)
	}
	if len(result.Content) != 2 {
		t.Fatalf("expected two content entries, got %d", len(result.Content))
	}

	summary := textContent(t, result.Content[0])
	if summary != "Found 1 devices." {
		t.Fatalf("unexpected summary: %s", summary)
	}

	var body models.DeviceListResult
	if err := json.Unmarshal([]byte(textContent(t, result.Content[1])), &body); err != nil {
		t.Fatalf("failed to unmarshal json content: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].ClientID != "client-1" {
		t.Fatalf("unexpected body payload: %+v", body)
	}
}

func TestHandleListDevicesServiceError(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"kapua error"}`))
	})

	_, _, err := handler.HandleListDevices(context.Background(), nil, &ListDevicesParams{})
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); !strings.Contains(got, "failed to list devices") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleUpdateDeviceSuccess(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tenant/devices/device-1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		var payload models.Device
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}
		if payload.ClientID != "client-1" {
			t.Fatalf("expected client ID client-1, got %s", payload.ClientID)
		}

		payload.DisplayName = "Updated"
		data, _ := json.Marshal(payload)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})

	params := &UpdateDeviceParams{
		DeviceID: "device-1",
		Device: map[string]any{
			"clientId":    "client-1",
			"displayName": "Sensor",
		},
	}

	result, updated, err := handler.HandleUpdateDevice(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleUpdateDevice returned error: %v", err)
	}
	if updated == nil {
		t.Fatalf("expected updated device data")
	}

	summary := textContent(t, result.Content[0])
	if !strings.Contains(summary, "Updated device client-1") {
		t.Fatalf("unexpected summary: %s", summary)
	}

	device, ok := updated.(*models.Device)
	if !ok {
		t.Fatalf("expected device type, got %T", updated)
	}
	if device.DisplayName != "Updated" {
		t.Fatalf("expected updated display name, got %s", device.DisplayName)
	}
}

func TestHandleUpdateDeviceEncodeError(t *testing.T) {
	handler := &KapuaHandler{logger: utils.NewDefaultLogger("test")}
	params := &UpdateDeviceParams{Device: map[string]any{"invalid": math.Inf(1)}}

	_, _, err := handler.HandleUpdateDevice(context.Background(), nil, params)
	if err == nil || !strings.Contains(err.Error(), "failed to encode device payload") {
		t.Fatalf("expected encode error, got %v", err)
	}
}

func TestHandleUpdateDeviceInvalidPayload(t *testing.T) {
	handler := &KapuaHandler{logger: utils.NewDefaultLogger("test")}
	params := &UpdateDeviceParams{Device: map[string]any{"status": 123}}

	_, _, err := handler.HandleUpdateDevice(context.Background(), nil, params)
	if err == nil || !strings.Contains(err.Error(), "invalid device payload") {
		t.Fatalf("expected invalid payload error, got %v", err)
	}
}

func TestHandleUpdateDeviceServiceError(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("bad gateway"))
	})

	params := &UpdateDeviceParams{DeviceID: "device-1", Device: map[string]any{"clientId": "client"}}
	_, _, err := handler.HandleUpdateDevice(context.Background(), nil, params)
	if err == nil || !strings.Contains(err.Error(), "failed to update device") {
		t.Fatalf("expected update error, got %v", err)
	}
}

func TestHandleDeleteDeviceSuccess(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tenant/devices/device-1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	result, data, err := handler.HandleDeleteDevice(context.Background(), nil, &DeleteDeviceParams{DeviceID: "device-1"})
	if err != nil {
		t.Fatalf("HandleDeleteDevice returned error: %v", err)
	}

	summary := textContent(t, result.Content[0])
	if summary != "Deleted device ID device-1" {
		t.Fatalf("unexpected summary %s", summary)
	}

	meta, ok := data.(map[string]string)
	if !ok || meta["status"] != "deleted" {
		t.Fatalf("unexpected metadata: %+v", data)
	}
}

func TestHandleDeleteDeviceServiceError(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, _, err := handler.HandleDeleteDevice(context.Background(), nil, &DeleteDeviceParams{DeviceID: "device-1"})
	if err == nil || !strings.Contains(err.Error(), "failed to delete device") {
		t.Fatalf("expected delete error, got %v", err)
	}
}
