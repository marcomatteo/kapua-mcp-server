package models

import "time"

// LogProperty describes a single property entry within a device log.
type LogProperty struct {
	Name      string `json:"name,omitempty"`
	Value     any    `json:"value,omitempty"`
	ValueType string `json:"valueType,omitempty"`
	Unit      string `json:"unit,omitempty"`
}

// LogProperties wraps the list of log properties returned by Kapua.
type LogProperties struct {
	LogProperty []LogProperty `json:"logProperty,omitempty"`
}

// DeviceLog represents a Kapua device log entry.
type DeviceLog struct {
	Type          string              `json:"type,omitempty"`
	ScopeID       KapuaID             `json:"scopeId,omitempty"`
	Channel       *DataMessageChannel `json:"channel,omitempty"`
	ClientID      string              `json:"clientId,omitempty"`
	DeviceID      KapuaID             `json:"deviceId,omitempty"`
	LogProperties *LogProperties      `json:"logProperties,omitempty"`
	ReceivedOn    time.Time           `json:"receivedOn,omitempty"`
	StoreID       string              `json:"storeId,omitempty"`
	Timestamp     time.Time           `json:"timestamp,omitempty"`
}

// DeviceLogListResult represents a paginated list of device logs.
type DeviceLogListResult struct {
	Type          string      `json:"type,omitempty"`
	LimitExceeded bool        `json:"limitExceeded,omitempty"`
	Size          int         `json:"size,omitempty"`
	TotalCount    int         `json:"totalCount,omitempty"`
	Items         []DeviceLog `json:"items,omitempty"`
}
