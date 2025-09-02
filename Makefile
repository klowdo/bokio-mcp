# Bokio MCP Server Makefile
# Development automation for the Bokio MCP server project
# Supports full development lifecycle from schema updates to releases
.PHONY: help update-schema generate-types build test lint dev clean deps security release-dry nix-build pre-commit pre-commit-install pre-commit-run pre-commit-update install-tools run watch tag release info status profile benchmark format check-deps

# Set shell and enable error checking
SHELL := $(shell which bash)
.SHELLFLAGS := -euo pipefail -c
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

# Default target - shows available commands
help: ## Show this help message with available targets
	@echo 'Bokio MCP Server - Development Automation'
	@echo '========================================='
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Main Development Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(update-schema|generate-types|build|test|lint|dev|clean|deps|security|pre-commit)"
	@echo ''
	@echo 'Release and Advanced Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[33m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(release-dry|nix-build|install-tools|run|watch|tag|release)"
	@echo ''
	@echo 'Variables:'
	@echo "  VERSION=$(VERSION)"
	@echo "  BINARY_NAME=$(BINARY_NAME)"
	@echo "  GO_VERSION=$$(go version 2>/dev/null | awk '{print $$3}' || echo 'not installed')"

# Project variables
BINARY_NAME := bokio-mcp
PACKAGE := github.com/klowdo/bokio-mcp
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILD_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.BuildCommit=$(BUILD_COMMIT) -s -w"

# Directory structure
SCHEMAS_DIR := schemas
GENERATED_DIR := bokio/generated
BIN_DIR := bin
DIST_DIR := dist
TOOLS_DIR := tools

# API URLs for schema updates
BOKIO_API_BASE := https://raw.githubusercontent.com/bokio/bokio-api/v1/api-specification
COMPANY_API_URL := $(BOKIO_API_BASE)/company-api.yaml
GENERAL_API_URL := $(BOKIO_API_BASE)/general-api.yaml

# Tool versions - keep up to date with latest stable releases
OAPI_CODEGEN_VERSION := v1.16.3
GOLANGCI_LINT_VERSION := v1.61.0
GORELEASER_VERSION := v2.4.0
GOVULNCHECK_VERSION := latest
GOSEC_VERSION := latest

