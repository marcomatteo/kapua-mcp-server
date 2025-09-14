package models

// Device Management Models

// DeviceStatus represents the status of a device
type DeviceStatus string

const (
	DeviceStatusEnabled  DeviceStatus = "ENABLED"
	DeviceStatusDisabled DeviceStatus = "DISABLED"
)

// Device represents a Kapua device
type Device struct {
	KapuaEntity
	GroupID                     KapuaID                  `json:"groupId,omitempty"`
	ClientID                    string                   `json:"clientId,omitempty"`
	ConnectionID                KapuaID                  `json:"connectionId,omitempty"`
	Status                      DeviceStatus             `json:"status,omitempty"`
	DisplayName                 string                   `json:"displayName,omitempty"`
	LastEventID                 KapuaID                  `json:"lastEventId,omitempty"`
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
