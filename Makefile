all: build run

APP := kapua-mcp-server
VERSION := 0.1.0

.PHONY: run build clean test lint help

run:
	docker run \
	--rm \
	-p 8080:8080 \
	$(APP):$(VERSION)

build:
	docker build -f Dockerfile -t $(APP):$(VERSION) .

clean:
	docker rmi $(APP):$(VERSION) || true

# test:
# 	go test ./...

# lint:
# 	golangci-lint run

# Help information
.PHONY: help
help:
	@echo "Usage:"
	@echo "  make               - Build and run"
	@echo "  make build         - Build for current platform"
	@echo "  make clean         - Clean build outputs"
	@echo "  make test          - Run tests"
	@echo "  make lint          - Run linter"
	@echo "  make run		    - Run the application in a Docker container"
	@echo "  make help          - Show this help information"