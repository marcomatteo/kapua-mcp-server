package models

// Device Configuration Models (per specs: componentConfigurations)

// DeviceConfiguration represents the response body of
// GET /{scopeId}/devices/{deviceId}/configurations
type DeviceConfiguration struct {
	Configuration []ComponentConfiguration `json:"configuration,omitempty"`
}

// ComponentConfiguration describes a single component configuration
type ComponentConfiguration struct {
	ID         string               `json:"id,omitempty"`
	Definition *ComponentDefinition `json:"definition,omitempty"`
	Properties *ComponentProperties `json:"properties,omitempty"`
}

// ComponentDefinition describes the meta-type for a component
type ComponentDefinition struct {
	ID          string                `json:"id,omitempty"`
	Name        string                `json:"name,omitempty"`
	Description string                `json:"description,omitempty"`
	AD          []AttributeDefinition `json:"AD,omitempty"`
	Icon        []Icon                `json:"Icon,omitempty"`
}

// AttributeDefinition mirrors the attributeDefinition schema
type AttributeDefinition struct {
	Option      []Option `json:"Option,omitempty"`
	Default     string   `json:"default,omitempty"`
	Type        string   `json:"type,omitempty"`
	Cardinality int      `json:"cardinality,omitempty"`
	Min         string   `json:"min,omitempty"`
	Max         string   `json:"max,omitempty"`
	Description string   `json:"description,omitempty"`
	ID          string   `json:"id,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Name        string   `json:"name,omitempty"`
}

// Option for attribute definitions
type Option struct {
	Label string `json:"label,omitempty"`
	Value string `json:"value,omitempty"`
}

// Icon resource reference
type Icon struct {
	Resource string `json:"resource,omitempty"`
	Size     int    `json:"size,omitempty"`
}

// ComponentProperties holds a list of property definitions
type ComponentProperties struct {
	Property []PropertyDefinition `json:"property,omitempty"`
}

// PropertyDefinition mirrors the propertyDefinition schema
type PropertyDefinition struct {
	Name      string   `json:"name,omitempty"`
	Array     bool     `json:"array,omitempty"`
	Encrypted bool     `json:"encrypted,omitempty"`
	Type      string   `json:"type,omitempty"`
	Value     []string `json:"value,omitempty"`
}
