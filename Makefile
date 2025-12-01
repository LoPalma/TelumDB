# TelumDB Makefile

.PHONY: help build test clean deps fmt lint vet coverage bench docker docker-build docker-run install uninstall

# Variables
BINARY_NAME=telumdb
CLI_NAME=telumdb-cli
MAIN_PATH=./cmd/telumdb
CLI_PATH=./cmd/telumdb-cli
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
DOCKER_IMAGE=telumdb/telumdb
DOCKER_TAG=latest

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
deps: ## Install dependencies
	go mod download
	go mod tidy

fmt: ## Format code
	go fmt ./...
	gofmt -s -w .

lint: ## Run linter
	golangci-lint run

vet: ## Run go vet
	go vet ./...

# Building
build: ## Build the binary
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(CLI_NAME) $(CLI_PATH)

build-all: ## Build for all platforms
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Testing
test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-race: ## Run tests with race detector
	go test -race -v ./...

bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

test-integration: ## Run integration tests
	go test -v -tags=integration ./test/integration/...

# Quality
quality: fmt lint vet test ## Run all quality checks

# Docker
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run Docker container
	docker run -p 5432:5432 $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-push: docker-build ## Push Docker image
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

# Installation
install: build ## Install binary
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	cp $(BUILD_DIR)/$(CLI_NAME) $(GOPATH)/bin/

uninstall: ## Remove binary
	rm -f $(GOPATH)/bin/$(BINARY_NAME)
	rm -f $(GOPATH)/bin/$(CLI_NAME)

# Documentation
docs: ## Generate documentation
	godoc -http=:6060

docs-api: ## Generate API documentation
	swag init -g cmd/telumdb/main.go

# Clean
clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean -cache

# Development helpers
run: build ## Run the server
	./$(BUILD_DIR)/$(BINARY_NAME)

dev: ## Run in development mode with hot reload
	air -c .air.toml

# Release
release: clean test build-all ## Create a release
	@echo "Release built in $(BUILD_DIR)"

# Security
security: ## Run security checks
	gosec ./...

# Performance
profile: ## Run with profiling
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-profile $(MAIN_PATH)
	./$(BUILD_DIR)/$(BINARY_NAME)-profile -cpuprofile=cpu.prof -memprofile=mem.prof

# Database management
db-init: ## Initialize database
	./$(BUILD_DIR)/$(CLI_NAME) init

db-migrate: ## Run database migrations
	./$(BUILD_DIR)/$(CLI_NAME) migrate up

db-reset: ## Reset database
	./$(BUILD_DIR)/$(CLI_NAME) reset

# Examples
examples: ## Run examples
	cd examples && go run main.go

# Python bindings
python-build: ## Build Python bindings
	cd api/python && python setup.py build_ext --inplace

python-install: ## Install Python bindings
	cd api/python && pip install -e .

python-test: ## Test Python bindings
	cd api/python && python -m pytest tests/

# C library
c-build: ## Build C library
	cd api/c && make

c-test: ## Test C library
	cd api/c && make test

# Java module
java-build: ## Build Java module
	cd api/java && mvn compile

java-test: ## Test Java module
	cd api/java && mvn test

# All language bindings
bindings-all: python-build c-build java-build ## Build all language bindings

# CI/CD helpers
ci: quality security test-integration ## Run CI pipeline

# Local development setup
setup: deps ## Set up development environment
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "Development environment setup complete!"