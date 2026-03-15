# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`kapua-mcp-server` is a Model Context Protocol (MCP) server for Eclipse Kapua IoT device management, written in Go 1.23+. It exposes 24 MCP tools and 2 MCP resources that allow AI assistants to manage IoT devices, query telemetry, inspect configurations, and assess fleet health.

## Commands

```bash
# Build
make build          # Compiles to bin/kapua-mcp-server

# Run
make run            # Runs the built binary (stdio transport by default)
./bin/kapua-mcp-server -http              # HTTP transport on localhost:8000
./bin/kapua-mcp-server -http -host 0.0.0.0 -port 9000

# Test
make test                                # All tests with coverage report
go test ./... -race                      # With race detector (as in CI)
go test ./internal/kapua/handlers/...   # Single package
go test -run TestFunctionName ./...     # Single test

# Lint
make lint           # golangci-lint run ./...
```

## Architecture

The codebase follows a clean separation of concerns under three top-level packages:

```
cmd/server/          - Entry point; selects stdio vs. HTTP transport
internal/
  mcp/               - MCP server wiring, tool/resource registration, health endpoint
  kapua/
    config/          - .env + env var loading, auth method validation
    handlers/        - MCP tool implementations (one file per Kapua domain)
    services/        - REST client, JWT token management, business logic
    models/          - Data structures for Kapua API responses
pkg/utils/           - Structured logger
```

**Request flow:** `MCP tool call → handler → service → Kapua REST API → model → response`

### Key design points

- **Authentication**: `services/kapua_client.go` manages JWT tokens with automatic refresh 5 minutes before expiry (fallback to full re-auth). Thread-safe via `sync.RWMutex`. Supports both `password` and `apikey` methods.
- **Transports**: stdio (default, for Claude Desktop) and streamable HTTP with CORS origin validation (`internal/mcp/origin_guard.go`). Health check at `/health`.
- **Concurrency**: Fleet health resource fetches events in parallel goroutine pools.
- **Generated code**: `openapi-generated/` is excluded from linting — do not edit manually.
- **Pagination**: Handlers honor Kapua's `limitExceeded` flag and expose limit/offset parameters.

### Adding a new tool

1. Add model structs in `internal/kapua/models/`
2. Add service methods in `internal/kapua/services/`
3. Add a handler file in `internal/kapua/handlers/`
4. Register the tool in `internal/mcp/server.go`

## Configuration

Copy `.env.example` to `.env` and set:

| Variable | Required | Default | Description |
|---|---|---|---|
| `KAPUA_API_ENDPOINT` | Yes | — | Kapua REST API base URL |
| `KAPUA_AUTH_METHOD` | No | `password` | `password` or `apikey` |
| `KAPUA_USER` / `KAPUA_PASSWORD` | If password auth | — | Credentials |
| `KAPUA_API_KEY` | If apikey auth | — | API key |
| `KAPUA_TIMEOUT` | No | `30` | HTTP timeout in seconds |
| `LOG_LEVEL` | No | `INFO` | `DEBUG`, `INFO`, `WARN`, `ERROR` |

Environment variables take precedence over `.env` file values.
