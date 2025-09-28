package models

import "time"

// ClientInfo mirrors Kapua's clientInfo payload representing a messaging client.
type ClientInfo struct {
	KapuaEntity
	Type           string    `json:"type,omitempty"`
	ClientID       string    `json:"clientId,omitempty"`
	FirstMessageID string    `json:"firstMessageId,omitempty"`
	FirstMessageOn time.Time `json:"firstMessageOn,omitempty"`
	LastMessageID  string    `json:"lastMessageId,omitempty"`
	LastMessageOn  time.Time `json:"lastMessageOn,omitempty"`
}

// ClientInfoListResult represents a paginated list of clientInfo objects.
type ClientInfoListResult struct {
	Type          string       `json:"type,omitempty"`
	LimitExceeded bool         `json:"limitExceeded,omitempty"`
	Size          int          `json:"size,omitempty"`
	TotalCount    int          `json:"totalCount,omitempty"`
	Items         []ClientInfo `json:"items,omitempty"`
}
