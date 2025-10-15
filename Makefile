# Makefile for CQVX - Crypto Quant Venue Exchange

.PHONY: help
help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy

.PHONY: test
test: ## Run all unit tests
	go test ./... -v

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-short
test-short: ## Run short tests only
	go test ./... -short -v

.PHONY: lint
lint: ## Run linter (requires golangci-lint)
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: build
build: ## Build all packages
	go build ./...

.PHONY: clean
clean: ## Clean build artifacts and coverage files
	go clean
	rm -f coverage.out coverage.html

.PHONY: check
check: fmt vet test ## Run format, vet, and test

.PHONY: ci
ci: tidy fmt vet test ## Run all CI checks

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "Installing golangci-lint..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed successfully"

.PHONY: mod-download
mod-download: ## Download Go module dependencies
	go mod download

.PHONY: mod-verify
mod-verify: ## Verify Go module dependencies
	go mod verify

.PHONY: generate
generate: ## Run go generate
	go generate ./...

.DEFAULT_GOAL := help
