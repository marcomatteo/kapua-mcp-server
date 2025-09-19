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

func TestHandleListDeviceEventsSuccess(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tenant/devices/device-123/events" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("resource") != "LOG" {
			t.Fatalf("expected resource filter LOG, got %s", query.Get("resource"))
		}
		if query.Get("startDate") != "2023-03-10T00:00:00Z" {
			t.Fatalf("expected startDate filter, got %s", query.Get("startDate"))
		}
		if query.Get("endDate") != "2023-03-11T00:00:00Z" {
			t.Fatalf("expected endDate filter, got %s", query.Get("endDate"))
		}
		if query.Get("sortParam") != "sentOn" {
			t.Fatalf("expected sortParam sentOn, got %s", query.Get("sortParam"))
		}
		if query.Get("sortDir") != "DESCENDING" {
			t.Fatalf("expected sortDir DESCENDING, got %s", query.Get("sortDir"))
		}
		if query.Get("askTotalCount") != "true" {
			t.Fatalf("expected askTotalCount true, got %s", query.Get("askTotalCount"))
		}
		if query.Get("limit") != "10" {
			t.Fatalf("expected limit 10, got %s", query.Get("limit"))
		}
		if query.Get("offset") != "5" {
			t.Fatalf("expected offset 5, got %s", query.Get("offset"))
		}

		payload := models.DeviceEventListResult{
			Items: []models.DeviceEvent{
				{
					KapuaEntity:  models.KapuaEntity{ID: models.KapuaID("event-1"), ScopeID: models.KapuaID("tenant")},
					DeviceID:     models.KapuaID("device-123"),
					EventMessage: "Device started",
					Resource:     "LOG",
				},
			},
			Size:       1,
			TotalCount: 25,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})

	askTotal := true
	params := &ListDeviceEventsParams{
		DeviceID:      "device-123",
		Resource:      "LOG",
		StartDate:     "2023-03-10T00:00:00Z",
		EndDate:       "2023-03-11T00:00:00Z",
		SortParam:     "sentOn",
		SortDir:       "DESCENDING",
		AskTotalCount: &askTotal,
		Limit:         10,
		Offset:        5,
	}

	result, data, err := handler.HandleListDeviceEvents(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleListDeviceEvents returned error: %v", err)
	}

	if len(result.Content) != 2 {
		t.Fatalf("expected two content entries, got %d", len(result.Content))
	}

	summary := textContent(t, result.Content[0])
	if summary != "Found 1 device events (total count: 25)." {
		t.Fatalf("unexpected summary: %s", summary)
	}

	var body models.DeviceEventListResult
	if err := json.Unmarshal([]byte(textContent(t, result.Content[1])), &body); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].EventMessage != "Device started" {
		t.Fatalf("unexpected payload: %+v", body)
	}

	typed, ok := data.(*models.DeviceEventListResult)
	if !ok {
		t.Fatalf("expected *models.DeviceEventListResult, got %T", data)
	}
	if typed.TotalCount != 25 {
		t.Fatalf("expected totalCount 25, got %d", typed.TotalCount)
	}
}

func TestHandleListDeviceEventsServiceError(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"kapua error"}`))
	})

	_, _, err := handler.HandleListDeviceEvents(context.Background(), nil, &ListDeviceEventsParams{DeviceID: "device-123"})
	if err == nil || !strings.Contains(err.Error(), "failed to list device events") {
		t.Fatalf("expected wrapped service error, got %v", err)
	}
}

func TestHandleListDeviceEventsMissingParams(t *testing.T) {
	handler := &KapuaHandler{logger: utils.NewDefaultLogger("test")}

	if _, _, err := handler.HandleListDeviceEvents(context.Background(), nil, nil); err == nil || !strings.Contains(err.Error(), "parameters are required") {
		t.Fatalf("expected missing params error, got %v", err)
	}

	if _, _, err := handler.HandleListDeviceEvents(context.Background(), nil, &ListDeviceEventsParams{}); err == nil || !strings.Contains(err.Error(), "deviceId is required") {
		t.Fatalf("expected missing deviceId error, got %v", err)
	}
}
