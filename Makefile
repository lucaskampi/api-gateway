.PHONY: build test test-unit test-integration lint fmt bench coverage clean help

BINARY_NAME=gateway
BUILD_DIR=bin
MAIN_PATH=./cmd/gateway

help:
	@echo "Available targets:"
	@echo "  build           - Build the gateway binary"
	@echo "  test            - Run all tests"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration - Run integration tests"
	@echo "  lint            - Run linters"
	@echo "  fmt             - Format code"
	@echo "  bench           - Run benchmarks"
	@echo "  coverage        - Run tests with coverage"
	@echo "  clean           - Remove built binaries"

build:
	@echo "Building gateway..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

test: test-unit test-integration

test-unit:
	@echo "Running unit tests..."
	@go test -v -race ./internal/adapter/...

test-integration:
	@echo "Running integration tests..."
	@go test -v -race ./tests/integration/...

lint:
	@echo "Running linters..."
	@golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	@gofmt -w .
	@goimports -w .

bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./internal/adapter/...

coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
