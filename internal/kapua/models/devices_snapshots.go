package models

// DeviceSnapshot represents the metadata of a device snapshot entry.
type DeviceSnapshot struct {
	ID        string `json:"id,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// DeviceSnapshots captures the response payload returned by the Kapua snapshots API.
type DeviceSnapshots struct {
	Type       string           `json:"type,omitempty"`
	SnapshotID []DeviceSnapshot `json:"snapshotId,omitempty"`
}
