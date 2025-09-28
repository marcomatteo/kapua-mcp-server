package models

import "time"

// DataMessageMetric represents a single metric entry attached to a data message payload.
type DataMessageMetric struct {
	Name      string `json:"name,omitempty"`
	Value     any    `json:"value,omitempty"`
	ValueType string `json:"valueType,omitempty"`
	Unit      string `json:"unit,omitempty"`
}

// DataMessagePayload captures the payload section of a Kapua data message.
type DataMessagePayload struct {
	Metrics []DataMessageMetric `json:"metrics,omitempty"`
	Body    string              `json:"body,omitempty"`
}

// DataMessageChannel identifies the channel information associated with a data message.
type DataMessageChannel struct {
	Type          string   `json:"type,omitempty"`
	SemanticParts []string `json:"semanticParts,omitempty"`
}

// DataMessage mirrors the Kapua datastore message representation.
type DataMessage struct {
	Type        string              `json:"type,omitempty"`
	DatastoreID string              `json:"datastoreId,omitempty"`
	ScopeID     KapuaID             `json:"scopeId,omitempty"`
	Timestamp   time.Time           `json:"timestamp,omitempty"`
	DeviceID    KapuaID             `json:"deviceId,omitempty"`
	ClientID    string              `json:"clientId,omitempty"`
	ReceivedOn  time.Time           `json:"receivedOn,omitempty"`
	SentOn      time.Time           `json:"sentOn,omitempty"`
	CapturedOn  time.Time           `json:"capturedOn,omitempty"`
	Position    *Position           `json:"position,omitempty"`
	Channel     *DataMessageChannel `json:"channel,omitempty"`
	Payload     *DataMessagePayload `json:"payload,omitempty"`
}

// DataMessageListResult represents a paginated list of data messages.
type DataMessageListResult struct {
	Type          string        `json:"type,omitempty"`
	LimitExceeded bool          `json:"limitExceeded,omitempty"`
	Size          int           `json:"size,omitempty"`
	TotalCount    int           `json:"totalCount,omitempty"`
	Items         []DataMessage `json:"items,omitempty"`
}
