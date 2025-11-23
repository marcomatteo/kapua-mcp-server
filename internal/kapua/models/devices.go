package models

import "time"

// Device Management Models

// DeviceStatus represents the status of a device
type DeviceStatus string

const (
	DeviceStatusEnabled  DeviceStatus = "ENABLED"
	DeviceStatusDisabled DeviceStatus = "DISABLED"
)

// ConnectionStatus represents the connection status filter for devices
type ConnectionStatus string

const (
	ConnectionStatusConnected    ConnectionStatus = "CONNECTED"
	ConnectionStatusDisconnected ConnectionStatus = "DISCONNECTED"
	ConnectionStatusMissing      ConnectionStatus = "MISSING"
	ConnectionStatusNull         ConnectionStatus = "NULL"
)

// Device represents a Kapua device
type Device struct {
	KapuaEntity
	GroupID                     KapuaID                  `json:"groupId,omitempty"`
	ClientID                    string                   `json:"clientId,omitempty"`
	ConnectionID                KapuaID                  `json:"connectionId,omitempty"`
	Connection                  *DeviceConnection        `json:"connection,omitempty"`
	Status                      DeviceStatus             `json:"status,omitempty"`
	DisplayName                 string                   `json:"displayName,omitempty"`
	LastEventID                 KapuaID                  `json:"lastEventId,omitempty"`
	LastEvent                   *DeviceEvent             `json:"lastEvent,omitempty"`
	SerialNumber                string                   `json:"serialNumber,omitempty"`
	ModelID                     string                   `json:"modelId,omitempty"`
	ModelName                   string                   `json:"modelName,omitempty"`
	Imei                        string                   `json:"imei,omitempty"`
	Imsi                        string                   `json:"imsi,omitempty"`
	Iccid                       string                   `json:"iccid,omitempty"`
	BiosVersion                 string                   `json:"biosVersion,omitempty"`
	FirmwareVersion             string                   `json:"firmwareVersion,omitempty"`
	OsVersion                   string                   `json:"osVersion,omitempty"`
	JvmVersion                  string                   `json:"jvmVersion,omitempty"`
	OsgiFrameworkVersion        string                   `json:"osgiFrameworkVersion,omitempty"`
	ApplicationFrameworkVersion string                   `json:"applicationFrameworkVersion,omitempty"`
	ConnectionInterface         string                   `json:"connectionInterface,omitempty"`
	ConnectionIP                string                   `json:"connectionIp,omitempty"`
	ApplicationIdentifiers      string                   `json:"applicationIdentifiers,omitempty"`
	AcceptEncoding              string                   `json:"acceptEncoding,omitempty"`
	CustomAttribute1            string                   `json:"customAttribute1,omitempty"`
	CustomAttribute2            string                   `json:"customAttribute2,omitempty"`
	CustomAttribute3            string                   `json:"customAttribute3,omitempty"`
	CustomAttribute4            string                   `json:"customAttribute4,omitempty"`
	CustomAttribute5            string                   `json:"customAttribute5,omitempty"`
	ExtendedProperties          []DeviceExtendedProperty `json:"extendedProperties,omitempty"`
	TagIDs                      []KapuaID                `json:"tagIds,omitempty"`
}

// DeviceConnection captures the current connection state reported for a device.
type DeviceConnection struct {
	KapuaID    KapuaID          `json:"id,omitempty"`
	Status     ConnectionStatus `json:"status,omitempty"`
	ClientID   string           `json:"clientId,omitempty"`
	CreatedOn  *time.Time       `json:"createdOn,omitempty"`
	ModifiedOn *time.Time       `json:"modifiedOn,omitempty"`
}

// DeviceExtendedProperty represents extended properties of a device
type DeviceExtendedProperty struct {
	GroupName string `json:"groupName,omitempty"`
	Name      string `json:"name,omitempty"`
	Value     string `json:"value,omitempty"`
}

// DeviceListResult represents a list of devices
type DeviceListResult struct {
	Type          string   `json:"type,omitempty"`
	LimitExceeded bool     `json:"limitExceeded,omitempty"`
	Size          int      `json:"size,omitempty"`
	TotalCount    int      `json:"totalCount,omitempty"`
	Items         []Device `json:"items,omitempty"`
}
