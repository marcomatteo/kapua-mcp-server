package models

import "time"

// KapuaID represents a Kapua entity ID
type KapuaID string

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

// Authentication Request/Response Models

// UsernamePasswordCredentials represents the request body for user/password authentication
type UsernamePasswordCredentials struct {
	Username           string `json:"username" jsonschema:"required,description=The username of the user,pattern=^[a-zA-Z0-9\\_\\-]{3,}$"`
	Password           string `json:"password" jsonschema:"required,description=The password of the user"`
	AuthenticationCode string `json:"authenticationCode,omitempty" jsonschema:"description=The MFA authentication code"`
	TrustKey           string `json:"trustKey,omitempty" jsonschema:"description=A long-lived key for trusted machine authentication"`
	TrustMe            bool   `json:"trustMe,omitempty" jsonschema:"description=Whether to generate a TrustKey or not"`
}

// APIKeyCredentials represents the request body for API key authentication
type APIKeyCredentials struct {
	APIKey string `json:"apiKey" jsonschema:"required,description=The API key for authentication,format=base64"`
}

// JWTCredentials represents the request body for JWT authentication
type JWTCredentials struct {
	JWT string `json:"jwt" jsonschema:"required,description=The JWT token for authentication"`
}

// RefreshTokenRequest represents the request body for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" jsonschema:"required,description=The refresh token,format=uuid"`
	TokenID      string `json:"tokenId" jsonschema:"required,description=The current access token ID"`
}

// AccessToken represents the authentication token response
type AccessToken struct {
	KapuaEntity
	UserID           KapuaID   `json:"userId,omitempty"`
	TokenID          string    `json:"tokenId,omitempty"`
	ExpiresOn        time.Time `json:"expiresOn,omitempty"`
	InvalidatedOn    time.Time `json:"invalidatedOn,omitempty"`
	RefreshToken     string    `json:"refreshToken,omitempty"`
	RefreshExpiresOn time.Time `json:"refreshExpiresOn,omitempty"`
	TrustKey         string    `json:"trustKey,omitempty"`
}

// AccessPermission represents a permission assigned to an access info
type AccessPermission struct {
	KapuaEntity
	AccessInfoID KapuaID    `json:"accessInfoId,omitempty"`
	Permission   Permission `json:"permission,omitempty"`
}

// RolePermission represents a permission assigned through a role
type RolePermission struct {
	KapuaEntity
	RoleID     KapuaID    `json:"roleId,omitempty"`
	Permission Permission `json:"permission,omitempty"`
}

// Permission represents a single permission
type Permission struct {
	Domain        string  `json:"domain,omitempty"`
	Action        string  `json:"action,omitempty"`
	TargetScopeID KapuaID `json:"targetScopeId,omitempty"`
	Forwardable   bool    `json:"forwardable,omitempty"`
}

// LoginInfo represents comprehensive authentication and authorization information
type LoginInfo struct {
	AccessPermissions []AccessPermission `json:"accessPermission,omitempty"`
	AccessToken       AccessToken        `json:"accessToken,omitempty"`
	RolePermissions   []RolePermission   `json:"rolePermissions,omitempty"`
}

// Credential represents user credential information
type Credential struct {
	KapuaEntity
	UserID             KapuaID           `json:"userId,omitempty"`
	CredentialType     CredentialType    `json:"credentialType,omitempty"`
	CredentialKey      string            `json:"credentialKey,omitempty"`
	Status             CredentialStatus  `json:"status,omitempty"`
	ExpirationDate     time.Time         `json:"expirationDate,omitempty"`
	LoginFailures      int               `json:"loginFailures,omitempty"`
	FirstLoginFailure  time.Time         `json:"firstLoginFailure,omitempty"`
	LoginFailuresReset time.Time         `json:"loginFailuresReset,omitempty"`
	LockoutReset       time.Time         `json:"lockoutReset,omitempty"`
}

// CredentialType represents the type of credential
type CredentialType string

const (
	CredentialTypePassword CredentialType = "PASSWORD"
	CredentialTypeAPIKey   CredentialType = "API_KEY"
	CredentialTypeJWT      CredentialType = "JWT"
)

// CredentialStatus represents the status of a credential
type CredentialStatus string

