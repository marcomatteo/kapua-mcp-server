package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"kapua-mcp-server/internal/kapua/models"
)

func TestHandleDeviceInventoryReadSuccess(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tenant/devices/device-1/inventory" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		payload := models.DeviceInventory{InventoryItems: []models.InventoryItem{{Name: "pkg", Version: "1.0"}}}
		data, _ := json.Marshal(payload)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})

	params := &DeviceInventoryParams{DeviceID: "device-1"}
	result, out, err := handler.HandleDeviceInventoryRead(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleDeviceInventoryRead returned error: %v", err)
	}
	if out == nil {
		t.Fatalf("expected inventory payload")
	}
	if len(result.Content) != 2 {
		t.Fatalf("expected summary and payload contents, got %d", len(result.Content))
	}
	if summary := textContent(t, result.Content[0]); summary != "Retrieved 1 inventory items" {
		t.Fatalf("unexpected summary: %s", summary)
	}
	var decoded models.DeviceInventory
	if err := json.Unmarshal([]byte(textContent(t, result.Content[1])), &decoded); err != nil {
		t.Fatalf("failed to decode json content: %v", err)
	}
	if len(decoded.InventoryItems) != 1 || decoded.InventoryItems[0].Name != "pkg" {
		t.Fatalf("unexpected decoded payload: %+v", decoded)
	}
}

func TestHandleDeviceInventoryReadMissingDeviceID(t *testing.T) {
	handler := &KapuaHandler{}
	if _, _, err := handler.HandleDeviceInventoryRead(context.Background(), nil, &DeviceInventoryParams{}); err == nil {
		t.Fatal("expected error for missing deviceId")
	}
}

func TestHandleDeviceInventoryReadError(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("kapua error"))
	})

	_, _, err := handler.HandleDeviceInventoryRead(context.Background(), nil, &DeviceInventoryParams{DeviceID: "device-1"})
	if err == nil || !strings.Contains(err.Error(), "failed to read device inventory") {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestHandleDeviceInventoryBundleStart(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tenant/devices/device-1/inventory/bundles/_start" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		var payload models.DeviceInventoryBundle
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if payload.ID != "bundle-1" {
			t.Fatalf("unexpected bundle payload: %+v", payload)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	params := &DeviceInventoryBundleActionParams{DeviceID: "device-1", Bundle: models.DeviceInventoryBundle{ID: "bundle-1"}}
	result, meta, err := handler.HandleDeviceInventoryBundleStart(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleDeviceInventoryBundleStart returned error: %v", err)
	}
	if textContent(t, result.Content[0]) != "Bundle inventory start requested" {
		t.Fatalf("unexpected content: %s", textContent(t, result.Content[0]))
	}
	status, ok := meta.(map[string]string)
	if !ok || status["status"] != "requested" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
}
