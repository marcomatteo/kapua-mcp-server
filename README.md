# kapua-mcp-server
kapua-mcp-server is an MCP server designed primarily for local troubleshooting and operator tooling when working with Eclipse Kapua or Eurotech Everyware Cloud IoT Device Management platforms. It exposes Kapua APIs through the Model Context Protocol so that assistants and diagnostics utilities can inspect devices, configurations, and telemetry without deploying additional infrastructure.

This project is developed with support from OpenAI Codex.

## Project Structure

```
kapua-mcp-server/
├── cmd/
│   └── server/
│       ├── logging_middleware.go   # HTTP request logging wrapper
│       └── main.go                 # CLI entry point (stdio default, -http optional)
├── internal/
│   ├── kapua/
│   │   ├── config/                 # Kapua-specific configuration loader
│   │   ├── handlers/               # Tool implementations (devices, telemetry, etc.)
│   │   ├── models/                 # Kapua API data models
│   │   └── services/               # Kapua REST/Authentication clients
│   └── mcp/
│       ├── http_config.go          # HTTP transport configuration helpers
│       ├── origin_guard.go         # Allowed-origin middleware
│       └── server.go               # MCP server wiring and transport helpers
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

## Quick Start (Local Binary)

1. Create a `.venv` file for credentials (keeps secrets out of your shell history):
   ```bash
   cat <<'EOF' > .venv
   KAPUA_API_ENDPOINT=https://kapua.example.com/api
   KAPUA_USER=my-user
   KAPUA_PASSWORD=We!come12345
   EOF
   ```
2. Build the binary (optional—`go run ./cmd/server` works too):
   ```bash
   make build
   ```
3. Launch the server over stdio (default; recommended for local tooling):
   ```bash
   ./bin/kapua-mcp-server
   ```

   The process remains attached to your terminal, exchanging JSON-RPC messages over standard input/output with your MCP client.

4. Switch to HTTP when you need a network-accessible endpoint:
   ```bash
   ./bin/kapua-mcp-server -http
   ```

   The HTTP transport listens on `host:port` (defaults to `localhost:8000`). You can override these with `-host` and `-port` at startup. Use `make run` to compile and start the binary in one step if you prefer.

## Quick Start (Docker)

Build a local image:

```bash
docker build -t kapua-mcp-server .
```

Run the container exposing the HTTP stream transport (ideal for remote clients). Inject your Kapua credentials as environment variables:

```bash
docker run --rm \
  -e KAPUA_API_ENDPOINT=https://kapua.example.com/api \
  -e KAPUA_USER=my-user \
  -e KAPUA_PASSWORD=We!come12345 \
  -p 8000:8000 \
  kapua-mcp-server
```

The image is based on `gcr.io/distroless/base-debian12:nonroot`; no shell is available in the container. Inspect logs with `docker logs <container>`.

> **Multi-architecture:** The Dockerfile honours BuildKit's `TARGETOS`/`TARGETARCH`. Building on Apple Silicon (`arm64`) or passing `--platform` via `docker buildx build --platform linux/amd64 .` produces a matching binary.

> **Origin-handling:** Origin validation follows the MCP HTTP Stream specification. When running behind Docker, ensure the client connects using an allowed host (defaults cover loopback and `host.docker.internal`) or extend `MCP_ALLOWED_ORIGINS`.

## MCP Client Configuration Examples

### Claude Desktop (macOS/Windows)
> Claude desktop supports only STDIO MCP servers.

1. Locate the Claude Desktop configuration file:
   - **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`
2. Add or update the `mcpServers` array with a stdio configuration:
   ```json
   {
     "mcpServers": {
        "kapua-mcp-server": {
          "command": "/Users/marco/dev/git-marcomatteo/kapua-mcp-server/bin/kapua-mcp-server",
          "args": [],
          "env": {
            "KAPUA_API_ENDPOINT": "https://api.kapua.io/",
            "KAPUA_USER": "kapua-user",
            "KAPUA_PASSWORD": "kapua-password"
          }
        }
     }
  }
   ```
3. Restart Claude Desktop. The Kapua tools appear under the **Servers** tab, and Claude will launch the Docker container when you connect.

   Replace the placeholder credential values with your Kapua settings before saving the configuration.

### Custom MCP Client

For HTTP-based setups, expose the container as shown in the Docker quick start and configure a MCP Client application to `http://host.docker.internal:8000` (macOS/Windows) or `http://127.0.0.1:8000` (Linux).


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