const (
	CredentialStatusEnabled  CredentialStatus = "ENABLED"
	CredentialStatusDisabled CredentialStatus = "DISABLED"
)

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
	GroupID                        KapuaID      `json:"groupId,omitempty"`
	ClientID                       string       `json:"clientId,omitempty"`
	ConnectionID                   KapuaID      `json:"connectionId,omitempty"`
	Status                         DeviceStatus `json:"status,omitempty"`
	DisplayName                    string       `json:"displayName,omitempty"`
	LastEventID                    KapuaID      `json:"lastEventId,omitempty"`
	SerialNumber                   string       `json:"serialNumber,omitempty"`
	ModelID                        string       `json:"modelId,omitempty"`
	ModelName                      string       `json:"modelName,omitempty"`
	Imei                           string       `json:"imei,omitempty"`
	Imsi                           string       `json:"imsi,omitempty"`
	Iccid                          string       `json:"iccid,omitempty"`
	BiosVersion                    string       `json:"biosVersion,omitempty"`
	FirmwareVersion                string       `json:"firmwareVersion,omitempty"`
	OsVersion                      string       `json:"osVersion,omitempty"`
	JvmVersion                     string       `json:"jvmVersion,omitempty"`
	OsgiFrameworkVersion           string       `json:"osgiFrameworkVersion,omitempty"`
	ApplicationFrameworkVersion    string       `json:"applicationFrameworkVersion,omitempty"`
	ConnectionInterface            string       `json:"connectionInterface,omitempty"`
	ConnectionIP                   string       `json:"connectionIp,omitempty"`
	ApplicationIdentifiers         string       `json:"applicationIdentifiers,omitempty"`
	AcceptEncoding                 string       `json:"acceptEncoding,omitempty"`
	CustomAttribute1               string       `json:"customAttribute1,omitempty"`
	CustomAttribute2               string       `json:"customAttribute2,omitempty"`
	CustomAttribute3               string       `json:"customAttribute3,omitempty"`
	CustomAttribute4               string       `json:"customAttribute4,omitempty"`
	CustomAttribute5               string       `json:"customAttribute5,omitempty"`
	ExtendedProperties             []DeviceExtendedProperty `json:"extendedProperties,omitempty"`
	TagIDs                         []KapuaID    `json:"tagIds,omitempty"`
}

// DeviceExtendedProperty represents extended properties of a device
type DeviceExtendedProperty struct {
	GroupName string `json:"groupName,omitempty"`
	Name      string `json:"name,omitempty"`
	Value     string `json:"value,omitempty"`
}

// DeviceCreator represents a device creation request
type DeviceCreator struct {
	ScopeID                        KapuaID                  `json:"scopeId,omitempty"`
	GroupID                        KapuaID                  `json:"groupId,omitempty"`
	ClientID                       string                   `json:"clientId" jsonschema:"required,description=The Kura Client ID of this device"`
	Status                         DeviceStatus             `json:"status,omitempty"`
	DisplayName                    string                   `json:"displayName,omitempty"`
	SerialNumber                   string                   `json:"serialNumber,omitempty"`
	ModelID                        string                   `json:"modelId,omitempty"`
	ModelName                      string                   `json:"modelName,omitempty"`
	Imei                           string                   `json:"imei,omitempty"`
	Imsi                           string                   `json:"imsi,omitempty"`
	Iccid                          string                   `json:"iccid,omitempty"`
	BiosVersion                    string                   `json:"biosVersion,omitempty"`
	FirmwareVersion                string                   `json:"firmwareVersion,omitempty"`
	OsVersion                      string                   `json:"osVersion,omitempty"`
	JvmVersion                     string                   `json:"jvmVersion,omitempty"`
	OsgiFrameworkVersion           string                   `json:"osgiFrameworkVersion,omitempty"`
	ApplicationFrameworkVersion    string                   `json:"applicationFrameworkVersion,omitempty"`
	ConnectionInterface            string                   `json:"connectionInterface,omitempty"`
	ConnectionIP                   string                   `json:"connectionIp,omitempty"`
	ApplicationIdentifiers         string                   `json:"applicationIdentifiers,omitempty"`
	AcceptEncoding                 string                   `json:"acceptEncoding,omitempty"`
	CustomAttribute1               string                   `json:"customAttribute1,omitempty"`
	CustomAttribute2               string                   `json:"customAttribute2,omitempty"`
	CustomAttribute3               string                   `json:"customAttribute3,omitempty"`
	CustomAttribute4               string                   `json:"customAttribute4,omitempty"`
	CustomAttribute5               string                   `json:"customAttribute5,omitempty"`
	ExtendedProperties             []DeviceExtendedProperty `json:"extendedProperties,omitempty"`
	TagIDs                         []KapuaID                `json:"tagIds,omitempty"`
}

// DeviceListResult represents a list of devices
type DeviceListResult struct {
	Type          string   `json:"type,omitempty"`
	LimitExceeded bool     `json:"limitExceeded,omitempty"`
	Size          int      `json:"size,omitempty"`
	TotalCount    int      `json:"totalCount,omitempty"`
	Items         []Device `json:"items,omitempty"`
}
