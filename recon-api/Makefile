.PHONY: help build run test clean docker-build docker-run fmt lint

# Variables
APP_NAME=recon-api
DOCKER_IMAGE=recon-api:latest
GO=go
GOFLAGS=-v

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	$(GO) build $(GOFLAGS) -o bin/$(APP_NAME) ./cmd/main.go

run: ## Run the application locally
	$(GO) run ./cmd/main.go

test: ## Run tests
	$(GO) test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	$(GO) tool cover -html=coverage.out

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out

fmt: ## Format code
	$(GO) fmt ./...
	gofmt -s -w .

lint: ## Run linter
	golangci-lint run ./...

deps: ## Download dependencies
	$(GO) mod download
	$(GO) mod tidy

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	docker run --rm -p 3000:3000 --env-file .env $(DOCKER_IMAGE)

dev: ## Run in development mode with auto-reload (requires air)
	@command -v air > /dev/null 2>&1 || { \
		echo "Air is not installed. Installing..."; \
		go install github.com/air-verse/air@latest; \
	}
	air

install-tools: ## Install development tools
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
