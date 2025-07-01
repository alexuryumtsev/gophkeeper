# Makefile for GophKeeper

# Build variables
APP_NAME=gophkeeper
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE = $(shell date -u +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE) -X main.gitCommit=$(GIT_COMMIT)"

# Binary names
BINARY_SERVER=bin/$(APP_NAME)-server
BINARY_CLIENT=bin/$(APP_NAME)-client

# Default target
.PHONY: all
all: clean deps build

# Install dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build all binaries
.PHONY: build
build: build-server build-client

# Build server
.PHONY: build-server
build-server:
	mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_SERVER) ./cmd/server

# Build client
.PHONY: build-client
build-client:
	mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_CLIENT) ./cmd/client

# Build for multiple platforms
.PHONY: build-all
build-all: clean deps
	# Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME)-server-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME)-client-linux-amd64 ./cmd/client
	# Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME)-server-windows-amd64.exe ./cmd/server
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME)-client-windows-amd64.exe ./cmd/client
	# macOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME)-server-darwin-amd64 ./cmd/server
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME)-client-darwin-amd64 ./cmd/client
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME)-server-darwin-arm64 ./cmd/server
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME)-client-darwin-arm64 ./cmd/client

# Run tests
.PHONY: test
test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
.PHONY: test-coverage
test-coverage: test
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
.PHONY: lint
lint:
	golangci-lint run

# Format code
.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html

# Run server in development mode
.PHONY: run-server
run-server: build-server
	./$(BINARY_SERVER) --jwt-secret=dev-secret-key

# Run database migrations (requires server binary)
.PHONY: migrate
migrate: build-server
	./$(BINARY_SERVER) migrate

# Docker targets
.PHONY: docker-build
docker-build:
	docker build -t $(APP_NAME):$(VERSION) .

.PHONY: docker-run
docker-run:
	docker run -p 8080:8080 $(APP_NAME):$(VERSION)

# Development database setup (PostgreSQL in Docker)
.PHONY: db-up
db-up:
	docker run --name gophkeeper-postgres \
		-e POSTGRES_DB=gophkeeper \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=password \
		-p 5432:5432 \
		-d postgres:15

.PHONY: db-down
db-down:
	docker stop gophkeeper-postgres || true
	docker rm gophkeeper-postgres || true

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Clean, install deps, and build"
	@echo "  deps         - Install dependencies"
	@echo "  build        - Build all binaries"
	@echo "  build-server - Build server binary"
	@echo "  build-client - Build client binary"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  clean        - Clean build artifacts"
	@echo "  run-server   - Run server in development mode"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  db-up        - Start PostgreSQL database in Docker"
	@echo "  db-down      - Stop PostgreSQL database"
	@echo "  help         - Show this help"