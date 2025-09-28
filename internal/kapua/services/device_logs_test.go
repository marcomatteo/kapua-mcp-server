package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"kapua-mcp-server/internal/kapua/models"
)

type deviceLogsRoundTripFunc func(*http.Request) (*http.Response, error)

func (f deviceLogsRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestListDeviceLogsSuccess(t *testing.T) {
	client := newTestKapuaClient()

	strict := true
	limit := 50
	offset := 0
	query := &DeviceLogsQuery{
		ClientID:        "client-1",
		Channel:         "/foo/bar",
		StrictChannel:   &strict,
		StartDate:       "2023-01-01T00:00:00Z",
		EndDate:         "2023-01-02T00:00:00Z",
		LogPropertyName: "MESSAGE",
		LogPropertyType: "string",
		LogPropertyMin:  "A",
		LogPropertyMax:  "Z",
		SortDir:         "DESCENDING",
		Limit:           &limit,
		Offset:          &offset,
	}

	sampleResp := models.DeviceLogListResult{
		Items: []models.DeviceLog{{StoreID: "store-1", ClientID: "client-1"}},
		Size:  1,
	}
	body, err := json.Marshal(sampleResp)
	if err != nil {
		t.Fatalf("failed to marshal sample response: %v", err)
	}

	client.httpClient = &http.Client{Transport: deviceLogsRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/deviceLogs" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}

		queryVals := req.URL.Query()
		expectations := map[string]string{
			"clientId":        "client-1",
			"channel":         "/foo/bar",
			"strictChannel":   "true",
			"startDate":       "2023-01-01T00:00:00Z",
			"endDate":         "2023-01-02T00:00:00Z",
			"logPropertyName": "MESSAGE",
			"logPropertyType": "string",
			"logPropertyMin":  "A",
			"logPropertyMax":  "Z",
			"sortDir":         "DESCENDING",
			"limit":           "50",
			"offset":          "0",
		}
		for key, expected := range expectations {
			if got := queryVals.Get(key); got != expected {
				t.Fatalf("expected %s=%s, got %s", key, expected, got)
			}
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(string(body))),
			Header:     make(http.Header),
		}, nil
	})}

	result, err := client.ListDeviceLogs(context.Background(), query)
	if err != nil {
		t.Fatalf("ListDeviceLogs returned error: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].StoreID != "store-1" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestListDeviceLogsNilQuery(t *testing.T) {
	client := newTestKapuaClient()

	client.httpClient = &http.Client{Transport: deviceLogsRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.RawQuery != "" {
			t.Fatalf("expected empty query string, got %q", req.URL.RawQuery)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"items":[]}`)),
			Header:     make(http.Header),
		}, nil
	})}

	result, err := client.ListDeviceLogs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListDeviceLogs returned error: %v", err)
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected empty items, got %+v", result)
	}
}

func TestListDeviceLogsRequestError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: deviceLogsRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})}

	_, err := client.ListDeviceLogs(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "list device logs request failed") {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestListDeviceLogsHandleError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: deviceLogsRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(`{"code":"ERR"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	_, err := client.ListDeviceLogs(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "failed to list device logs") {
		t.Fatalf("expected response error, got %v", err)
	}
}

func TestDeviceLogsQueryToValues(t *testing.T) {
	strict := false
	limit := 100
	query := &DeviceLogsQuery{
		ClientID:        "client",
		Channel:         "channel",
		StrictChannel:   &strict,
		StartDate:       "start",
		EndDate:         "end",
		LogPropertyName: "name",
		LogPropertyType: "type",
		LogPropertyMin:  "min",
		LogPropertyMax:  "max",
		SortDir:         "ASCENDING",
		Limit:           &limit,
	}

	values := query.toValues()
	checks := map[string]string{
		"clientId":        "client",
		"channel":         "channel",
		"strictChannel":   "false",
		"startDate":       "start",
		"endDate":         "end",
		"logPropertyName": "name",
		"logPropertyType": "type",
		"logPropertyMin":  "min",
		"logPropertyMax":  "max",
		"sortDir":         "ASCENDING",
		"limit":           "100",
	}
	for key, expected := range checks {
		if got := values.Get(key); got != expected {
			t.Fatalf("expected %s=%s, got %s", key, expected, got)
		}
	}
}

func TestDeviceLogsQueryNil(t *testing.T) {
	var query *DeviceLogsQuery
	if values := query.toValues(); len(values) != 0 {
		t.Fatalf("expected empty values, got %v", values)
	}
}
