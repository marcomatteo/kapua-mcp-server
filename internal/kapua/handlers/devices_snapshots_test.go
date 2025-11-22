package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"kapua-mcp-server/internal/kapua/models"
)

func TestHandleDeviceSnapshotsListSuccess(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tenant/devices/device-1/snapshots" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		payload := models.DeviceSnapshots{SnapshotID: []models.DeviceSnapshot{{ID: "snap-1", Timestamp: 1}}}
		data, _ := json.Marshal(payload)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})

	params := &DeviceSnapshotsParams{DeviceID: "device-1"}
	result, out, err := handler.HandleDeviceSnapshotsList(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleDeviceSnapshotsList returned error: %v", err)
	}
	if out == nil {
		t.Fatalf("expected snapshots payload")
	}
	if len(result.Content) != 2 {
		t.Fatalf("expected summary and payload contents, got %d", len(result.Content))
	}
	if summary := textContent(t, result.Content[0]); summary != "Retrieved 1 snapshots" {
		t.Fatalf("unexpected summary: %s", summary)
	}
	var decoded models.DeviceSnapshots
	if err := json.Unmarshal([]byte(textContent(t, result.Content[1])), &decoded); err != nil {
		t.Fatalf("failed to decode json content: %v", err)
	}
	if len(decoded.SnapshotID) != 1 || decoded.SnapshotID[0].ID != "snap-1" {
		t.Fatalf("unexpected decoded payload: %+v", decoded)
	}
}

func TestHandleDeviceSnapshotsListMissingDeviceID(t *testing.T) {
	handler := &KapuaHandler{}
	if _, _, err := handler.HandleDeviceSnapshotsList(context.Background(), nil, &DeviceSnapshotsParams{}); err == nil {
		t.Fatal("expected error for missing deviceId")
	}
}

func TestHandleDeviceSnapshotsListError(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("kapua error"))
	})

	_, _, err := handler.HandleDeviceSnapshotsList(context.Background(), nil, &DeviceSnapshotsParams{DeviceID: "device-1"})
	if err == nil || !strings.Contains(err.Error(), "failed to list device snapshots") {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestHandleDeviceSnapshotConfigurationsReadSuccess(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tenant/devices/device-1/snapshots/snap-1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"configuration":[{"id":"component-1"}]}`))
	})

	params := &DeviceSnapshotLookupParams{DeviceID: "device-1", SnapshotID: "snap-1"}
	result, out, err := handler.HandleDeviceSnapshotConfigurationsRead(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleDeviceSnapshotConfigurationsRead returned error: %v", err)
	}
	if out == nil {
		t.Fatalf("expected configuration payload")
	}
	if len(result.Content) != 2 {
		t.Fatalf("expected summary and payload contents, got %d", len(result.Content))
	}
	if summary := textContent(t, result.Content[0]); summary != "Retrieved 1 component configurations from snapshot snap-1" {
		t.Fatalf("unexpected summary: %s", summary)
	}
}

func TestHandleDeviceSnapshotConfigurationsReadMissingParams(t *testing.T) {
	handler := &KapuaHandler{}
	if _, _, err := handler.HandleDeviceSnapshotConfigurationsRead(context.Background(), nil, nil); err == nil {
		t.Fatal("expected error for nil params")
	}
	if _, _, err := handler.HandleDeviceSnapshotConfigurationsRead(context.Background(), nil, &DeviceSnapshotLookupParams{SnapshotID: "snap-1"}); err == nil {
		t.Fatal("expected error for missing deviceId")
	}
	if _, _, err := handler.HandleDeviceSnapshotConfigurationsRead(context.Background(), nil, &DeviceSnapshotLookupParams{DeviceID: "device-1"}); err == nil {
		t.Fatal("expected error for missing snapshotId")
	}
}

func TestHandleDeviceSnapshotConfigurationsReadError(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("kapua error"))
	})

	_, _, err := handler.HandleDeviceSnapshotConfigurationsRead(context.Background(), nil, &DeviceSnapshotLookupParams{DeviceID: "device-1", SnapshotID: "snap-1"})
	if err == nil || !strings.Contains(err.Error(), "failed to read device snapshot configurations") {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestHandleDeviceSnapshotRollbackSuccess(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tenant/devices/device-1/snapshots/snap-1/_rollback" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	params := &DeviceSnapshotLookupParams{DeviceID: "device-1", SnapshotID: "snap-1"}
	result, meta, err := handler.HandleDeviceSnapshotRollback(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleDeviceSnapshotRollback returned error: %v", err)
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected single summary content, got %d", len(result.Content))
	}
	if summary := textContent(t, result.Content[0]); summary != "Rollback to snapshot snap-1 requested for device device-1" {
		t.Fatalf("unexpected summary: %s", summary)
	}
	metaMap, ok := meta.(map[string]string)
	if !ok || metaMap["status"] != "requested" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
}

func TestHandleDeviceSnapshotRollbackMissingParams(t *testing.T) {
	handler := &KapuaHandler{}
	if _, _, err := handler.HandleDeviceSnapshotRollback(context.Background(), nil, nil); err == nil {
		t.Fatal("expected error for nil params")
	}
	if _, _, err := handler.HandleDeviceSnapshotRollback(context.Background(), nil, &DeviceSnapshotLookupParams{SnapshotID: "snap-1"}); err == nil {
		t.Fatal("expected error for missing deviceId")
	}
	if _, _, err := handler.HandleDeviceSnapshotRollback(context.Background(), nil, &DeviceSnapshotLookupParams{DeviceID: "device-1"}); err == nil {
		t.Fatal("expected error for missing snapshotId")
	}
}

func TestHandleDeviceSnapshotRollbackError(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("kapua error"))
	})

	_, _, err := handler.HandleDeviceSnapshotRollback(context.Background(), nil, &DeviceSnapshotLookupParams{DeviceID: "device-1", SnapshotID: "snap-1"})
	if err == nil || !strings.Contains(err.Error(), "failed to rollback device snapshot") {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}
