package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"kapua-mcp-server/internal/kapua/models"
)

func TestReadFleetHealthResource(t *testing.T) {
	fixedNow := time.Date(2024, 8, 1, 12, 0, 0, 0, time.UTC)
	originalNow := timeNow
	timeNow = func() time.Time { return fixedNow }
	defer func() { timeNow = originalNow }()

	handler := newHandlerWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/tenant/devices":
			if q := r.URL.Query(); q.Get("limit") != "5" || q.Get("askTotalCount") != "true" {
				t.Fatalf("unexpected query params: %v", q)
			}
			result := models.DeviceListResult{
				TotalCount: 3,
				Items: []models.Device{
					{
						KapuaEntity: models.KapuaEntity{ID: models.KapuaID("dev-1")},
						ClientID:    "alpha",
						Connection: &models.DeviceConnection{
							Status:     models.ConnectionStatusConnected,
							ModifiedOn: ptrTime(fixedNow.Add(-5 * time.Minute)),
						},
						LastEvent: &models.DeviceEvent{ReceivedOn: fixedNow.Add(-5 * time.Minute)},
					},
					{
						KapuaEntity: models.KapuaEntity{ID: models.KapuaID("dev-2")},
						ClientID:    "bravo",
						Connection: &models.DeviceConnection{
							Status: models.ConnectionStatusDisconnected,
						},
						LastEvent: &models.DeviceEvent{ReceivedOn: fixedNow.Add(-2 * time.Hour)},
					},
					{
						KapuaEntity: models.KapuaEntity{ID: models.KapuaID("dev-3")},
						ClientID:    "charlie",
						Connection: &models.DeviceConnection{
							Status:     models.ConnectionStatusMissing,
							ModifiedOn: ptrTime(fixedNow.Add(-3 * time.Hour)),
						},
					},
				},
			}
			body, _ := json.Marshal(result)
			_, _ = w.Write(body)

		case strings.HasSuffix(r.URL.Path, "/devices/dev-1/events"):
			result := models.DeviceEventListResult{
				Items: []models.DeviceEvent{
					{Action: "CRITICAL", EventMessage: "CRITICAL temp spike", ReceivedOn: fixedNow.Add(-30 * time.Minute)},
					{Action: "INFO", EventMessage: "heartbeat ok", ReceivedOn: fixedNow.Add(-10 * time.Minute)},
				},
			}
			body, _ := json.Marshal(result)
			_, _ = w.Write(body)

		case strings.HasSuffix(r.URL.Path, "/devices/dev-2/events"):
			result := models.DeviceEventListResult{
				Items: []models.DeviceEvent{
					{Action: "INFO", EventMessage: "deploy complete", ReceivedOn: fixedNow.Add(-20 * time.Minute)},
				},
			}
			body, _ := json.Marshal(result)
			_, _ = w.Write(body)

		case strings.HasSuffix(r.URL.Path, "/devices/dev-3/events"):
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`kapua error`))

		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	})

	result, err := handler.ReadResource(context.Background(), "kapua://fleet-health?staleMinutes=90&criticalMinutes=120&limit=5")
	if err != nil {
		t.Fatalf("ReadResource returned error: %v", err)
	}
	if result == nil || len(result.Contents) != 1 {
		t.Fatalf("expected single content entry, got %+v", result)
	}

	var report fleetHealthReport
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &report); err != nil {
		t.Fatalf("failed to unmarshal fleet health report: %v", err)
	}

	if report.TotalDevices != 3 || report.Online != 1 || report.Offline != 2 {
		t.Fatalf("unexpected counts: %+v", report)
	}
	if report.StaleSinceMinutes != 90 || report.CriticalLookbackMinutes != 120 {
		t.Fatalf("unexpected windows: %+v", report)
	}
	if len(report.StaleDevices) != 2 {
		t.Fatalf("expected two stale devices, got %+v", report.StaleDevices)
	}
	if len(report.DevicesWithCriticalEvents) != 1 || report.DevicesWithCriticalEvents[0].ID != "dev-1" {
		t.Fatalf("expected only dev-1 with critical events, got %+v", report.DevicesWithCriticalEvents)
	}
	if len(report.Warnings) != 1 || !strings.Contains(report.Warnings[0], "charlie") {
		t.Fatalf("expected warning for dev-3, got %+v", report.Warnings)
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
