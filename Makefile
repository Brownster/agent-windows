# Windows Agent Collector Makefile

.PHONY: build test clean lint help

# Build variables
BINARY_NAME = windows-agent-collector
BINARY_WINDOWS = $(BINARY_NAME).exe
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT = $(shell git rev-parse HEAD)

# Go build flags
LDFLAGS = -ldflags "-X github.com/prometheus/common/version.Version=$(VERSION) \
                   -X github.com/prometheus/common/version.Revision=$(GIT_COMMIT) \
                   -X github.com/prometheus/common/version.BuildDate=$(BUILD_DATE) \
                   -X github.com/prometheus/common/version.Branch=$(shell git rev-parse --abbrev-ref HEAD)"

# Default target
all: build

## build: Build the Windows agent collector
build:
	@echo "Building $(BINARY_WINDOWS)..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_WINDOWS) ./cmd/agent

## build-arm64: Build for Windows ARM64
build-arm64:
	@echo "Building $(BINARY_NAME)-arm64.exe..."
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-arm64.exe ./cmd/agent

## test: Run tests
test:
	@echo "Running tests..."
	GOOS=windows GOARCH=amd64 go test -v ./...

## test-compile: Compile tests without running them
test-compile:
	@echo "Compiling tests..."
	GOOS=windows GOARCH=amd64 go test -c ./cmd/agent
	GOOS=windows GOARCH=amd64 go test -c ./internal/collector/cpu
	GOOS=windows GOARCH=amd64 go test -c ./internal/collector/memory
	GOOS=windows GOARCH=amd64 go test -c ./internal/collector/net
	GOOS=windows GOARCH=amd64 go test -c ./internal/collector/pagefile

## clean: Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_WINDOWS)
	rm -f $(BINARY_NAME)-arm64.exe
	rm -f *.test.exe
	go clean

## lint: Run linters
lint:
	@echo "Running linters..."
	go fmt ./...
	go vet ./...

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

## verify: Verify dependencies and build
verify: deps lint test-compile
	@echo "Verification complete"

## package: Build for multiple architectures
package: build build-arm64
	@echo "Built packages:"
	@ls -la *.exe

## help: Show this help
help:
	@echo "Windows Agent Collector Build System"
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | sort

# Generate version info
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"