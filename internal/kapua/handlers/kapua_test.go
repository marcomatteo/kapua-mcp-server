package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"kapua-mcp-server/internal/kapua/config"
	"kapua-mcp-server/internal/kapua/models"
	"kapua-mcp-server/internal/kapua/services"
	"kapua-mcp-server/pkg/utils"
)

func newHandlerWithServer(t *testing.T, handler http.HandlerFunc) *KapuaHandler {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	client := services.NewKapuaClient(&config.KapuaConfig{APIEndpoint: ts.URL, Timeout: 5})
	client.SetTokenInfo(&models.AccessToken{KapuaEntity: models.KapuaEntity{ScopeID: models.KapuaID("tenant")}})
	return &KapuaHandler{client: client, logger: utils.NewDefaultLogger("KapuaHandlerTest")}
}

func TestNewKapuaHandlerListResources(t *testing.T) {
	handler := NewKapuaHandler(&services.KapuaClient{})

	resources, err := handler.ListResources(context.Background())
	if err != nil {
		t.Fatalf("ListResources returned error: %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("expected one resource, got %d", len(resources))
	}
	if resources[0].URI != "kapua://devices" {
		t.Fatalf("unexpected resource URI %s", resources[0].URI)
	}
}

func TestReadResourceDevicesSuccess(t *testing.T) {
	handler := newHandlerWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tenant/devices" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if limit := r.URL.Query().Get("limit"); limit != "100" {
			t.Fatalf("expected limit=100, got %s", limit)
		}
		if offset := r.URL.Query().Get("offset"); offset != "0" {
			t.Fatalf("expected offset=0, got %s", offset)
		}
		if ask := r.URL.Query().Get("askTotalCount"); ask != "true" {
			t.Fatalf("expected askTotalCount=true, got %s", ask)
		}
		result := models.DeviceListResult{
			Items:      []models.Device{{KapuaEntity: models.KapuaEntity{ID: models.KapuaID("device-1")}, ClientID: "client-1"}},
			Size:       1,
			TotalCount: 1,
		}
		body, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("failed to marshal result: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})

	result, err := handler.ReadResource(context.Background(), "kapua://devices")
	if err != nil {
		t.Fatalf("ReadResource returned error: %v", err)
	}
	if result == nil || len(result.Contents) != 1 {
		t.Fatalf("expected single resource content, got %+v", result)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &payload); err != nil {
		t.Fatalf("failed to unmarshal resource content: %v", err)
	}

	if count, ok := payload["total_count"].(float64); !ok || count != 1 {
		t.Fatalf("expected total_count 1, got %v", payload["total_count"])
	}
	devices, ok := payload["devices"].([]any)
	if !ok || len(devices) != 1 {
		t.Fatalf("expected one device entry, got %+v", payload["devices"])
	}
	if processed, ok := payload["processed_count"].(float64); !ok || processed != 1 {
		t.Fatalf("expected processed_count 1, got %v", payload["processed_count"])
	}
	if _, err := strconv.ParseInt(payload["last_updated"].(string), 10, 64); err != nil {
		t.Fatalf("expected numeric last_updated, got %v", payload["last_updated"])
	}
}

func TestReadResourceUnknown(t *testing.T) {
	handler := &KapuaHandler{logger: utils.NewDefaultLogger("test")}

	if _, err := handler.ReadResource(context.Background(), "kapua://unknown"); err == nil {
		t.Fatal("expected error for unknown resource")
	}
}

func TestReadDevicesResourceErrorPropagates(t *testing.T) {
	handler := newHandlerWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("kapua error"))
	})

	_, err := handler.ReadResource(context.Background(), "kapua://devices")
	if err == nil || !strings.Contains(err.Error(), "failed to read devices resource") {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}
