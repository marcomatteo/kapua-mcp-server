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

type inventoryRoundTripFunc func(*http.Request) (*http.Response, error)

func (f inventoryRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestReadDeviceInventorySuccess(t *testing.T) {
	client := newTestKapuaClient()
	sampleResp := `{"inventoryItems":[{"name":"pkg","version":"1.0","itemType":"DEB"}]}`
	client.httpClient = &http.Client{Transport: inventoryRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/devices/device-123/inventory" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(sampleResp)),
			Header:     make(http.Header),
		}, nil
	})}

	inv, err := client.ReadDeviceInventory(context.Background(), "device-123")
	if err != nil {
		t.Fatalf("ReadDeviceInventory returned error: %v", err)
	}
	if inv == nil || len(inv.InventoryItems) != 1 || inv.InventoryItems[0].Name != "pkg" {
		t.Fatalf("unexpected inventory payload: %+v", inv)
	}
}

func TestReadDeviceInventoryRequestError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: inventoryRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})}

	_, err := client.ReadDeviceInventory(context.Background(), "device-123")
	if err == nil || !strings.Contains(err.Error(), "read device inventory request failed") {
		t.Fatalf("expected wrapped request error, got %v", err)
	}
}

func TestListDeviceInventoryBundlesSuccess(t *testing.T) {
	client := newTestKapuaClient()
	sampleResp := `{"inventoryBundles":[{"id":"0","name":"org.eclipse.kura","version":"1.0","status":"ACTIVE"}]}`
	client.httpClient = &http.Client{Transport: inventoryRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/devices/device-123/inventory/bundles" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(sampleResp)),
			Header:     make(http.Header),
		}, nil
	})}

	inv, err := client.ListDeviceInventoryBundles(context.Background(), "device-123")
	if err != nil {
		t.Fatalf("ListDeviceInventoryBundles returned error: %v", err)
	}
	if inv == nil || len(inv.InventoryBundles) != 1 || inv.InventoryBundles[0].ID != "0" {
		t.Fatalf("unexpected bundles payload: %+v", inv)
	}
}

func TestStartDeviceInventoryBundle(t *testing.T) {
	client := newTestKapuaClient()
	bundle := models.DeviceInventoryBundle{ID: "0", Name: "org.eclipse.kura"}
	client.httpClient = &http.Client{Transport: inventoryRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/devices/device-123/inventory/bundles/_start" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		var body models.DeviceInventoryBundle
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.ID != "0" {
			t.Fatalf("unexpected bundle body: %+v", body)
		}
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})}

	if err := client.StartDeviceInventoryBundle(context.Background(), "device-123", bundle); err != nil {
		t.Fatalf("StartDeviceInventoryBundle returned error: %v", err)
	}
}

func TestListDeviceInventoryDeploymentPackagesSuccess(t *testing.T) {
	client := newTestKapuaClient()
	sampleResp := `{"deploymentPackages":[{"name":"org.eclipse.kura.example","version":"1.0.0","packageBundles":[{"id":"1","name":"bundle","version":"1.0","status":"ACTIVE"}]}]}`
	client.httpClient = &http.Client{Transport: inventoryRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/devices/device-123/inventory/packages" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(sampleResp)),
			Header:     make(http.Header),
		}, nil
	})}

	inv, err := client.ListDeviceInventoryDeploymentPackages(context.Background(), "device-123")
	if err != nil {
		t.Fatalf("ListDeviceInventoryDeploymentPackages returned error: %v", err)
	}
	if inv == nil || len(inv.DeploymentPackages) != 1 || inv.DeploymentPackages[0].Name != "org.eclipse.kura.example" {
		t.Fatalf("unexpected deployment packages payload: %+v", inv)
	}
}