# Colors for pretty output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_BLUE := \033[34m
COLOR_YELLOW := \033[33m
COLOR_RED := \033[31m

# Helper function to print status messages
define print_status
	@printf "$(COLOR_BOLD)$(COLOR_BLUE)▶$(COLOR_RESET) $(COLOR_BOLD)%s$(COLOR_RESET)\n" $(1)
endef

define print_success
	@printf "$(COLOR_BOLD)$(COLOR_GREEN)✓$(COLOR_RESET) $(COLOR_BOLD)%s$(COLOR_RESET)\n" $(1)
endef

define print_warning
	@printf "$(COLOR_BOLD)$(COLOR_YELLOW)⚠$(COLOR_RESET) $(COLOR_BOLD)%s$(COLOR_RESET)\n" $(1)
endef

define print_error
	@printf "$(COLOR_BOLD)$(COLOR_RED)✗$(COLOR_RESET) $(COLOR_BOLD)%s$(COLOR_RESET)\n" $(1)
endef

# Check if command exists
define check_command
	@command -v $(1) >/dev/null 2>&1 || { \
		$(call print_error,"$(1) is not installed. Run 'make install-tools' first."); \
		exit 1; \
	}
endef

# =============================================================================
# Tool Installation
# =============================================================================

install-tools: ## Download Go tools (not needed - using go run instead)
	$(call print_warning,"This target is deprecated. Tools are now run directly with 'go run'")
	$(call print_success,"No installation needed - tools run directly with go run")

# =============================================================================
# Schema Management
# =============================================================================

update-schema: ## Download latest OpenAPI specs from Bokio GitHub
	$(call print_status,"Downloading Bokio API specifications...")
	@mkdir -p $(SCHEMAS_DIR)
	@if ! command -v curl >/dev/null 2>&1; then \
		$(call print_error,"curl is not installed. Please install curl first."); \
		exit 1; \
	fi
	@$(call print_status,"Downloading company-api.yaml...")
	@if ! curl -sSL --fail "$(COMPANY_API_URL)" -o "$(SCHEMAS_DIR)/company-api.yaml"; then \
		$(call print_error,"Failed to download company-api.yaml"); \
		exit 1; \
	fi
	@$(call print_status,"Downloading general-api.yaml...")
	@if ! curl -sSL --fail "$(GENERAL_API_URL)" -o "$(SCHEMAS_DIR)/general-api.yaml"; then \
		$(call print_error,"Failed to download general-api.yaml"); \
		exit 1; \
	fi
	@$(call print_status,"Validating downloaded schemas...")
	@if [ ! -s "$(SCHEMAS_DIR)/company-api.yaml" ] || [ ! -s "$(SCHEMAS_DIR)/general-api.yaml" ]; then \
		$(call print_error,"One or more schema files are empty"); \
		exit 1; \
	fi
	$(call print_success,"API specifications downloaded to $(SCHEMAS_DIR)/")

generate-types: ## Generate Go types from OpenAPI specifications using go generate
	$(call print_status,"Generating Go types from OpenAPI specs...")
	@mkdir -p $(GENERATED_DIR)/company $(GENERATED_DIR)/general
	@$(call print_status,"Running go generate...")
	@if ! go generate ./$(GENERATED_DIR)/...; then \
		$(call print_error,"Failed to generate types"); \
		exit 1; \
	fi
	$(call print_success,"Generated types and clients in $(GENERATED_DIR)/")

# =============================================================================
# Build Targets
# =============================================================================

build: ## Build the MCP server binary (run go generate first if needed)
	$(call print_status,"Building $(BINARY_NAME) $(VERSION)...")
	@mkdir -p $(BIN_DIR)
	@$(call print_status,"Checking Go module dependencies...")
	@go mod verify
	@$(call print_status,"Compiling binary...")
	@if ! go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) .; then \
		$(call print_error,"Build failed"); \
		exit 1; \
	fi
	@$(call print_status,"Verifying binary...")
	@if [ ! -x "$(BIN_DIR)/$(BINARY_NAME)" ]; then \
		$(call print_error,"Binary is not executable"); \
		exit 1; \
	fi
	@file_size=$$(du -h $(BIN_DIR)/$(BINARY_NAME) | cut -f1); \
	$(call print_success,"Built $(BINARY_NAME) ($$file_size) in $(BIN_DIR)/")

build-only: ## Build the MCP server binary without generating types (for Nix)
	$(call print_status,"Building $(BINARY_NAME) $(VERSION) (build-only mode)...")
	@mkdir -p $(BIN_DIR)
	@$(call print_status,"Checking Go module dependencies...")
	@go mod verify
	@$(call print_status,"Compiling binary...")
	@if ! go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) .; then \
		$(call print_error,"Build failed"); \
		exit 1; \
	fi
	@$(call print_status,"Verifying binary...")
	@if [ ! -x "$(BIN_DIR)/$(BINARY_NAME)" ]; then \
		$(call print_error,"Binary is not executable"); \
		exit 1; \
	fi
	@file_size=$$(du -h $(BIN_DIR)/$(BINARY_NAME) | cut -f1); \
	$(call print_success,"Built $(BINARY_NAME) ($$file_size) in $(BIN_DIR)/")

# =============================================================================
# Testing and Quality
# =============================================================================

test: ## Run all tests with coverage reporting
	$(call print_status,"Running tests with coverage...")
	@$(call print_status,"Checking for test files...")
	@if [ -z "$$(find . -name '*_test.go' -not -path './vendor/*')" ]; then \
		printf "$(COLOR_BOLD)$(COLOR_YELLOW)⚠$(COLOR_RESET) $(COLOR_BOLD)%s$(COLOR_RESET)\n" "No test files found"; \
	else \
		printf "$(COLOR_BOLD)$(COLOR_BLUE)▶$(COLOR_RESET) $(COLOR_BOLD)%s$(COLOR_RESET)\n" "Running tests with race detection..."; \
		if ! go test -v -race -coverprofile=coverage.out ./...; then \
			printf "$(COLOR_BOLD)$(COLOR_RED)✗$(COLOR_RESET) $(COLOR_BOLD)%s$(COLOR_RESET)\n" "Tests failed"; \
			exit 1; \
		fi; \
		printf "$(COLOR_BOLD)$(COLOR_BLUE)▶$(COLOR_RESET) $(COLOR_BOLD)%s$(COLOR_RESET)\n" "Generating coverage report..."; \
		go tool cover -html=coverage.out -o coverage.html; \
		coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}'); \
		printf "$(COLOR_BOLD)$(COLOR_GREEN)✓$(COLOR_RESET) $(COLOR_BOLD)%s$(COLOR_RESET)\n" "Tests passed with $$coverage coverage"; \
		printf "$(COLOR_BOLD)$(COLOR_BLUE)▶$(COLOR_RESET) $(COLOR_BOLD)%s$(COLOR_RESET)\n" "Coverage report: coverage.html"; \
	fi

