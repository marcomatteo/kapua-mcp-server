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
│       ├── models/
│       └── services/
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

## Testing and Coverage

- Run the full unit test suite:
  - `go test ./...`
  - or `make test` (wrapper around the same command).
- Generate a coverage profile while running tests:
  - `go test ./... -coverprofile=coverage.out`
- Inspect coverage results:
  - Summary in the terminal: `go tool cover -func=coverage.out`
  - Annotated HTML report: `go tool cover -html=coverage.out`

The coverage report commands reuse the `coverage.out` file produced in the previous step; delete it when no longer needed.

## MCP Tools

### Device Directory
- `kapua-list-devices` — list devices in scope using filters such as `clientId`, `status`, `matchTerm`, `limit`, `offset` (`GET /{scopeId}/devices`).
- `kapua-update-device` — update an existing device (`PUT /{scopeId}/devices/{deviceId}`).
- `kapua-delete-device` — delete a device (`DELETE /{scopeId}/devices/{deviceId}`).

### Device Events
- `kapua-list-device-events` — enumerate device log events with optional filters for time range, resource, pagination, and sort options (`GET /{scopeId}/devices/{deviceId}/events`).

### Data Clients
- `kapua-list-data-messages` — list data messages with optional filters (multiple `clientId`, `channel`, pagination) (`GET /{scopeId}/data/messages`).

### Device Configuration
- `kapua-configurations-read` — retrieve all component configurations for a device (`GET /{scopeId}/devices/{deviceId}/configurations`).

### Device Inventory
- `kapua-inventory-read` — fetch general inventory details for a device (`GET /{scopeId}/devices/{deviceId}/inventory`).
- `kapua-inventory-bundles` — list bundles in the device inventory (`GET /{scopeId}/devices/{deviceId}/inventory/bundles`).
- `kapua-inventory-bundle-start` — trigger bundle inventory collection (`POST /{scopeId}/devices/{deviceId}/inventory/bundles/_start`).
- `kapua-inventory-bundle-stop` — stop bundle inventory collection (`POST /{scopeId}/devices/{deviceId}/inventory/bundles/_stop`).
- `kapua-inventory-containers` — list container inventory entries (`GET /{scopeId}/devices/{deviceId}/inventory/containers`).
- `kapua-inventory-container-start` — trigger container inventory collection (`POST /{scopeId}/devices/{deviceId}/inventory/containers/_start`).
- `kapua-inventory-container-stop` — stop container inventory collection (`POST /{scopeId}/devices/{deviceId}/inventory/containers/_stop`).
- `kapua-inventory-system-packages` — list device system packages (`GET /{scopeId}/devices/{deviceId}/inventory/system`).
- `kapua-inventory-deployment-packages` — list deployment packages (`GET /{scopeId}/devices/{deviceId}/inventory/packages`).

## MCP Resources
- `kapua://devices` — discoverable via MCP `resources/list` and readable through `resources/read`; returns JSON with up to 100 devices for the default scope `AQ` (`application/json`).

## Kapua Client Helpers
- Data message APIs exposed by `KapuaClient` back the MCP tools listed above:
  - `ListDataMessages` → `GET /{scopeId}/data/messages` (supports multiple `clientId` query parameters)
- Each helper accepts common pagination parameters and surfaces Kapua errors for precise handling.
- Extend these helpers with additional MCP endpoints whenever new Kapua features are needed.

## Authentication and Token Refresh
- On startup, the server authenticates with Kapua using username/password.
- Automatic refresh before expiry; on `401 Unauthorized`, it forces a token refresh and retries once.
- If the access token is expired and the refresh token is also expired/missing, the client performs a full re-authentication.

## API Spec
The Kapua REST API surface used by this server is documented in `specs/kapua_openapi.yaml`.

## Notes
- Docker files and extra scripts are not included yet; the Makefile builds a local binary.
- MCP tool inputs must be JSON objects; even single-value inputs are wrapped (e.g., `{ "deviceId": "..." }`).