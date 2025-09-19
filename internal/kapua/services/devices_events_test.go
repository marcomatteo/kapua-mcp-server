package services

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type eventRoundTripFunc func(*http.Request) (*http.Response, error)

func (f eventRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestListDeviceEventsSuccess(t *testing.T) {
	client := newTestKapuaClient()

	params := map[string]string{
		"resource":  "LOG",
		"startDate": "2023-03-10T12:00:00Z",
	}

	sampleResp := `{"type":"deviceEventListResult","limitExceeded":false,"size":1,"items":[{"id":"event-1","scopeId":"tenant","deviceId":"device-123","sentOn":"2023-03-10T12:00:01Z","receivedOn":"2023-03-10T12:00:02Z","resource":"LOG","action":"CREATE","responseCode":"ACCEPTED","eventMessage":"Device started","position":{"latitude":45.0,"longitude":9.0,"timestamp":"2023-03-10T12:00:01Z"}}]}`

	client.httpClient = &http.Client{Transport: eventRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/devices/device-123/events" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		values, err := url.ParseQuery(req.URL.RawQuery)
		if err != nil {
			t.Fatalf("failed to parse query: %v", err)
		}
		if got := values.Get("resource"); got != "LOG" {
			t.Fatalf("expected resource filter LOG, got %q", got)
		}
		if got := values.Get("startDate"); got != "2023-03-10T12:00:00Z" {
			t.Fatalf("expected startDate filter, got %q", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(sampleResp)),
			Header:     make(http.Header),
		}, nil
	})}

	result, err := client.ListDeviceEvents(context.Background(), "device-123", params)
	if err != nil {
		t.Fatalf("ListDeviceEvents returned error: %v", err)
	}
	if result == nil || len(result.Items) != 1 {
		t.Fatalf("expected one device event, got %+v", result)
	}
	event := result.Items[0]
	if event.EventMessage != "Device started" {
		t.Fatalf("unexpected event message: %+v", event)
	}
	if event.Position == nil || event.Position.Latitude != 45.0 {
		t.Fatalf("unexpected position data: %+v", event.Position)
	}
}

func TestListDeviceEventsRequestError(t *testing.T) {
	client := newTestKapuaClient()

	client.httpClient = &http.Client{Transport: eventRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network error")
	})}

	_, err := client.ListDeviceEvents(context.Background(), "device-123", nil)
	if err == nil || !strings.Contains(err.Error(), "list device events request failed") {
		t.Fatalf("expected wrapped request error, got %v", err)
	}
}

func TestListDeviceEventsHandleError(t *testing.T) {
	client := newTestKapuaClient()

	client.httpClient = &http.Client{Transport: eventRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(`{"code":"500","message":"kapua error"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	_, err := client.ListDeviceEvents(context.Background(), "device-123", nil)
	if err == nil || !strings.Contains(err.Error(), "failed to list device events") {
		t.Fatalf("expected response error, got %v", err)
	}
}
