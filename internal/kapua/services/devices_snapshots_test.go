package services

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type snapshotRoundTripFunc func(*http.Request) (*http.Response, error)

func (f snapshotRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestListDeviceSnapshotsSuccess(t *testing.T) {
	client := newTestKapuaClient()
	response := `{"type":"deviceSnapshots","snapshotId":[{"id":"snap-1","timestamp":1}]}`

	client.httpClient = &http.Client{Transport: snapshotRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/devices/device-1/snapshots" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(response)),
			Header:     make(http.Header),
		}, nil
	})}

	snaps, err := client.ListDeviceSnapshots(context.Background(), "device-1")
	if err != nil {
		t.Fatalf("ListDeviceSnapshots returned error: %v", err)
	}
	if snaps == nil || len(snaps.SnapshotID) != 1 || snaps.SnapshotID[0].ID != "snap-1" {
		t.Fatalf("unexpected snapshots payload: %+v", snaps)
	}
}

func TestListDeviceSnapshotsRequestError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: snapshotRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("timeout")
	})}

	_, err := client.ListDeviceSnapshots(context.Background(), "device-1")
	if err == nil || !strings.Contains(err.Error(), "list device snapshots request failed") {
		t.Fatalf("expected wrapped request error, got %v", err)
	}
}

func TestListDeviceSnapshotsHandleError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: snapshotRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("kapua error")),
			Header:     make(http.Header),
		}, nil
	})}

	_, err := client.ListDeviceSnapshots(context.Background(), "device-1")
	if err == nil || !strings.Contains(err.Error(), "failed to list device snapshots") {
		t.Fatalf("expected response error, got %v", err)
	}
}

func TestReadDeviceSnapshotConfigurationsSuccess(t *testing.T) {
	client := newTestKapuaClient()
	response := `{"configuration":[{"id":"component-1"}]}`

	client.httpClient = &http.Client{Transport: snapshotRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/devices/device-1/snapshots/snap-1" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(response)),
			Header:     make(http.Header),
		}, nil
	})}

	conf, err := client.ReadDeviceSnapshotConfigurations(context.Background(), "device-1", "snap-1")
	if err != nil {
		t.Fatalf("ReadDeviceSnapshotConfigurations returned error: %v", err)
	}
	if conf == nil || len(conf.Configuration) != 1 || conf.Configuration[0].ID != "component-1" {
		t.Fatalf("unexpected configuration payload: %+v", conf)
	}
}

func TestReadDeviceSnapshotConfigurationsRequestError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: snapshotRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("timeout")
	})}

	_, err := client.ReadDeviceSnapshotConfigurations(context.Background(), "device-1", "snap-1")
	if err == nil || !strings.Contains(err.Error(), "read device snapshot configurations request failed") {
		t.Fatalf("expected wrapped request error, got %v", err)
	}
}

func TestReadDeviceSnapshotConfigurationsHandleError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: snapshotRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("kapua error")),
			Header:     make(http.Header),
		}, nil
	})}

	_, err := client.ReadDeviceSnapshotConfigurations(context.Background(), "device-1", "snap-1")
	if err == nil || !strings.Contains(err.Error(), "failed to read device snapshot configurations") {
		t.Fatalf("expected response error, got %v", err)
	}
}

func TestRollbackDeviceSnapshotSuccess(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: snapshotRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/devices/device-1/snapshots/snap-1/_rollback" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})}

	if err := client.RollbackDeviceSnapshot(context.Background(), "device-1", "snap-1"); err != nil {
		t.Fatalf("RollbackDeviceSnapshot returned error: %v", err)
	}
}

func TestRollbackDeviceSnapshotRequestError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: snapshotRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("timeout")
	})}

	err := client.RollbackDeviceSnapshot(context.Background(), "device-1", "snap-1")
	if err == nil || !strings.Contains(err.Error(), "rollback device snapshot request failed") {
		t.Fatalf("expected wrapped request error, got %v", err)
	}
}

func TestRollbackDeviceSnapshotHandleError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: snapshotRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("kapua error")),
			Header:     make(http.Header),
		}, nil
	})}

	err := client.RollbackDeviceSnapshot(context.Background(), "device-1", "snap-1")
	if err == nil || !strings.Contains(err.Error(), "failed to rollback device snapshot") {
		t.Fatalf("expected response error, got %v", err)
	}
}
