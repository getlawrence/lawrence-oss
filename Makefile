.PHONY: all ui build run docker test clean deps docker-build docker-run docker-dev docker-stop docker-clean

# Variables
BINARY_NAME=lawrence
BUILD_DIR=bin
UI_DIR=ui
DATA_DIR=data

all: ui build

# Install dependencies
deps:
	go mod download
	cd $(UI_DIR) && npm install

# Build UI
ui:
	cd $(UI_DIR) && npm install && npm run build

# Build Go binary
build: ui
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/all-in-one

# Build for Linux (for Docker)
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux ./cmd/all-in-one

# Run locally
run: build
	@mkdir -p $(DATA_DIR)
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run with config
run-config: build
	@mkdir -p $(DATA_DIR)
	./$(BUILD_DIR)/$(BINARY_NAME) --config lawrence.yaml

# Build Docker image
docker:
	docker build -t lawrence/all-in-one:latest .

# Run Docker container
docker-run:
	docker run -p 8080:8080 -p 4317:4317 -p 4318:4318 \
		-v $(PWD)/$(DATA_DIR):/data \
		lawrence/all-in-one:latest

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(UI_DIR)/dist
	rm -rf $(DATA_DIR)
	rm -f coverage.out coverage.html

# Development mode (watch and reload)
dev:
	air

# Install development tools
install-tools:
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# =============================================================================
# Docker Commands
# =============================================================================

# Build Docker image
docker-build:
	docker build -t lawrence-oss:latest .

# Run with Docker Compose (production)
docker-run:
	docker compose up -d

# Run with Docker Compose (development)
docker-dev:
	docker compose -f docker-compose.dev.yml up -d

# Stop all containers
docker-stop:
	docker compose down
	docker compose -f docker-compose.dev.yml down

# Clean up Docker resources
docker-clean:
	docker compose down -v
	docker compose -f docker-compose.dev.yml down -v
	docker system prune -f
	docker volume prune -f

# Build and run in one command
docker-quick:
	docker build -t lawrence-oss:latest . && docker compose up -d

# View logs
docker-logs:
	docker compose logs -f

# View development logs
docker-logs-dev:
	docker compose -f docker-compose.dev.yml logs -f

# Shell into container
docker-shell:
	docker compose exec lawrence sh

# Shell into development container
docker-shell-dev:
	docker compose -f docker-compose.dev.yml exec lawrence sh
