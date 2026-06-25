BINARY_NAME := omniclean
BUILD_DIR   := bin
MAIN_PKG    := ./cmd/omniclean

.PHONY: build run test lint fmt clean install-tools install vet vuln tidy-check sbom release-snapshot

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

## lint: golangci-lint v2
lint:
	golangci-lint run

## fmt: Format all Go source files
fmt:
	gofmt -w .
	goimports -w . 2>/dev/null || true

## install: Build and install binary to /usr/local/bin
install: build
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR) coverage.out

## vet: go vet
vet:
	go vet ./...

## vuln: govulncheck
vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## tidy-check: fail if go.mod/go.sum drift
tidy-check:
	go mod tidy
	git diff --exit-code -- go.mod go.sum

## sbom: generate SBOM for current build (needs syft)
sbom:
	syft . -o spdx-json=omniclean.sbom.json

## release-snapshot: dry release (no publish)
release-snapshot:
	goreleaser release --snapshot --clean --skip=publish,sign

## install-tools: Install required development tools
install-tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
	go install github.com/goreleaser/goreleaser/v2@v2.16.0
	go install golang.org/x/tools/cmd/goimports@latest

## help: Show this help
help:
	@grep -E '^##' Makefile | sed 's/## //'
