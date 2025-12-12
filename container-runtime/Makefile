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

# # --- BUILD COMMANDS ---
# # Build the binary with version info
# build:
# 	@echo "Building $(BINARY_NAME) $(VERSION)..."
# 	@echo "Commit: $(GIT_COMMIT)"
# 	@echo "Date: $(BUILD_DATE)"
# 	@mkdir -p $(BIN_DIR)
# 	$(GO) build $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)
# 	@echo "Build complete: $(BIN_DIR)/$(BINARY_NAME)"
# 	@echo "Usage: sudo $(BIN_DIR)/$(BINARY_NAME) run <container-name> <config.json>"

# # Build with race detection (for testing)
# build-race:
# 	@echo "Building $(BINARY_NAME) with race detection..."
# 	@mkdir -p $(BIN_DIR)
# 	$(GO) build -race $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-race $(CMD_DIR)


# # --- TEST COMMANDS ---
# test: test-unit
# 	@echo "All tests passed!"

# test-unit:
# 	@echo "Running unit tests..."
# 	$(GO) test -v ./internal/...

# test-integration: build
# 	@echo "Running integration tests..."
# 	$(GO) test -v -tags=integration ./tests/...

# test-coverage:
# 	@echo "Running tests with coverage..."
# 	$(GO) test -coverprofile=coverage.out ./internal/... ./tests/...
# 	$(GO) tool cover -html=coverage.out -o coverage.html
# 	@echo "Coverage report generated: coverage.html"

# test-all: test-unit test-integration
# 	@echo "All unit and integration tests completed!"


# # --- RUN COMMANDS ---
# # Quick development run (without building binary)
# run-dev:
# 	@echo "Running mrunc in development mode with default config..."
# 	@echo "Config: $(DEFAULT_CONFIG)"
# 	@if [ ! -f "$(DEFAULT_CONFIG)" ]; then echo "Error: Config file $(DEFAULT_CONFIG) not found"; exit 1; fi
# 	sudo $(GO) run $(CMD_DIR) run dev $(DEFAULT_CONFIG)

# # Run with Ubuntu container config
# run-ubuntu:
# 	@echo "Running mrunc with Ubuntu config..."
# 	@if [ ! -f "$(DEFAULT_CONFIG)" ]; then echo "Error: Config file $(DEFAULT_CONFIG) not found"; exit 1; fi
# 	sudo $(GO) run $(CMD_DIR) run ubuntu $(DEFAULT_CONFIG)

# # Run with custom config (usage: make run-custom NAME=mycontainer CONFIG=path/to/config.json)
# run-custom:
# 	@if [ -z "$(NAME)" ]; then echo "Usage: make run-custom NAME=<container-name> CONFIG=path/to/config.json"; exit 1; fi
# 	@if [ -z "$(CONFIG)" ]; then echo "Usage: make run-custom NAME=<container-name> CONFIG=path/to/config.json"; exit 1; fi
# 	@if [ ! -f "$(CONFIG)" ]; then echo "Error: Config file $(CONFIG) not found"; exit 1; fi
# 	@echo "Running mrunc with custom config: $(CONFIG)"
# 	sudo $(GO) run $(CMD_DIR) run $(NAME) $(CONFIG)

# # Run built binary with default config
# run: build
# 	@echo "Running built binary with default config..."
# 	@if [ ! -f "$(DEFAULT_CONFIG)" ]; then echo "Error: Config file $(DEFAULT_CONFIG) not found"; exit 1; fi
# 	sudo ./$(BIN_DIR)/$(BINARY_NAME) run default $(DEFAULT_CONFIG)

# # Run built binary with custom config (usage: make run-bin NAME=mycontainer CONFIG=path/to/config.json)
# run-bin: build
# 	@if [ -z "$(NAME)" ]; then echo "Usage: make run-bin NAME=<container-name> CONFIG=path/to/config.json"; exit 1; fi
# 	@if [ -z "$(CONFIG)" ]; then echo "Usage: make run-bin NAME=<container-name> CONFIG=path/to/config.json"; exit 1; fi
# 	@if [ ! -f "$(CONFIG)" ]; then echo "Error: Config file $(CONFIG) not found"; exit 1; fi
# 	@echo "Running built binary with config: $(CONFIG)"
# 	sudo ./$(BIN_DIR)/$(BINARY_NAME) run $(NAME) $(CONFIG)

