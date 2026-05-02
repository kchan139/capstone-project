# MRUNC Container Runtime Makefile

.PHONY: all build check-go-version clean test test-integration test-namespaces test-filesystem test-lifecycle benchmark security-scan lint help run-dev run-ubuntu run create start list kill kill-all delete delete-force lifecycle clean-container

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
BUNDLE_DIR ?= /tmp/mrunc-default-bundle
NAME ?= default
SIGNAL ?= TERM

# Go configuration
GO=go
GO_REQUIRED_VERSION ?= go1.25.9
export GOTOOLCHAIN ?= local
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

check-go-version:
	@test "$$($(GO) env GOVERSION)" = "$(GO_REQUIRED_VERSION)" || \
	( echo "Expected $(GO_REQUIRED_VERSION), got $$($(GO) env GOVERSION)"; exit 1 )

##@ Build Commands
build: check-go-version ## Build the mrunc binary
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
	@rm -rf "$(BUNDLE_DIR)"
	@mkdir -p "$(BUNDLE_DIR)"
	@cp "$(DEFAULT_CONFIG)" "$(BUNDLE_DIR)/config.json"
	sudo $(GO) run $(CMD_DIR) run --bundle "$(BUNDLE_DIR)" dev

run-ubuntu: ## Run with Ubuntu config (go run)
	@echo "Running mrunc with Ubuntu config..."
	@if [ ! -f "$(DEFAULT_CONFIG)" ]; then echo "Error: Config file $(DEFAULT_CONFIG) not found"; exit 1; fi
	@rm -rf "$(BUNDLE_DIR)"
	@mkdir -p "$(BUNDLE_DIR)"
	@cp "$(DEFAULT_CONFIG)" "$(BUNDLE_DIR)/config.json"
	sudo $(GO) run $(CMD_DIR) run --bundle "$(BUNDLE_DIR)" ubuntu

run: build ## Run built binary with default config
	@echo "Running built binary..."
	@if [ ! -f "$(DEFAULT_CONFIG)" ]; then echo "Error: Config file $(DEFAULT_CONFIG) not found"; exit 1; fi
	@rm -rf "$(BUNDLE_DIR)"
	@mkdir -p "$(BUNDLE_DIR)"
	@cp "$(DEFAULT_CONFIG)" "$(BUNDLE_DIR)/config.json"
	sudo ./$(BIN_DIR)/$(BINARY_NAME) run --bundle "$(BUNDLE_DIR)" "$(NAME)"

##@ Container Commands
create: build ## Create container (Usage: make create NAME=<name> [CONFIG=path])
	@CONFIG_FILE=$${CONFIG:-$(DEFAULT_CONFIG)}; \
	if [ ! -f "$$CONFIG_FILE" ]; then echo "Error: Config file $$CONFIG_FILE not found"; exit 1; fi; \
	rm -rf "$(BUNDLE_DIR)"; \
	mkdir -p "$(BUNDLE_DIR)"; \
	cp "$$CONFIG_FILE" "$(BUNDLE_DIR)/config.json"; \
	sudo ./$(BIN_DIR)/$(BINARY_NAME) create --bundle "$(BUNDLE_DIR)" $(NAME)

start: build ## Start a created container (Usage: make start NAME=<name>)
	sudo ./$(BIN_DIR)/$(BINARY_NAME) start "$(NAME)"

list: build ## List containers
	sudo ./$(BIN_DIR)/$(BINARY_NAME) list

kill: build ## Send signal to container init process (Usage: make kill NAME=<name> [SIGNAL=TERM|KILL|9])
	sudo ./$(BIN_DIR)/$(BINARY_NAME) kill "$(NAME)" "$(SIGNAL)"

kill-all: build ## Kill all processes in the container cgroup
	sudo ./$(BIN_DIR)/$(BINARY_NAME) kill --all "$(NAME)"

delete: build ## Delete a stopped container
	sudo ./$(BIN_DIR)/$(BINARY_NAME) delete "$(NAME)"

delete-force: build ## Force-delete a running/stuck container
	sudo ./$(BIN_DIR)/$(BINARY_NAME) delete --force "$(NAME)"

clean-container: build ## Best-effort cleanup for one container
	-sudo ./$(BIN_DIR)/$(BINARY_NAME) kill --all "$(NAME)"
	-sudo ./$(BIN_DIR)/$(BINARY_NAME) delete --force "$(NAME)"

lifecycle: create start list ## Create, start, then list container state

##@ Test Commands
test: test-unit ## Run all unit tests

test-unit: check-go-version ## Run unit tests
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
	@rm -rf $(BIN_DIR) coverage.out coverage.html
	@echo "Removed build artifacts"

install: build ## Install binary to GOPATH
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(CMD_DIR)

dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	$(GO) mod download
	$(GO) mod tidy
