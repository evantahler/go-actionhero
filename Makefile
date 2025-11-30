.PHONY: help build clean test test-coverage lint fmt vet install dev

# Default target
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME=actionhero
BUILD_DIR=.

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build the project
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/actionhero
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
clean: ## Remove build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f coverage.out
	@echo "Clean complete"

# Run tests
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter (requires golangci-lint)
lint: ## Run golangci-lint
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it from https://golangci-lint.run/"; \
		exit 1; \
	fi

# Format code
fmt: ## Format Go code
	@echo "Formatting code..."
	$(GOFMT) ./...

# Run go vet
vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

# Install dependencies
install: ## Install dependencies
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Run in development mode
dev: build ## Build and run the server
	@echo "Starting development server..."
	./$(BINARY_NAME) start

# Run all checks (fmt, vet, lint, test)
check: fmt vet lint test ## Run all checks (format, vet, lint, test)

# Display help
help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