# # Create a container with default config (usage: make create NAME=mycontainer)
# create: build
# 	@if [ -z "$(NAME)" ]; then echo "Usage: make create NAME=<container-name> [CONFIG=path/to/config.json]"; exit 1; fi
# 	@CONFIG_FILE=$${CONFIG:-$(DEFAULT_CONFIG)}; \
# 	if [ ! -f "$$CONFIG_FILE" ]; then echo "Error: Config file $$CONFIG_FILE not found"; exit 1; fi; \
# 	echo "Creating container '$(NAME)' with config: $$CONFIG_FILE"; \
# 	sudo ./$(BIN_DIR)/$(BINARY_NAME) create $(NAME) $$CONFIG_FILE

# # Show version information
# version:
# 	@echo "Version: $(VERSION)"
# 	@echo "Commit: $(GIT_COMMIT)"
# 	@echo "Build Date: $(BUILD_DATE)"

# # Clean build artifacts
# clean:
# 	@echo "Cleaning build artifacts..."
# 	@rm -rf $(BIN_DIR)
# 	@$(GO) clean

# # Development helpers
# dev-setup:
# 	@echo "Setting up development environment..."
# 	$(GO) mod download
# 	$(GO) mod tidy
# 	@echo "Development setup complete"

# # Install binary to GOPATH (for local development)
# install: build
# 	@echo "Installing $(BINARY_NAME)..."
# 	$(GO) install $(CMD_DIR)
# 	@echo "$(BINARY_NAME) installed to $(shell $(GO) env GOPATH)/bin/"


# Show help
help:
	@echo "MRUNC Container Runtime - Available Commands"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "   \033[36m%-20s\033[0m %s\n", "Target", "Description"} \
		/^[a-zA-Z_-]+:.*?##/ { printf "   \033[36m%-20s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "    \n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo ""
	@echo "VERSION INFO:"
	@echo "  Current Version: $(VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Build Date: $(BUILD_DATE)"

##@ Build Commands
build: ## Build the mrunc binary
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)

build-race: ## Build with race detection
	@echo "Building $(BINARY_NAME) with race detection..."
	@mkdir -p $(BIN_DIR)
	$(GO) build -race $(BUILD_FLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-race $(CMD_DIR)

version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

##@ Run Commands
run-dev: ## Quick run with default config (go run)
	@echo "Running mrunc in development mode..."
	@if [ ! -f "$(DEFAULT_CONFIG)" ]; then echo "Error: Config file $(DEFAULT_CONFIG) not found"; exit 1; fi
	sudo $(GO) run $(CMD_DIR) run dev $(DEFAULT_CONFIG)

run-ubuntu: ## Run with Ubuntu config (go run)
	@echo "Running mrunc with Ubuntu config..."
	@if [ ! -f "$(DEFAULT_CONFIG)" ]; then echo "Error: Config file $(DEFAULT_CONFIG) not found"; exit 1; fi
	sudo $(GO) run $(CMD_DIR) run ubuntu $(DEFAULT_CONFIG)

run: build ## Run built binary with default config
	@echo "Running built binary..."
	@if [ ! -f "$(DEFAULT_CONFIG)" ]; then echo "Error: Config file $(DEFAULT_CONFIG) not found"; exit 1; fi
	sudo ./$(BIN_DIR)/$(BINARY_NAME) run default $(DEFAULT_CONFIG)

##@ Container Commands
create: build ## Create container (Usage: make create NAME=<name> [CONFIG=path])
	@if [ -z "$(NAME)" ]; then echo "Usage: make create NAME=<container-name> [CONFIG=path]"; exit 1; fi
	@CONFIG_FILE=$${CONFIG:-$(DEFAULT_CONFIG)}; \
	if [ ! -f "$$CONFIG_FILE" ]; then echo "Error: Config file $$CONFIG_FILE not found"; exit 1; fi; \
	sudo ./$(BIN_DIR)/$(BINARY_NAME) create $(NAME) $$CONFIG_FILE

##@ Test Commands
test: test-unit ## Run all unit tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	$(GO) test -v ./internal/...

test-integration: build ## Run integration tests
	@echo "Running integration tests..."
	$(GO) test -v -tags=integration ./tests/...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(GO) test -coverprofile=coverage.out ./internal/... ./tests/...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-all: test-unit test-integration ## Run all tests

##@ Utility Commands
clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR) coverage.out coverage.html
	@$(GO) clean

install: build ## Install binary to GOPATH
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(CMD_DIR)

dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	$(GO) mod download
	$(GO) mod tidy
