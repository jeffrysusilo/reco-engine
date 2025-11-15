.PHONY: build test clean docker-build docker-up docker-down

# Build all services
build:
	@echo "Building services..."
	@go build -o bin/ingest.exe ./cmd/ingest
	@go build -o bin/processor.exe ./cmd/processor
	@go build -o bin/api.exe ./cmd/api
	@echo "Build complete!"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@if exist bin rmdir /s /q bin
	@if exist coverage.out del coverage.out
	@if exist coverage.html del coverage.html

# Docker operations
docker-build:
	@echo "Building Docker images..."
	@docker-compose build

docker-up:
	@echo "Starting services..."
	@docker-compose up -d

docker-down:
	@echo "Stopping services..."
	@docker-compose down

docker-logs:
	@docker-compose logs -f

# Run specific service locally
run-ingest:
	@go run ./cmd/ingest

run-processor:
	@go run ./cmd/processor

run-api:
	@go run ./cmd/api

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@golangci-lint run

# Generate mocks (requires mockgen)
mocks:
	@echo "Generating mocks..."
	@go generate ./...
