package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"kapua-mcp-server/internal/kapua/models"
)

const (
	defaultFleetHealthLimit    = 200
	defaultStaleMinutes        = 60
	defaultCriticalMinutes     = 60
	defaultEventConcurrency    = 5
	maxCriticalEventsPerDevice = 5
)

type fleetHealthConfig struct {
	staleMinutes     int
	criticalMinutes  int
	deviceLimit      int
	eventConcurrency int
}

type fleetHealthReport struct {
	GeneratedAt               string           `json:"generated_at"`
	TotalDevices              int              `json:"total_devices"`
	Online                    int              `json:"online"`
	Offline                   int              `json:"offline"`
	Unknown                   int              `json:"unknown"`
	StaleSinceMinutes         int              `json:"stale_since_minutes"`
	StaleDevices              []staleDevice    `json:"stale_devices,omitempty"`
	CriticalLookbackMinutes   int              `json:"critical_lookback_minutes"`
	DevicesWithCriticalEvents []criticalDevice `json:"devices_with_critical_events,omitempty"`
	Warnings                  []string         `json:"warnings,omitempty"`
}

type staleDevice struct {
	ID             string                  `json:"id"`
	ClientID       string                  `json:"clientId,omitempty"`
	Status         models.ConnectionStatus `json:"status,omitempty"`
	LastSeen       string                  `json:"lastSeen,omitempty"`
	LastSeenSource string                  `json:"lastSeenSource,omitempty"`
}

type criticalDevice struct {
	ID       string                  `json:"id"`
	ClientID string                  `json:"clientId,omitempty"`
	Status   models.ConnectionStatus `json:"status,omitempty"`
	Events   []models.DeviceEvent    `json:"events,omitempty"`
}

func (h *KapuaHandler) readFleetHealthResource(ctx context.Context, uri *url.URL) (*mcp.ReadResourceResult, error) {
	cfg := parseFleetHealthConfig(uri)
	h.logger.Info("Building fleet health report (stale>%d min, critical>%d min, limit=%d)", cfg.staleMinutes, cfg.criticalMinutes, cfg.deviceLimit)

	deviceParams := map[string]string{
		"limit":         strconv.Itoa(cfg.deviceLimit),
		"askTotalCount": "true",
	}
	devicesResult, err := h.client.ListDevices(ctx, deviceParams)
	if err != nil {
		return nil, fmt.Errorf("failed to build fleet health: %w", err)
	}

	totalDevices := devicesResult.TotalCount
	if totalDevices == 0 {
		totalDevices = len(devicesResult.Items)
	}

	cutoff := timeNow().Add(-time.Duration(cfg.staleMinutes) * time.Minute)
	criticalSince := timeNow().Add(-time.Duration(cfg.criticalMinutes) * time.Minute)

	online, offline, unknown := 0, 0, 0
	var staleDevices []staleDevice
	var eventTargets []struct {
		device models.Device
		status models.ConnectionStatus
	}

	for _, device := range devicesResult.Items {
		status := connectionStatus(device)
		switch status {
		case models.ConnectionStatusConnected:
			online++
		case models.ConnectionStatusDisconnected, models.ConnectionStatusMissing, models.ConnectionStatusNull:
			offline++
		default:
			unknown++
		}

		if lastSeen, source := lastSeenForDevice(device); !lastSeen.IsZero() && lastSeen.Before(cutoff) {
			staleDevices = append(staleDevices, staleDevice{
				ID:             string(device.ID),
				ClientID:       device.ClientID,
				Status:         status,
				LastSeen:       lastSeen.UTC().Format(time.RFC3339),
				LastSeenSource: source,
			})
		}

		deviceID := string(device.ID)
		if deviceID == "" {
			continue
		}

		eventTargets = append(eventTargets, struct {
			device models.Device
			status models.ConnectionStatus
		}{device: device, status: status})
	}

	var criticalDevices []criticalDevice
	var warnings []string

	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, cfg.eventConcurrency)

	for _, target := range eventTargets {
		target := target
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			deviceID := string(target.device.ID)
			eventsResult, err := h.client.ListDeviceEvents(ctx, deviceID, map[string]string{
				"startDate": criticalSince.Format(time.RFC3339),
				"limit":     "20",
				"sortParam": "receivedOn",
				"sortDir":   "DESCENDING",
			})
			if err != nil {
				mu.Lock()
				warnings = append(warnings, fmt.Sprintf("device %s: %v", labelForDevice(target.device), err))
				mu.Unlock()
				return
			}

			criticalEvents := filterCriticalEvents(eventsResult.Items)
			if len(criticalEvents) == 0 {
				return
			}

			if len(criticalEvents) > maxCriticalEventsPerDevice {
				criticalEvents = criticalEvents[:maxCriticalEventsPerDevice]
			}

			mu.Lock()
			criticalDevices = append(criticalDevices, criticalDevice{
				ID:       deviceID,
				ClientID: target.device.ClientID,
				Status:   target.status,
				Events:   criticalEvents,
			})
			mu.Unlock()
		}()
	}
	wg.Wait()

	report := fleetHealthReport{
		GeneratedAt:               timeNow().UTC().Format(time.RFC3339),
		TotalDevices:              totalDevices,
		Online:                    online,
		Offline:                   offline,
		Unknown:                   unknown,
		StaleSinceMinutes:         cfg.staleMinutes,
		StaleDevices:              staleDevices,
		CriticalLookbackMinutes:   cfg.criticalMinutes,
		DevicesWithCriticalEvents: criticalDevices,
		Warnings:                  warnings,
	}

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fleet health resource: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      "kapua://fleet-health",
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}

