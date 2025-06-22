# Makefile for slack-buddy-ai
# 
# Common targets:
#   make build    - Build the binary
#   make test     - Run tests
#   make coverage - Generate test coverage
#   make clean    - Clean build artifacts

# Variables
BINARY_NAME=slack-buddy
BINARY_PATH=./bin/$(BINARY_NAME)
MODULE_NAME=slack-buddy-ai
GO_VERSION=1.24.4

# Build directories
BUILD_DIR=./build
COVERAGE_DIR=$(BUILD_DIR)/coverage
REPORTS_DIR=$(BUILD_DIR)/reports
ARTIFACTS_DIR=$(BUILD_DIR)/artifacts

# Build info
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Linker flags to embed build info and optimize binary size
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Default target
.PHONY: all
all: clean test build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	go build $(LDFLAGS) -o $(BINARY_PATH) .
	@echo "Binary built: $(BINARY_PATH)"

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(ARTIFACTS_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(ARTIFACTS_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(ARTIFACTS_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(ARTIFACTS_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(ARTIFACTS_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Cross-platform binaries built in $(ARTIFACTS_DIR)/"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with race detection
.PHONY: test-race
test-race:
	@echo "Running tests with race detection..."
	go test -race -v ./...

# Generate test coverage
.PHONY: coverage
coverage:
	@echo "Generating test coverage..."
	@mkdir -p $(COVERAGE_DIR)
	go test -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

# Show coverage stats
.PHONY: coverage-stats
coverage-stats:
	@echo "Coverage statistics:"
	go test -cover ./...

# Clean build artifacts and coverage files
.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -rf bin/ $(BUILD_DIR)/
	rm -f $(BINARY_NAME)
	@echo "Cleaned up build artifacts and coverage files"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Vet code
.PHONY: vet
vet:
	@echo "Vetting code..."
	go vet ./...

# Run security scan (requires gosec)
.PHONY: security
security:
	@echo "Running security scan..."
	@mkdir -p $(REPORTS_DIR)
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -fmt=json -out=$(REPORTS_DIR)/security-report.json ./...; \
		gosec ./...; \
	else \
		echo "gosec not installed. Run: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

# Run vulnerability check (requires govulncheck)
.PHONY: vuln-check
vuln-check:
	@echo "Running vulnerability check..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed. Run: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# Run comprehensive security analysis
.PHONY: security-full
security-full: security vuln-check
	@echo "Running module verification..."
	go mod verify
	@echo "Security analysis complete!"

# Install code quality tools
.PHONY: install-quality
install-quality:
	@echo "Installing code quality tools..."
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@echo "Code quality tools installed!"

# Install security tools
.PHONY: install-security
install-security:
	@echo "Installing security tools..."
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "Security tools installed!"

# Install release tools
.PHONY: install-release
install-release:
	@echo "Installing release tools..."
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "Installing goreleaser..."; \
		go install github.com/goreleaser/goreleaser@latest; \
	else \
		echo "goreleaser already installed"; \
	fi
	@echo "Release tools installed!"

# Build release artifacts locally with GoReleaser
.PHONY: release-build
release-build: clean
	@echo "Building release artifacts with GoReleaser..."
	@mkdir -p $(BUILD_DIR)/dist
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "goreleaser not installed. Run: make install-release"; \
		exit 1; \
	fi
	goreleaser build --snapshot --clean --output $(BUILD_DIR)/dist/
	@echo "Release artifacts built in $(BUILD_DIR)/dist/"

# Create full release with checksums
.PHONY: release
release: clean
	@echo "Creating full release with checksums..."
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "goreleaser not installed. Run: make install-release"; \
		exit 1; \
	fi
	goreleaser release --snapshot --clean
	@echo "Release created in dist/ with checksums"
	@echo "Files:"
	@ls -la dist/

# Check GoReleaser configuration
.PHONY: release-check
release-check:
	@echo "Checking GoReleaser configuration..."
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "goreleaser not installed. Run: make install-release"; \
		exit 1; \
	fi
	goreleaser check
	@echo "GoReleaser configuration is valid!"

# Quick development cycle: format, vet, test, build
.PHONY: dev
dev: fmt vet test build

# Quality checks: format and complexity
.PHONY: quality
quality: fmt-check complexity-check

# Full CI-like checks
.PHONY: ci
ci: clean deps fmt vet lint complexity-check security-full test-race coverage build

# Check code formatting
.PHONY: fmt-check
fmt-check:
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Code is not properly formatted. Run 'make fmt' to fix:"; \
		gofmt -l .; \
		exit 1; \
	else \
		echo "Code is properly formatted"; \
	fi

# Check cyclomatic complexity
.PHONY: complexity-check
complexity-check:
	@echo "Checking cyclomatic complexity..."
	@GOCYCLO_PATH=$$(which gocyclo 2>/dev/null || echo "$$(go env GOPATH)/bin/gocyclo"); \
	if [ -x "$$GOCYCLO_PATH" ]; then \
		if $$GOCYCLO_PATH -over 15 . | grep -q .; then \
			echo "Functions with high cyclomatic complexity (>15):"; \
			$$GOCYCLO_PATH -over 15 .; \
			exit 1; \
		else \
			echo "All functions have acceptable complexity"; \
		fi; \
	else \
		echo "gocyclo not installed. Run: make install-quality"; \
		exit 1; \
	fi

# Install the binary to $GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) .

# Uninstall the binary from $GOPATH/bin
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f $(GOPATH)/bin/$(BINARY_NAME)

# Run the binary with help
.PHONY: run
run: build
	$(BINARY_PATH) --help

# Check Go version
.PHONY: version-check
version-check:
	@echo "Required Go version: $(GO_VERSION)"
	@echo "Current Go version: $(shell go version)"

# Show available targets
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary (output: bin/)"
	@echo "  build-all    - Build for multiple platforms (output: build/artifacts/)"
	@echo "  test         - Run tests"
	@echo "  test-race    - Run tests with race detection"
	@echo "  coverage     - Generate test coverage report (output: build/coverage/)"
	@echo "  coverage-stats - Show coverage statistics"
	@echo "  clean        - Clean build artifacts (removes bin/ and build/)"
	@echo "  deps         - Install dependencies"
	@echo "  fmt          - Format code"
	@echo "  fmt-check    - Check if code is properly formatted (CI-friendly)"
	@echo "  complexity-check - Check cyclomatic complexity (requires gocyclo)"
	@echo "  lint         - Lint code (requires golangci-lint)"
	@echo "  vet          - Vet code"
	@echo "  security     - Run security scan (requires gosec)"
	@echo "  vuln-check   - Run vulnerability check (requires govulncheck)"
	@echo "  security-full - Run comprehensive security analysis"
	@echo "  install-quality - Install code quality tools"
	@echo "  install-security - Install security tools"
	@echo "  quality      - Run quality checks (fmt-check, complexity-check)"
	@echo "  dev          - Quick development cycle (fmt, vet, test, build)"
	@echo "  ci           - Full CI-like checks"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  uninstall    - Remove binary from GOPATH/bin"
	@echo "  run          - Build and run with --help"
	@echo "  version-check - Check Go version"
	@echo "  help         - Show this help"
	@echo ""
	@echo "Release targets:"
	@echo "  install-release - Install GoReleaser and release tools"
	@echo "  release-build   - Build release artifacts (no checksums)"
	@echo "  release         - Create full release with checksums"
	@echo "  release-check   - Validate GoReleaser configuration"