package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestKapuaIDString(t *testing.T) {
	id := KapuaID("device-123")
	if id.String() != "device-123" {
		t.Fatalf("expected KapuaID string to match, got %q", id.String())
	}
}

func TestKapuaErrorErrorMessage(t *testing.T) {
	err := KapuaError{Message: "failure"}
	if err.Error() != "failure" {
		t.Fatalf("expected error message 'failure', got %q", err.Error())
	}

	err = KapuaError{Message: "failure", Details: "extra"}
	if err.Error() != "failure: extra" {
		t.Fatalf("expected combined message, got %q", err.Error())
	}
}

func TestAccessTokenJSONMarshalling(t *testing.T) {
	payload := `{"tokenId":"new-token","refreshToken":"refresh-token","expiresOn":"2025-01-02T15:04:05Z","refreshExpiresOn":"2025-01-03T15:04:05Z","scopeId":"tenant","userId":"user-1"}`

	var token AccessToken
	if err := json.Unmarshal([]byte(payload), &token); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if token.TokenID != "new-token" {
		t.Fatalf("expected tokenId to be set, got %q", token.TokenID)
	}
	if token.RefreshToken != "refresh-token" {
		t.Fatalf("expected refreshToken to be set, got %q", token.RefreshToken)
	}
	if string(token.ScopeID) != "tenant" {
		t.Fatalf("expected scopeId tenant, got %q", token.ScopeID)
	}

	expires, _ := time.Parse(time.RFC3339, "2025-01-02T15:04:05Z")
	refreshExpires, _ := time.Parse(time.RFC3339, "2025-01-03T15:04:05Z")
	if !token.ExpiresOn.Equal(expires) {
		t.Fatalf("expected expiry %v, got %v", expires, token.ExpiresOn)
	}
	if !token.RefreshExpiresOn.Equal(refreshExpires) {
		t.Fatalf("expected refresh expiry %v, got %v", refreshExpires, token.RefreshExpiresOn)
	}
}

func TestDeviceConfigurationJSON(t *testing.T) {
	payload := `{"configuration":[{"id":"component-1","definition":{"id":"org.eclipse.kura.sample","name":"Sample Component","description":"A test","AD":[{"id":"poll.interval","name":"Polling Interval","type":"INTEGER","default":"60"}]},"properties":{"property":[{"name":"poll.interval","array":false,"type":"INTEGER","value":["60"]}]}}]}`

	var cfg DeviceConfiguration
	if err := json.Unmarshal([]byte(payload), &cfg); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if len(cfg.Configuration) != 1 {
		t.Fatalf("expected one configuration entry, got %d", len(cfg.Configuration))
	}

	comp := cfg.Configuration[0]
	if comp.ID != "component-1" {
		t.Fatalf("expected component id component-1, got %q", comp.ID)
	}
	if comp.Definition == nil || comp.Definition.Name != "Sample Component" {
		t.Fatalf("expected definition name Sample Component, got %+v", comp.Definition)
	}
	if comp.Properties == nil || len(comp.Properties.Property) != 1 {
		t.Fatalf("expected one property definition, got %+v", comp.Properties)
	}
	if comp.Properties.Property[0].Value[0] != "60" {
		t.Fatalf("expected property value 60, got %+v", comp.Properties.Property[0].Value)
	}
}

func TestDeviceJSON(t *testing.T) {
	payload := `{"id":"device-123","scopeId":"tenant","clientId":"client-1","status":"ENABLED","displayName":"Temperature Sensor","tagIds":["tag-1","tag-2"],"extendedProperties":[{"groupName":"system","name":"fw","value":"1.0"}]}`

	var device Device
	if err := json.Unmarshal([]byte(payload), &device); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if device.ID.String() != "device-123" {
		t.Fatalf("expected id device-123, got %s", device.ID.String())
	}
	if device.Status != DeviceStatusEnabled {
		t.Fatalf("expected status ENABLED, got %s", device.Status)
	}
	if device.DisplayName != "Temperature Sensor" {
		t.Fatalf("expected display name Temperature Sensor, got %q", device.DisplayName)
	}
	if len(device.TagIDs) != 2 || device.TagIDs[0] != "tag-1" {
		t.Fatalf("expected tag IDs to be unmarshalled, got %+v", device.TagIDs)
	}
	if len(device.ExtendedProperties) != 1 || device.ExtendedProperties[0].Value != "1.0" {
		t.Fatalf("expected extended property value 1.0, got %+v", device.ExtendedProperties)
	}
}
