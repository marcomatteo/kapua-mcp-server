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
