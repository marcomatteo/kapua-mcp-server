package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"kapua-mcp-server/internal/kapua/models"
	"kapua-mcp-server/pkg/utils"
)

func newConfigHandler(t *testing.T, fn http.HandlerFunc) *KapuaHandler {
	return newKapuaTestHandler(t, fn, "KapuaConfigHandlerTest")
}

func TestHandleDeviceConfigurationsReadSuccess(t *testing.T) {
	handler := newConfigHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tenant/devices/device-1/configurations" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		payload := models.DeviceConfiguration{Configuration: []models.ComponentConfiguration{{ID: "comp"}}}
		data, _ := json.Marshal(payload)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})

	result, out, err := handler.HandleDeviceConfigurationsRead(context.Background(), nil, &DeviceID{DeviceID: "device-1"})
	if err != nil {
		t.Fatalf("HandleDeviceConfigurationsRead returned error: %v", err)
	}
	if out == nil {
		t.Fatalf("expected configuration data")
	}

	txt := textContent(t, result.Content[0])
	var decoded models.DeviceConfiguration
	if err := json.Unmarshal([]byte(txt), &decoded); err != nil {
		t.Fatalf("failed to decode json content: %v", err)
	}
	if len(decoded.Configuration) != 1 || decoded.Configuration[0].ID != "comp" {
		t.Fatalf("unexpected configuration: %+v", decoded)
	}
}

func TestHandleDeviceConfigurationsReadMissingID(t *testing.T) {
	handler := &KapuaHandler{logger: utils.NewDefaultLogger("test")}
	if _, _, err := handler.HandleDeviceConfigurationsRead(context.Background(), nil, &DeviceID{}); err == nil {
		t.Fatal("expected error for missing deviceId")
	}
}

func TestHandleDeviceConfigurationsReadError(t *testing.T) {
	handler := newConfigHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("kapua error"))
	})

	_, _, err := handler.HandleDeviceConfigurationsRead(context.Background(), nil, &DeviceID{DeviceID: "device-1"})
	if err == nil || !strings.Contains(err.Error(), "failed to read device configurations") {
		t.Fatalf("expected failure, got %v", err)
	}
}

func TestHandleDeviceConfigurationsWriteSuccess(t *testing.T) {
	handler := newConfigHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tenant/devices/device-1/configurations" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	params := &DeviceConfigurationsWriteParams{
		Device:  models.Device{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-1")}},
		Payload: map[string]any{"configuration": []string{"value"}},
	}

	result, meta, err := handler.HandleDeviceConfigurationsWrite(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleDeviceConfigurationsWrite returned error: %v", err)
	}
	if textContent(t, result.Content[0]) != "Updated configurations for device device-1" {
		t.Fatalf("unexpected summary %s", textContent(t, result.Content[0]))
	}
	status, ok := meta.(map[string]string)
	if !ok || status["status"] != "updated" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
}

func TestHandleDeviceConfigurationsWriteError(t *testing.T) {
	handler := newConfigHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad"))
	})

	params := &DeviceConfigurationsWriteParams{Device: models.Device{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-1")}}, Payload: map[string]any{}}
	_, _, err := handler.HandleDeviceConfigurationsWrite(context.Background(), nil, params)
	if err == nil || !strings.Contains(err.Error(), "failed to write device configurations") {
		t.Fatalf("expected write error, got %v", err)
	}
}

func TestHandleDeviceComponentConfigurationReadSuccess(t *testing.T) {
	handler := newConfigHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tenant/devices/device-1/configurations/service-1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		payload := models.DeviceConfiguration{Configuration: []models.ComponentConfiguration{{ID: "service-1"}}}
		data, _ := json.Marshal(payload)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})

	params := &DeviceComponentConfigurationReadParams{Device: models.Device{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-1")}}, ComponentID: "service-1"}
	result, out, err := handler.HandleDeviceComponentConfigurationRead(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleDeviceComponentConfigurationRead returned error: %v", err)
	}
	if out == nil {
		t.Fatalf("expected configuration data")
	}
	var decoded models.DeviceConfiguration
	if err := json.Unmarshal([]byte(textContent(t, result.Content[0])), &decoded); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	if len(decoded.Configuration) != 1 || decoded.Configuration[0].ID != "service-1" {
		t.Fatalf("unexpected configuration: %+v", decoded)
	}
}

func TestHandleDeviceComponentConfigurationReadError(t *testing.T) {
	handler := newConfigHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("kapua error"))
	})

	params := &DeviceComponentConfigurationReadParams{Device: models.Device{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-1")}}, ComponentID: "service-1"}
	_, _, err := handler.HandleDeviceComponentConfigurationRead(context.Background(), nil, params)
	if err == nil || !strings.Contains(err.Error(), "failed to read device component configuration") {
		t.Fatalf("expected error, got %v", err)
	}
}

func TestHandleDeviceComponentConfigurationWriteSuccess(t *testing.T) {
	handler := newConfigHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tenant/devices/device-1/configurations/service-1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	params := &DeviceComponentConfigurationWriteParams{
		Device:      models.Device{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-1")}},
		ComponentID: "service-1",
		Payload:     map[string]any{"configuration": []string{"value"}},
	}

	result, meta, err := handler.HandleDeviceComponentConfigurationWrite(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleDeviceComponentConfigurationWrite returned error: %v", err)
	}
	if textContent(t, result.Content[0]) != "Updated component service-1 configuration for device device-1" {
		t.Fatalf("unexpected summary: %s", textContent(t, result.Content[0]))
	}
	status, ok := meta.(map[string]string)
	if !ok || status["status"] != "updated" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
}

func TestHandleDeviceComponentConfigurationWriteError(t *testing.T) {
	handler := newConfigHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("kapua error"))
	})

	params := &DeviceComponentConfigurationWriteParams{Device: models.Device{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-1")}}, ComponentID: "service-1"}
	_, _, err := handler.HandleDeviceComponentConfigurationWrite(context.Background(), nil, params)
	if err == nil || !strings.Contains(err.Error(), "failed to write device component configuration") {
		t.Fatalf("expected error, got %v", err)
	}
}
