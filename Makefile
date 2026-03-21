BINARY_NAME := omniclean
BUILD_DIR   := bin
MAIN_PKG    := ./cmd/omniclean

.PHONY: build run test lint fmt clean install-tools

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

## build: Compile the binary to bin/omniclean
build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PKG)

## run: Run the application directly (no build)
run:
	go run $(MAIN_PKG)

## run-dry: Run with --dry-run flag
run-dry:
	go run $(MAIN_PKG) --dry-run

## test: Run all tests with race detector and coverage
test:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

## lint: Run golangci-lint
lint:
	golangci-lint run

## fmt: Format all Go source files
fmt:
	gofmt -w .
	goimports -w . 2>/dev/null || true

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR) coverage.out

## install-tools: Install required development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/goreleaser/goreleaser@latest
	go install golang.org/x/tools/cmd/goimports@latest

## help: Show this help
help:
	@grep -E '^##' Makefile | sed 's/## //'
