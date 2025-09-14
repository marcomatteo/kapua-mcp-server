# kapua-mcp-server
MCP Server for Eclipse Kapua IoT Device Management.

## Project Structure

```
kapua-mcp-server/
├── cmd/
│   └── server/
│       ├── logging_middleware.go   # HTTP request logging wrapper
│       └── main.go                 # Application entry point (MCP HTTP server)
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration loading from env/.venv
│   └── kapua/
│       ├── handlers/
│       │   ├── kapua.go            # Common MCP resource plumbing
│       │   ├── devices.go          # Device tools (list/update/delete)
│       │   └── devices_configurations.go # Configuration tools
│       ├── models/
│       │   ├── devices.go          # Device structs, list results
│       │   ├── devices_configuration.go  # Component configuration schemas
│       │   └── authentication.go   # Auth models
│       └── services/
│           ├── kapua_client.go     # HTTP client + auth/refresh
│           ├── authentication.go   # /authentication endpoints
│           ├── devices.go          # /{scopeId}/devices endpoints
│           └── devices_configurations.go # /devices/{deviceId}/configurations
├── pkg/
│   └── utils/
│       └── logger.go               # Structured logging helper
├── specs/
│   └── kapua_openapi.yaml          # Kapua REST API (OpenAPI)
├── bin/
│   └── kapua-mcp-server            # Built binary output
├── Makefile
├── go.mod
└── go.sum
```

## Requirements
- Go 1.23+

## Configuration
The server reads configuration from environment variables and a simple `.venv` file (if present). The `.venv` file uses `KEY=VALUE` lines.

Required settings:
- `KAPUA_API_ENDPOINT`: Kapua REST base URL (e.g., `https://kapua.example.com/api`)
- `KAPUA_USER`: Kapua username
- `KAPUA_PASSWORD`: Kapua password

Example `.venv`:
```
KAPUA_API_ENDPOINT=https://kapua.example.com/api
KAPUA_USER=my-user
KAPUA_PASSWORD=We!come12345
```

## Build and Run

Using Makefile:
- Build: `make build`
- Run: `make run` (runs `./bin/kapua-mcp-server`)
- Build and run: `make`

Server listens on `host:port` (defaults: `localhost:8000`).

## MCP Tools and Resources
- Tool: `kapua-list-devices`
  - Lists devices within the authenticated Kapua scope (derived from the access token)
  - Parameters: `clientId`, `status`, `matchTerm`, `limit`, `offset`
  - API: `GET /{scopeId}/devices`

- Tool: `kapua-update-device`
  - Updates an existing device in the authenticated scope
  - Parameters: `deviceId`, `device` (Device payload)
  - API: `PUT /{scopeId}/devices/{deviceId}`

- Tool: `kapua-delete-device`
  - Deletes a device in the authenticated scope
  - Parameters: `deviceId`
  - API: `DELETE /{scopeId}/devices/{deviceId}`

- Tool: `kapua-configurations-read`
  - Reads all component configurations for a device
  - Parameters: `deviceId` (object input: `{ "deviceId": "..." }`)
  - Returns: typed `DeviceConfiguration` JSON (component configurations)
  - API: `GET /{scopeId}/devices/{deviceId}/configurations`

- Resource: `kapua://devices`
  - Registered and discoverable via MCP `resources/list`
  - Readable via MCP `resources/read`
  - Returns JSON with up to 100 devices for default scope `AQ`
  - MIME type: `application/json`

## Authentication and Token Refresh
- On startup, the server authenticates with Kapua using username/password.
- Automatic refresh before expiry; on `401 Unauthorized`, it forces a token refresh and retries once.
- If the access token is expired and the refresh token is also expired/missing, the client performs a full re-authentication.

## API Spec
The Kapua REST API surface used by this server is documented in `specs/kapua_openapi.yaml`.

## Notes
- Docker files and extra scripts are not included yet; the Makefile builds a local binary.
 - MCP tool inputs must be JSON objects; even single-value inputs are wrapped (e.g., `{ "deviceId": "..." }`).
