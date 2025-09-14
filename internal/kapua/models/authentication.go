package models

import "time"

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
	UserID             KapuaID          `json:"userId,omitempty"`
	CredentialType     CredentialType   `json:"credentialType,omitempty"`
	CredentialKey      string           `json:"credentialKey,omitempty"`
	Status             CredentialStatus `json:"status,omitempty"`
	ExpirationDate     time.Time        `json:"expirationDate,omitempty"`
	LoginFailures      int              `json:"loginFailures,omitempty"`
	FirstLoginFailure  time.Time        `json:"firstLoginFailure,omitempty"`
	LoginFailuresReset time.Time        `json:"loginFailuresReset,omitempty"`
	LockoutReset       time.Time        `json:"lockoutReset,omitempty"`
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
