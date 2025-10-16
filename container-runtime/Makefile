# MRUNC Container Runtime Makefile

.PHONY: all build clean test test-integration test-namespaces test-filesystem test-lifecycle benchmark security-scan lint help run-dev run-ubuntu create

# Build configuration
BINARY_NAME=mrunc
CMD_DIR=./cmd/mrunc
BIN_DIR=./bin
PKG_DIR=./pkg/...
INTERNAL_DIR=./internal/...

# Configuration
CONFIG_DIR=./configs/examples
DEFAULT_CONFIG=$(CONFIG_DIR)/ubuntu.json
TEST_CONFIG=$(CONFIG_DIR)/test-container.json

# Go configuration  
GO=go
GOOS=$(shell $(GO) env GOOS)
GOARCH=$(shell $(GO) env GOARCH)

# Version information (automatically derived from git)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags with version injection
LDFLAGS=-ldflags "\
	-s -w \
	-X 'mrunc/internal/cli.Version=$(VERSION)' \
	-X 'mrunc/internal/cli.GitCommit=$(GIT_COMMIT)' \
	-X 'mrunc/internal/cli.BuildDate=$(BUILD_DATE)'"
BUILD_FLAGS=-trimpath $(LDFLAGS)

# Default target
all: clean build test

# Build the binary with version info
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@echo "Commit: $(GIT_COMMIT)"
	@echo "Date: $(BUILD_DATE)"
	@mkdir -p $(BIN_DIR)
	$(GO) build $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: $(BIN_DIR)/$(BINARY_NAME)"
	@echo "Usage: sudo $(BIN_DIR)/$(BINARY_NAME) run <config.json>"

# Build with race detection (for testing)
build-race:
	@echo "Building $(BINARY_NAME) with race detection..."
	@mkdir -p $(BIN_DIR)
	$(GO) build -race $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-race $(CMD_DIR)

# Quick development run (without building binary)
run-dev:
	@echo "Running mrunc in development mode with default config..."
	@echo "Config: $(DEFAULT_CONFIG)"
	@if [ ! -f "$(DEFAULT_CONFIG)" ]; then echo "Error: Config file $(DEFAULT_CONFIG) not found"; exit 1; fi
	sudo $(GO) run $(CMD_DIR) run $(DEFAULT_CONFIG)

# Run with Ubuntu container config
run-ubuntu:
	@echo "Running mrunc with Ubuntu config..."
	@if [ ! -f "$(DEFAULT_CONFIG)" ]; then echo "Error: Config file $(DEFAULT_CONFIG) not found"; exit 1; fi
	sudo $(GO) run $(CMD_DIR) run $(DEFAULT_CONFIG)

# Run with custom config (usage: make run-custom CONFIG=path/to/config.json)
run-custom:
	@if [ -z "$(CONFIG)" ]; then echo "Usage: make run-custom CONFIG=path/to/config.json"; exit 1; fi
	@if [ ! -f "$(CONFIG)" ]; then echo "Error: Config file $(CONFIG) not found"; exit 1; fi
	@echo "Running mrunc with custom config: $(CONFIG)"
	sudo $(GO) run $(CMD_DIR) run $(CONFIG)

# Run built binary with default config
run: build
	@echo "Running built binary with default config..."
	@if [ ! -f "$(DEFAULT_CONFIG)" ]; then echo "Error: Config file $(DEFAULT_CONFIG) not found"; exit 1; fi
	sudo ./$(BIN_DIR)/$(BINARY_NAME) run $(DEFAULT_CONFIG)

# Run built binary with custom config (usage: make run-bin CONFIG=path/to/config.json)
run-bin: build
	@if [ -z "$(CONFIG)" ]; then echo "Usage: make run-bin CONFIG=path/to/config.json"; exit 1; fi
	@if [ ! -f "$(CONFIG)" ]; then echo "Error: Config file $(CONFIG) not found"; exit 1; fi
	@echo "Running built binary with config: $(CONFIG)"
	sudo ./$(BIN_DIR)/$(BINARY_NAME) run $(CONFIG)

# Create a container with default config (usage: make create NAME=mycontainer)
create: build
	@if [ -z "$(NAME)" ]; then echo "Usage: make create NAME=<container-name> [CONFIG=path/to/config.json]"; exit 1; fi
	@CONFIG_FILE=$${CONFIG:-$(DEFAULT_CONFIG)}; \
	if [ ! -f "$$CONFIG_FILE" ]; then echo "Error: Config file $$CONFIG_FILE not found"; exit 1; fi; \
	echo "Creating container '$(NAME)' with config: $$CONFIG_FILE"; \
	sudo ./$(BIN_DIR)/$(BINARY_NAME) create $(NAME) $$CONFIG_FILE

# Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@$(GO) clean

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "Development setup complete"

# Install binary to GOPATH (for local development)
install: build
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(CMD_DIR)
	@echo "$(BINARY_NAME) installed to $(shell $(GO) env GOPATH)/bin/"

# Show help
help:
	@echo "MRUNC Container Runtime - Build & Test Commands"
	@echo ""
	@echo "BUILD COMMANDS:"
	@echo "  build          - Build the binary with version info"
	@echo "  build-race     - Build with race detection"
	@echo "  version        - Show version information"
	@echo ""
	@echo "RUN COMMANDS:"
	@echo "  run-dev        - Quick run with go run (default config)"
	@echo "  run-ubuntu     - Run with Ubuntu config via go run"
	@echo "  run-custom     - Run with custom config via go run"
	@echo "                   Usage: make run-custom CONFIG=path/to/config.json"
	@echo "  run            - Run built binary with default config"
	@echo "  run-bin        - Run built binary with custom config"
	@echo "                   Usage: make run-bin CONFIG=path/to/config.json"
	@echo ""
	@echo "CONTAINER COMMANDS:"
	@echo "  create         - Create a container"
	@echo "                   Usage: make create NAME=<container-name> [CONFIG=path/to/config.json]"
	@echo "                   Example: make create NAME=mycontainer"
	@echo "                   Example: make create NAME=mycontainer CONFIG=configs/examples/ubuntu.json"
	@echo ""
	@echo "UTILITY COMMANDS:"
	@echo "  clean          - Remove build artifacts"
	@echo "  install        - Install binary locally"
	@echo "  dev-setup      - Setup development environment"
	@echo ""
	@echo "VERSION INFO:"
	@echo "  Current Version: $(VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Build Date: $(BUILD_DATE)"