lint: ## Run golangci-lint for code quality analysis using go tool
	$(call print_status,"Running code quality checks...")
	@$(call print_status,"Checking golangci-lint configuration...")
	@if [ ! -f ".golangci.yml" ] && [ ! -f ".golangci.yaml" ]; then \
		printf "$(COLOR_BOLD)$(COLOR_YELLOW)⚠$(COLOR_RESET) $(COLOR_BOLD)%s$(COLOR_RESET)\n" "No golangci-lint config found, using defaults"; \
	fi
	@if ! go tool golangci-lint run ./...; then \
		$(call print_error,"Linting failed"); \
		exit 1; \
	fi
	$(call print_success,"Code quality checks passed")

# =============================================================================
# Development
# =============================================================================

dev: ## Run server in development mode with enhanced output
	$(call print_status,"Starting development server...")
	@$(call print_status,"Development mode with hot reload enabled")
	@$(call print_warning,"Press Ctrl+C to stop the server")
	@if ! go run $(LDFLAGS) . --dev; then \
		$(call print_error,"Development server failed to start"); \
		exit 1; \
	fi

watch: ## Watch for file changes and rebuild automatically
	$(call print_status,"Starting file watcher...")
	@if ! command -v inotifywait >/dev/null 2>&1; then \
		$(call print_error,"inotifywait not found. Install inotify-tools first."); \
		exit 1; \
	fi
	@$(call print_warning,"Watching for changes... Press Ctrl+C to stop")
	@while true; do \
		inotifywait -e modify,create,delete -r . --exclude '\.git|$(BIN_DIR)|$(DIST_DIR)|\..*\.swp|coverage\.' 2>/dev/null || break; \
		$(call print_status,"Change detected, rebuilding..."); \
		make build || $(call print_error,"Build failed, waiting for next change..."); \
	done

# =============================================================================
# Maintenance
# =============================================================================

clean: ## Clean build artifacts and generated files
	$(call print_status,"Cleaning build artifacts...")
	@rm -rf $(BIN_DIR)/
	@rm -rf $(DIST_DIR)/
	@rm -rf $(GENERATED_DIR)/
	@rm -f coverage.out coverage.html
	@rm -f *.prof *.test
	@go clean -cache -testcache -modcache 2>/dev/null || true
	$(call print_success,"Cleanup completed")

deps: ## Update and verify Go dependencies
	$(call print_status,"Managing Go dependencies...")
	@$(call print_status,"Tidying module dependencies...")
	@go mod tidy
	@$(call print_status,"Downloading dependencies...")
	@go mod download
	@$(call print_status,"Verifying dependencies...")
	@go mod verify
	@$(call print_status,"Checking for available updates...")
	@go list -u -m all | grep -v "$(PACKAGE)" || true
	$(call print_success,"Dependencies updated and verified")

# =============================================================================
# Security
# =============================================================================

security: ## Run comprehensive security scans using go tool
	$(call print_status,"Running security scans...")
	@$(call print_status,"Checking for known vulnerabilities...")
	@if ! go tool govulncheck ./...; then \
		$(call print_error,"Vulnerability check failed"); \
		exit 1; \
	fi
	@$(call print_status,"Running static security analysis...")
	@if ! go tool gosec -quiet -exclude-dir=bokio/generated ./...; then \
		$(call print_error,"Security analysis found issues"); \
		exit 1; \
	fi
	$(call print_success,"Security scans completed successfully")

# =============================================================================
# Release Management
# =============================================================================

release-dry: ## Test GoReleaser configuration without publishing using go tool
	$(call print_status,"Testing release configuration...")
	@if [ ! -f ".goreleaser.yml" ] && [ ! -f ".goreleaser.yaml" ]; then \
		$(call print_error,"No GoReleaser configuration found"); \
		exit 1; \
	fi
	@if ! go tool goreleaser check; then \
		$(call print_error,"GoReleaser configuration is invalid"); \
		exit 1; \
	fi
	@$(call print_status,"Running dry-run release...")
	@if ! go tool goreleaser release --snapshot --clean --skip=publish; then \
		$(call print_error,"Release dry-run failed"); \
		exit 1; \
	fi
	$(call print_success,"Release configuration is valid")

nix-build: ## Build using Nix flake (if available)
	@if [ ! -f "flake.nix" ]; then \
		$(call print_warning,"No flake.nix found, skipping Nix build"); \
		exit 0; \
	fi
	$(call print_status,"Building with Nix...")
	@if ! command -v nix >/dev/null 2>&1; then \
		$(call print_error,"Nix is not installed"); \
		exit 1; \
	fi
	@if ! nix build .; then \
		$(call print_error,"Nix build failed"); \
		exit 1; \
	fi
	$(call print_success,"Nix build completed")

