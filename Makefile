all: build run

APP := kapua-mcp-server
VERSION := 0.1.0

.PHONY: run build clean test lint help

.PHONY: run
run:
	./bin/$(APP)

.PHONY: build
build:
	GO111MODULE=on go build -o bin/$(APP) ./cmd/server

.PHONY: clean
clean:
	rm -rf bin/*

test:
	go test ./...

# lint:
# 	golangci-lint run

# Help information
.PHONY: help
help:
	@echo "Usage:"
	@echo "  make               - Build and run"
	@echo "  make build         - Build for current platform"
	@echo "  make clean         - Clean build outputs"
	@echo "  make run		    - Run the application"
	@echo "  make help          - Show this help information"