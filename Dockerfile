# syntax=docker/dockerfile:1

ARG BUILDPLATFORM
ARG TARGETPLATFORM

# Build stage
FROM --platform=$BUILDPLATFORM golang:1.23-bookworm AS builder
WORKDIR /src

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

# Cache go modules first
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the source
COPY . .

# Build the Kapua MCP server for the target platform
RUN --mount=type=cache,target=/root/.cache/go-build \
    set -eux; \
    if [ "$TARGETARCH" = "arm" ] && [ -n "$TARGETVARIANT" ]; then \
        export GOARM="${TARGETVARIANT#v}"; \
    fi; \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o /out/kapua-mcp-server ./cmd/server

# Runtime stage
FROM scratch as release-slim
WORKDIR /app

# Copy binary and TLS certificates
COPY --from=builder /out/kapua-mcp-server /app/kapua-mcp-server
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8000
ENTRYPOINT ["/app/kapua-mcp-server"]
CMD ["--host", "0.0.0.0", "--port", "8000"]
