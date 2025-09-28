package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"kapua-mcp-server/internal/kapua/models"
)

func TestHandleListDeviceLogsSuccess(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tenant/deviceLogs" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		expected := map[string]string{
			"clientId":        "client-1",
			"channel":         "/foo",
			"strictChannel":   "true",
			"logPropertyName": "MESSAGE",
			"logPropertyType": "string",
			"logPropertyMin":  "A",
			"logPropertyMax":  "Z",
			"sortDir":         "DESCENDING",
			"limit":           "25",
			"offset":          "5",
		}
		for key, value := range expected {
			if r.URL.Query().Get(key) != value {
				t.Fatalf("expected %s=%s, got %s", key, value, r.URL.Query().Get(key))
			}
		}

		payload := models.DeviceLogListResult{
			Items: []models.DeviceLog{{StoreID: "store-1", ClientID: "client-1"}},
			Size:  1,
		}
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})

	params := &ListDeviceLogsParams{
		ClientID:        "client-1",
		Channel:         "/foo",
		StrictChannel:   boolPtr(true),
		LogPropertyName: "MESSAGE",
		LogPropertyType: "string",
		LogPropertyMin:  "A",
		LogPropertyMax:  "Z",
		SortDir:         "DESCENDING",
		Limit:           intPtr(25),
		Offset:          intPtr(5),
	}

	result, data, err := handler.HandleListDeviceLogs(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleListDeviceLogs returned error: %v", err)
	}

	summary := textContent(t, result.Content[0])
	if summary != "Found 1 device logs." {
		t.Fatalf("unexpected summary: %s", summary)
	}

	var list models.DeviceLogListResult
	if err := json.Unmarshal([]byte(textContent(t, result.Content[1])), &list); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].StoreID != "store-1" {
		t.Fatalf("unexpected list payload: %+v", list)
	}

	typed, ok := data.(*models.DeviceLogListResult)
	if !ok || len(typed.Items) != 1 {
		t.Fatalf("unexpected data: %+v", data)
	}
}

func TestHandleListDeviceLogsNoParams(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Fatalf("expected empty query, got %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	result, data, err := handler.HandleListDeviceLogs(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("HandleListDeviceLogs returned error: %v", err)
	}

	if summary := textContent(t, result.Content[0]); summary != "Found 0 device logs." {
		t.Fatalf("unexpected summary: %s", summary)
	}

	typed, ok := data.(*models.DeviceLogListResult)
	if !ok || len(typed.Items) != 0 {
		t.Fatalf("unexpected result: %+v", data)
	}
}

func TestHandleListDeviceLogsServiceError(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code":"ERR"}`))
	})

	_, _, err := handler.HandleListDeviceLogs(context.Background(), nil, &ListDeviceLogsParams{})
	if err == nil || !strings.Contains(err.Error(), "failed to list device logs") {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}