func parseFleetHealthConfig(uri *url.URL) fleetHealthConfig {
	cfg := fleetHealthConfig{
		staleMinutes:     defaultStaleMinutes,
		criticalMinutes:  defaultCriticalMinutes,
		deviceLimit:      defaultFleetHealthLimit,
		eventConcurrency: defaultEventConcurrency,
	}
	if uri == nil {
		return cfg
	}

	query := uri.Query()
	cfg.staleMinutes = parsePositiveInt(query.Get("staleMinutes"), cfg.staleMinutes)
	cfg.criticalMinutes = parsePositiveInt(query.Get("criticalMinutes"), cfg.criticalMinutes)
	cfg.deviceLimit = parsePositiveInt(query.Get("limit"), cfg.deviceLimit)
	cfg.eventConcurrency = parsePositiveInt(query.Get("eventConcurrency"), cfg.eventConcurrency)

	return cfg
}

func parsePositiveInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func connectionStatus(device models.Device) models.ConnectionStatus {
	if device.Connection != nil && device.Connection.Status != "" {
		return device.Connection.Status
	}
	return ""
}

func lastSeenForDevice(device models.Device) (time.Time, string) {
	if device.LastEvent != nil {
		if !device.LastEvent.ReceivedOn.IsZero() {
			return device.LastEvent.ReceivedOn, "lastEvent.receivedOn"
		}
		if !device.LastEvent.SentOn.IsZero() {
			return device.LastEvent.SentOn, "lastEvent.sentOn"
		}
	}
	if device.Connection != nil {
		if device.Connection.ModifiedOn != nil && !device.Connection.ModifiedOn.IsZero() {
			return *device.Connection.ModifiedOn, "connection.modifiedOn"
		}
		if device.Connection.CreatedOn != nil && !device.Connection.CreatedOn.IsZero() {
			return *device.Connection.CreatedOn, "connection.createdOn"
		}
	}
	return time.Time{}, ""
}

func labelForDevice(device models.Device) string {
	if device.ClientID != "" {
		return device.ClientID
	}
	return string(device.ID)
}

func filterCriticalEvents(events []models.DeviceEvent) []models.DeviceEvent {
	var critical []models.DeviceEvent
	for _, event := range events {
		if isCriticalEvent(event) {
			critical = append(critical, event)
		}
	}
	return critical
}

func isCriticalEvent(event models.DeviceEvent) bool {
	fields := []string{
		event.Action,
		event.ResponseCode,
		event.EventMessage,
	}

	for _, field := range fields {
		upper := strings.ToUpper(field)
		if upper == "" {
			continue
		}
		if strings.Contains(upper, "CRITICAL") || strings.Contains(upper, "ERROR") || strings.Contains(upper, "FAIL") || strings.Contains(upper, "EXCEPTION") {
			return true
		}
	}
	return false
}
