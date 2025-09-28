# kapua-mcp-server
MCP Server for Eclipse Kapua IoT Device Management.

This project is developed with support from OpenAI Codex.

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
├── specs/                          # REST API (OpenAPI)
├── Dockerfile                      # Multi-arch container build
├── .dockerignore                   # Docker build context filter
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
- `MCP_ALLOWED_ORIGINS` (optional): comma-separated list of additional origins allowed to call the HTTP Stream endpoint. Defaults include loopback hosts (`localhost`, `127.0.0.1`, `host.docker.internal`) for any port. Set to `*` to disable origin checks.

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

## Container Image

A multi-stage Dockerfile is provided for container deployments. Build the image from the project root:

```bash
docker build -t kapua-mcp-server .
```

Run the container by supplying the Kapua credentials via environment variables and exposing the MCP port (defaults to `8000`):

```bash
docker run --rm \
  -e KAPUA_API_ENDPOINT=https://api-sbx.everyware.io/v1 \
  -e KAPUA_USER=your-user \
  -e KAPUA_PASSWORD=your-password \
  -p 8000:8000 \
  kapua-mcp-server
```

or more simply:
```bash
docker run --rm \
  --env-file ./.venv \
  -p 8000:8000 \
  kapua-mcp-server
```

The image is based on `gcr.io/distroless/base-debian12:nonroot`; no shell is available in the container. Use `docker logs` for runtime inspection.

> **Multi-architecture:** The Dockerfile honours BuildKit's `TARGETOS`/`TARGETARCH`. Building on Apple Silicon (`arm64`) or passing `--platform` via `docker buildx build --platform linux/amd64 .` produces a matching binary.

> **Origin-handling**: Origin validation follows the MCP HTTP Stream specification. When running behind Docker, ensure the client connects using an allowed host (defaults cover loopback and `host.docker.internal`) or extend `MCP_ALLOWED_ORIGINS`.


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
- `kapua-update-device` [1]— update an existing device (`PUT /{scopeId}/devices/{deviceId}`).
- `kapua-delete-device` [1]— delete a device (`DELETE /{scopeId}/devices/{deviceId}`).

### Device Events
- `kapua-list-device-events` — enumerate device log events with optional filters for time range, resource, pagination, and sort options (`GET /{scopeId}/devices/{deviceId}/events`).

### Data Clients
- `kapua-list-data-messages` — list data messages with optional filters (multiple `clientId`, `channel`, pagination) (`GET /{scopeId}/data/messages`).

### Device Logs
- `kapua-list-device-logs` — list device logs with optional channel and property filters (`GET /{scopeId}/deviceLogs`).
  - _Availability:_ This tool requires an Eurotech Everyware Cloud endpoint; open-source Kapua deployments do not expose `/deviceLogs` and the tool will return guidance instead of data.

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
- Kapua client helpers exposed by `KapuaClient` back the MCP tools listed above:
  - `ListDeviceLogs` → `GET /{scopeId}/deviceLogs`
  - `ListDataMessages` → `GET /{scopeId}/data/messages` (supports multiple `clientId` query parameters)
  - `GetDataMessage` → `GET /{scopeId}/data/messages/{datastoreMessageId}`
- Each helper accepts common pagination parameters and surfaces Kapua errors for precise handling.
- Extend these helpers with additional MCP endpoints whenever new Kapua features are needed.

## Authentication and Token Refresh
- On startup, the server authenticates with Kapua using username/password.
- Automatic refresh before expiry; on `401 Unauthorized`, it forces a token refresh and retries once.
- If the access token is expired and the refresh token is also expired/missing, the client performs a full re-authentication.

## API Spec
- `specs/kapua_openapi.yaml` — community Kapua REST interface used by most tools.
- `specs/ec_openapi.yaml` — Everyware Cloud-specific extensions (e.g., `/deviceLogs`). Device log support relies on this specification and is unavailable on vanilla Kapua.

## Notes
[1]: Not tested.