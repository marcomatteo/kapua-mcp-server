# Release Notes

## 1.0.0

- Restructured configuration so Kapua settings live under `internal/kapua/config`, with separate HTTP transport helpers in `internal/mcp/http_config.go`.
- Simplified the server CLI: stdio is the default transport and the new `-http` flag opts into the streamable HTTP mode while still allowing `-host`/`-port` overrides.
- Updated MCP server wiring to build once and reuse across transports, including HTTP origin-guard middleware now accepting explicit HTTP configuration.
- Refreshed documentation to describe the preferred local troubleshooting workflow, updated project layout, and clarified `.venv` usage for stdio credentials.
