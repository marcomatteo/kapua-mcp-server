FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git for module downloads
RUN apk add --no-cache git

# Copy go module files first to leverage Docker layer caching
COPY go.mod .
RUN go mod download || true

# Copy the rest of the source
COPY . .

# Ensure dependencies are resolved (writes go.sum inside image)
RUN go mod tidy

# Build a static binary for linux
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:3.20 AS final
WORKDIR /app

COPY --from=builder /app/server /app/server

EXPOSE 8080

# Use exec form and pass default args via CMD
ENTRYPOINT ["/app/server", "server"]
