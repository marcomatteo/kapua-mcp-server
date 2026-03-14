# kapua-mcp-server

An MCP server that connects AI assistants to [Eclipse Kapua](https://eclipse.dev/kapua/) and [Eurotech Everyware Cloud](https://ec.eurotech.com) IoT platforms — enabling natural-language device management, telemetry exploration, and fleet diagnostics.

[![Go 1.23+](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![MCP](https://img.shields.io/badge/MCP-compatible-blue)](https://modelcontextprotocol.io)

## Why use this?

Instead of navigating dashboards, writing curl commands, or learning the Kapua REST API, you can manage your IoT fleet through conversation:

> **You:** "Show me all disconnected devices"
> **Assistant:** _calls `kapua-devices-list` with status=DISCONNECTED_ — lists 3 offline gateways
>
> **You:** "What happened to gateway-07 in the last hour?"
> **Assistant:** _calls `kapua-device-events-list`_ — shows a disconnect event 42 minutes ago
>
> **You:** "Show me its current configuration"
> **Assistant:** _calls `kapua-device-configurations-read`_ — displays all component configs
>
> **You:** "Roll it back to the previous snapshot"
> **Assistant:** _calls `kapua-device-snapshot-rollback`_ — triggers rollback successfully

This turns multi-step diagnostic workflows into a guided conversation without context-switching away from your IDE or terminal.

## Quick Start

### 1. Configure credentials

Create a `.venv` file in the project root:

```bash
cat <<'EOF' > .venv
KAPUA_API_ENDPOINT=https://kapua.example.com/api
KAPUA_USER=my-user
KAPUA_PASSWORD=my-password
EOF
```

### 2. Build and run

```bash
make build
./bin/kapua-mcp-server          # stdio transport (default)
./bin/kapua-mcp-server -http    # HTTP transport on localhost:8000
```

Or without building: `go run ./cmd/server`

### Docker

```bash
docker build -t kapua-mcp-server .

docker run --rm \
  -e KAPUA_API_ENDPOINT=https://kapua.example.com/api \
  -e KAPUA_USER=my-user \
  -e KAPUA_PASSWORD=my-password \
  -p 8000:8000 \
  kapua-mcp-server
```

The image is based on `gcr.io/distroless/base-debian12:nonroot` and supports multi-architecture builds (amd64/arm64).

## Configuration

| Variable | Required | Default | Description |
|---|---|---|---|
| `KAPUA_API_ENDPOINT` | Yes | — | Kapua REST API base URL |
| `KAPUA_USER` | Yes | — | Kapua username |
| `KAPUA_PASSWORD` | Yes | — | Kapua password |
| `MCP_ALLOWED_ORIGINS` | No | common local hosts (`localhost`, `127.0.0.1`, `::1`, `0.0.0.0`, `host.docker.internal`) | Comma-separated allowed origins for HTTP mode (both HTTP/HTTPS variants, with and without the default port). Set `*` to disable checks. |
| `LOG_LEVEL` | No | `INFO` | Log level: `DEBUG`, `INFO`, `WARN`, `ERROR` |

Settings can be provided as environment variables or in a `.venv` file (one `KEY=VALUE` per line). Environment variables take precedence.

## MCP Client Setup

### Claude Desktop

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "kapua-mcp-server": {
      "command": "/path/to/kapua-mcp-server",
      "args": [],
      "env": {
        "KAPUA_API_ENDPOINT": "https://kapua.example.com/api",
        "KAPUA_USER": "my-user",
        "KAPUA_PASSWORD": "my-password"
      }
    }
  }
}
```

### Claude Code

Add to your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "kapua-mcp-server": {
      "command": "/path/to/kapua-mcp-server",
      "args": [],
      "env": {
        "KAPUA_API_ENDPOINT": "https://kapua.example.com/api",
        "KAPUA_USER": "my-user",
        "KAPUA_PASSWORD": "my-password"
      }
    }
  }
}
```

### HTTP clients

For HTTP-based setups, start the server with `-http` and point your MCP client to `http://localhost:8000` (or `http://host.docker.internal:8000` from Docker on macOS/Windows).

## Available Tools

### Devices

| Tool | Description |
|---|---|
| `kapua-devices-list` | List devices with filters: `clientId`, `status` (CONNECTED/DISCONNECTED/MISSING), `matchTerm`, pagination |

### Telemetry

| Tool | Description |
|---|---|
| `kapua-data-messages-list` | Query telemetry data messages by channel, time range, client IDs |

### Events & Logs

| Tool | Description |
|---|---|
| `kapua-device-events-list` | List device lifecycle events with time range, resource, and sort filters |
| `kapua-device-logs-list` | List device logs (Everyware Cloud only; not available on open-source Kapua) |

### Configuration & Snapshots

| Tool | Description |
|---|---|
| `kapua-device-configurations-read` | Read all component configurations for a device |
| `kapua-device-snapshots-list` | List available configuration snapshots |
| `kapua-device-snapshot-configurations-read` | Read the configuration stored in a specific snapshot |
| `kapua-device-snapshot-rollback` | Rollback a device to a previous snapshot |

### Device Inventory

| Tool | Description |
|---|---|
| `kapua-device-inventory-read` | General inventory summary for a device |
| `kapua-device-inventory-bundles-list` | List OSGi bundles |
| `kapua-device-inventory-bundle-start` | Trigger bundle inventory collection |
| `kapua-device-inventory-bundle-stop` | Stop bundle inventory collection |
| `kapua-device-inventory-containers-list` | List containers (Docker, etc.) |
| `kapua-device-inventory-container-start` | Trigger container inventory collection |
| `kapua-device-inventory-container-stop` | Stop container inventory collection |
| `kapua-device-inventory-system-packages-list` | List OS-level system packages |
| `kapua-device-inventory-deployment-packages-list` | List application deployment packages |

## Available Resources

| Resource URI | Description |
|---|---|
| `kapua://devices` | Live JSON list of devices in the current scope |
| `kapua://fleet-health` | Aggregated fleet health: online/offline counts, stale devices, critical events. Tunable via `staleMinutes` and `criticalMinutes` (default: 60). |

## Architecture

```
kapua-mcp-server/
├── cmd/server/             # CLI entry point, HTTP logging middleware
├── internal/
│   ├── kapua/
│   │   ├── config/         # Configuration loader (.venv + env vars)
│   │   ├── handlers/       # MCP tool and resource implementations
│   │   ├── models/         # Kapua API data models
│   │   └── services/       # REST client, auth, pagination
│   └── mcp/                # MCP server wiring, HTTP transport, origin guard
├── pkg/utils/              # Structured logger
├── specs/                  # OpenAPI specs (Kapua + Everyware Cloud)
├── Dockerfile              # Multi-arch container build
└── Makefile
```

**Key design decisions:**
- **Authentication:** JWT with automatic token refresh (5 min before expiry) and full re-auth fallback
- **Pagination:** Generic paginator that follows Kapua's `limitExceeded` flag across all endpoints
- **Transports:** Stdio (default, recommended for local use) or Streamable HTTP with CORS origin validation
- **Concurrency:** Fleet health uses goroutine pools for parallel event fetching; thread-safe token management

## Development

**Requirements:** Go 1.23+

```bash
make build          # Build binary to bin/
make test           # Run tests with coverage
make run            # Run the built binary
make clean          # Remove build artifacts
```

Generate and view a coverage report:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## API Specifications

- `specs/kapua_openapi.yaml` — Eclipse Kapua REST API
- `specs/ec_openapi.yaml` — Eurotech Everyware Cloud extensions (e.g., device logs)

## License

[MIT](LICENSE) — Marco Matteo Buzzulini