# =============================================================================
# Pre-commit Pipeline
# =============================================================================

pre-commit: deps lint test security ## Run comprehensive pre-commit checks
	$(call print_status,"Running pre-commit pipeline...")
	@$(call print_status,"Checking Git status...")
	@if [ -n "$$(git status --porcelain 2>/dev/null)" ]; then \
		$(call print_warning,"Working directory has uncommitted changes"); \
	fi
	@$(call print_status,"Verifying build...")
	@make build >/dev/null
	$(call print_success,"All pre-commit checks passed! ✨")

pre-commit-install: ## Install pre-commit git hooks
	$(call print_status,"Installing pre-commit hooks...")
	@if ! command -v pre-commit >/dev/null 2>&1; then \
		$(call print_error,"pre-commit is not installed. Use 'nix develop' or install manually."); \
		exit 1; \
	fi
	@if ! pre-commit install; then \
		$(call print_error,"Failed to install pre-commit hooks"); \
		exit 1; \
	fi
	$(call print_success,"Pre-commit hooks installed successfully")

pre-commit-run: ## Run pre-commit hooks on all files
	$(call print_status,"Running pre-commit hooks on all files...")
	@if ! command -v pre-commit >/dev/null 2>&1; then \
		$(call print_error,"pre-commit is not installed. Use 'nix develop' or install manually."); \
		exit 1; \
	fi
	@if ! pre-commit run --all-files; then \
		$(call print_error,"Pre-commit hooks failed"); \
		exit 1; \
	fi
	$(call print_success,"All pre-commit hooks passed")

pre-commit-update: ## Update pre-commit hooks to latest versions
	$(call print_status,"Updating pre-commit hooks...")
	@if ! command -v pre-commit >/dev/null 2>&1; then \
		$(call print_error,"pre-commit is not installed. Use 'nix develop' or install manually."); \
		exit 1; \
	fi
	@if ! pre-commit autoupdate; then \
		$(call print_error,"Failed to update pre-commit hooks"); \
		exit 1; \
	fi
	$(call print_success,"Pre-commit hooks updated successfully")

# =============================================================================
# Development Shortcuts
# =============================================================================

run: build ## Build and run the server
	$(call print_status,"Running $(BINARY_NAME)...")
	@./$(BIN_DIR)/$(BINARY_NAME)

# =============================================================================
# Release Shortcuts
# =============================================================================

tag: ## Create and push a new tag (usage: make tag VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "dev" ]; then \
		$(call print_error,"Please specify a version: make tag VERSION=v1.0.0"); \
		exit 1; \
	fi
	$(call print_status,"Creating tag $(VERSION)...")
	@if git tag -l | grep -q "^$(VERSION)$$"; then \
		$(call print_error,"Tag $(VERSION) already exists"); \
		exit 1; \
	fi
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	$(call print_success,"Tag $(VERSION) created and pushed")

release: ## Create a new release (requires tag to be pushed) using go tool
	$(call print_status,"Creating release...")
	@if [ ! -f ".goreleaser.yml" ] && [ ! -f ".goreleaser.yaml" ]; then \
		$(call print_error,"No GoReleaser configuration found"); \
		exit 1; \
	fi
	@if ! go tool goreleaser release --clean; then \
		$(call print_error,"Release failed"); \
		exit 1; \
	fi
	$(call print_success,"Release created successfully")

# =============================================================================
# Information and Status
# =============================================================================

