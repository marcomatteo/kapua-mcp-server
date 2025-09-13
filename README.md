# kapua-mcp-server
MCP Server for Eclipse Kapua for IoT Device Management tools

## Project structure

```
kapua-mcp-server/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go           # Configuration management
│   ├── handlers/
│   │   ├── mcp.go              # MCP protocol handlers
│   │   └── kapua.go            # Kapua API integration
│   ├── models/
│   │   ├── mcp.go              # MCP message structures
│   │   └── kapua.go            # Kapua data structures
│   └── services/
│       ├── kapua_client.go     # Kapua REST API client
│       └── mcp_server.go       # MCP server implementation
├── pkg/
│   └── utils/
│       └── logger.go           # Logging utilities
├── api/
│   └── openapi.yaml            # API documentation
├── deployments/
│   ├── Dockerfile              # Docker configuration
│   └── docker-compose.yml      # Local development setup
├── scripts/
│   ├── build.sh               # Build scripts
│   └── run.sh                 # Run scripts
├── docs/
│   ├── README.md              # Project documentation
│   └── SETUP.md               # Setup instructions
├── .gitignore
├── go.mod                     # Go module file
├── go.sum                     # Go dependencies
└── Makefile                   # Build automation
```

## Claude configuration

```json
{
  "mcpServers": {
    "kapua-mcp-server": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-p",
        "8080:8080",
        "-e",
        "MCP_AUTH_TOKEN=secret",
        "-e",
        "MCP_ALLOWED_ORIGINS='http://localhost:8080'",
        "kapua-mcp-server:0.1.0"
      ]
    }
  }
}
```