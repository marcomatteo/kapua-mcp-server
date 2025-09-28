package models

import "time"

// KapuaID represents a Kapua entity ID
type KapuaID string

func (k KapuaID) String() string {
	return string(k)
}

// KapuaEntity represents the base entity structure in Kapua
type KapuaEntity struct {
	ID         KapuaID   `json:"id,omitempty"`
	ScopeID    KapuaID   `json:"scopeId,omitempty"`
	CreatedOn  time.Time `json:"createdOn,omitempty"`
	CreatedBy  KapuaID   `json:"createdBy,omitempty"`
	ModifiedOn time.Time `json:"modifiedOn,omitempty"`
	ModifiedBy KapuaID   `json:"modifiedBy,omitempty"`
	OptLock    int       `json:"optlock,omitempty"`
}

// KapuaCountResult conveys the total number of entities available for a query.
type KapuaCountResult struct {
	Count int `json:"count,omitempty"`
}

// KapuaQuery captures the common pagination arguments supported by Kapua queries.
type KapuaQuery struct {
	Limit         int  `json:"limit,omitempty"`
	Offset        int  `json:"offset,omitempty"`
	AskTotalCount bool `json:"askTotalCount,omitempty"`
}

// Error Models

// KapuaError represents a standard Kapua error response
type KapuaError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Details string `json:"details,omitempty"`
}

func (e KapuaError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// Position represents geographic and device telemetry positioning data
// as defined by the Kapua position schema.
type Position struct {
	Latitude   float64   `json:"latitude,omitempty"`
	Longitude  float64   `json:"longitude,omitempty"`
	Altitude   float64   `json:"altitude,omitempty"`
	Precision  float64   `json:"precision,omitempty"`
	Heading    float64   `json:"heading,omitempty"`
	Speed      float64   `json:"speed,omitempty"`
	Timestamp  time.Time `json:"timestamp,omitempty"`
	Satellites int       `json:"satellites,omitempty"`
	Status     int       `json:"status,omitempty"`
}