info: ## Show project information and current status
	@echo ""
	@echo "$(COLOR_BOLD)Bokio MCP Server - Project Information$(COLOR_RESET)"
	@echo "======================================"
	@echo ""
	@echo "$(COLOR_BOLD)Project Details:$(COLOR_RESET)"
	@echo "  Name:           $(BINARY_NAME)"
	@echo "  Package:        $(PACKAGE)"
	@echo "  Version:        $(VERSION)"
	@echo "  Build Time:     $(BUILD_TIME)"
	@echo "  Build Commit:   $(BUILD_COMMIT)"
	@echo ""
	@echo "$(COLOR_BOLD)Directories:$(COLOR_RESET)"
	@echo "  Schemas:        $(SCHEMAS_DIR)/"
	@echo "  Generated:      $(GENERATED_DIR)/"
	@echo "  Binary:         $(BIN_DIR)/"
	@echo "  Distribution:   $(DIST_DIR)/"
	@echo ""
	@echo "$(COLOR_BOLD)Go Environment:$(COLOR_RESET)"
	@go version 2>/dev/null || echo "  Go: not installed"
	@echo "  GOPATH:         $${GOPATH:-not set}"
	@echo "  GO111MODULE:    $${GO111MODULE:-auto}"
	@echo ""
	@echo "$(COLOR_BOLD)Tool Versions:$(COLOR_RESET)"
	@echo "  oapi-codegen:   $(OAPI_CODEGEN_VERSION)"
	@echo "  golangci-lint:  $(GOLANGCI_LINT_VERSION)"
	@echo "  goreleaser:     $(GORELEASER_VERSION)"
	@echo ""
	@echo "$(COLOR_BOLD)File Status:$(COLOR_RESET)"
	@if [ -f "$(SCHEMAS_DIR)/company-api.yaml" ]; then \
		echo "  Company API:    ✓ present"; \
	else \
		echo "  Company API:    ✗ missing"; \
	fi
	@if [ -f "$(SCHEMAS_DIR)/general-api.yaml" ]; then \
		echo "  General API:    ✓ present"; \
	else \
		echo "  General API:    ✗ missing"; \
	fi
	@if [ -d "$(GENERATED_DIR)" ] && [ -n "$$(ls -A $(GENERATED_DIR) 2>/dev/null)" ]; then \
		echo "  Generated Code: ✓ present"; \
	else \
		echo "  Generated Code: ✗ missing"; \
	fi
	@if [ -f "$(BIN_DIR)/$(BINARY_NAME)" ]; then \
		echo "  Binary:         ✓ built"; \
	else \
		echo "  Binary:         ✗ not built"; \
	fi
	@echo ""

status: info ## Alias for info target

# =============================================================================
# Advanced Development Targets
# =============================================================================

profile: build ## Build with profiling enabled and run basic profiling
	$(call print_status,"Building with profiling enabled...")
	@go build -o $(BIN_DIR)/$(BINARY_NAME)-profile .
	$(call print_status,"Running CPU profiling (30 seconds)...")
	@timeout 30s $(BIN_DIR)/$(BINARY_NAME)-profile -cpuprofile=cpu.prof || true
	$(call print_status,"Analyzing profile data...")
	@go tool pprof -text cpu.prof | head -20
	$(call print_success,"Profile data saved to cpu.prof")

benchmark: ## Run benchmarks (if any exist)
	$(call print_status,"Running benchmarks...")
	@if [ -z "$$(find . -name '*_test.go' -exec grep -l 'func Benchmark' {} \;)" ]; then \
		$(call print_warning,"No benchmark tests found"); \
	else \
		go test -bench=. -benchmem ./...; \
	fi

format: ## Format all Go code using gofmt and goimports
	$(call print_status,"Formatting Go code...")
	@find . -name '*.go' -not -path './vendor/*' -not -path './$(GENERATED_DIR)/*' | xargs gofmt -w
	@if command -v goimports >/dev/null 2>&1; then \
		find . -name '*.go' -not -path './vendor/*' -not -path './$(GENERATED_DIR)/*' | xargs goimports -w; \
	else \
		printf "$(COLOR_BOLD)$(COLOR_YELLOW)⚠$(COLOR_RESET) $(COLOR_BOLD)%s$(COLOR_RESET)\n" "goimports not found, install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi
	$(call print_success,"Code formatting completed")

# =============================================================================
# Utility Functions
# =============================================================================

.PHONY: check-deps
check-deps: ## Check if all required dependencies are available
	$(call print_status,"Checking dependencies...")
	@$(call print_status,"Checking Go installation...")
	@go version >/dev/null 2>&1 || { $(call print_error,"Go is not installed"); exit 1; }
	@$(call print_status,"Checking required tools...")
	@for tool in curl git make; do \
		command -v $$tool >/dev/null 2>&1 || { \
			$(call print_error,"$$tool is not installed"); \
			exit 1; \
		}; \
	done
	@$(call print_status,"Checking optional tools...")
	@for tool in oapi-codegen golangci-lint goreleaser govulncheck gosec; do \
		if command -v $$tool >/dev/null 2>&1; then \
			echo "  ✓ $$tool"; \
		else \
			echo "  ✗ $$tool (install with 'make install-tools')"; \
		fi; \
	done
	$(call print_success,"Dependency check completed")

# Make sure we don't accidentally run targets in parallel that shouldn't be
.NOTPARALLEL: update-schema generate-types build

# Set default goal
.DEFAULT_GOAL := help
