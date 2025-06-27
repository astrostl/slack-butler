# Makefile for slack-butler
# 
# Main workflows:
#   make dev         - Quick development cycle (format + vet + test + build)
#   make quality     - Complete quality validation (security + format + vet + lint + complexity)
#   make maintenance - Monthly maintenance (deps-update + quality + test)
#   make ci          - Full CI pipeline (clean + deps + quality + coverage + build)

# Variables
BINARY_NAME=slack-butler
BINARY_PATH=./bin/$(BINARY_NAME)
GO_VERSION=1.24.4

# Build directories
BUILD_DIR=./build
COVERAGE_DIR=$(BUILD_DIR)/coverage
REPORTS_DIR=$(BUILD_DIR)/reports

# Build info
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Linker flags to embed build info and optimize binary size
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Default target
.PHONY: all
all: dev

# =============================================================================
# Core Build Targets
# =============================================================================

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	go build $(LDFLAGS) -o $(BINARY_PATH) .
	@echo "Binary built: $(BINARY_PATH)"

# Run tests (with race detection)
.PHONY: test
test:
	@echo "Running tests with race detection..."
	go test -race -v ./...

# Generate test coverage
.PHONY: coverage
coverage:
	@echo "Generating test coverage..."
	@mkdir -p $(COVERAGE_DIR)
	go test -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report: $(COVERAGE_DIR)/coverage.html"

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

# Install development tools (versions pinned in tools.go/go.mod)
.PHONY: install-tools
install-tools:
	@echo "Installing development tools from go.mod versions..."
	@echo "Tools: golangci-lint, gocyclo, gosec, govulncheck"
	@go list -f '{{range .Imports}}{{.}} {{end}}' ./tools.go | xargs go install
	@echo "✅ All development tools installed successfully!"

# =============================================================================
# Helper Functions (Internal Use)
# =============================================================================

define run-fmt
	@gofmt -s -w .
endef

define run-fmt-check
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		echo "❌ Code formatting issues found. Run 'make dev' to fix."; \
		gofmt -s -l .; \
		exit 1; \
	fi
endef

define run-vet
	@go vet ./...
endef

define run-lint
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: make install-tools"; \
		exit 1; \
	fi
endef

define run-lint-maintenance
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --config .golangci-override.yml; \
	else \
		echo "golangci-lint not installed. Run: make install-tools"; \
		exit 1; \
	fi
endef

define run-complexity-check
	@if command -v gocyclo >/dev/null 2>&1; then \
		if gocyclo -over 15 . | grep -q .; then \
			echo "❌ High complexity functions found:"; \
			gocyclo -over 15 .; \
			exit 1; \
		fi; \
	else \
		echo "gocyclo not installed. Run: make install-tools"; \
		exit 1; \
	fi
endef

define run-gosec
	@echo "Running static security analysis..."
	@mkdir -p $(REPORTS_DIR)
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -fmt=json -out=$(REPORTS_DIR)/security-report.json ./...; \
		gosec ./...; \
	else \
		echo "gosec not installed. Run: make install-tools"; \
		exit 1; \
	fi
endef

define run-vuln-check
	@echo "Checking for known vulnerabilities..."
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...
endef

define run-mod-verify
	@echo "Verifying module integrity..."
	@go mod verify
endef

define run-deps-update
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "✅ Dependencies updated. Run 'make test' to verify compatibility."
endef

define run-deps-audit
	@echo "Auditing dependencies for security vulnerabilities..."
	$(call run-vuln-check)
	$(call run-mod-verify)
	@echo "✅ Dependency audit completed!"
endef

define run-security
	$(call run-gosec)
	$(call run-vuln-check)
	$(call run-mod-verify)
endef

# =============================================================================
# Main Workflow Suites (Recommended)
# =============================================================================

# Quick development cycle (format + vet + test + build)
.PHONY: dev
dev:
	$(call run-fmt)
	$(call run-vet)
	@$(MAKE) test
	@$(MAKE) build
	@echo "✅ Development cycle complete!"

# Complete quality validation (security + format + vet + lint + complexity)
.PHONY: quality
quality:
	$(call run-security)
	$(call run-fmt-check)
	$(call run-vet)
	$(call run-lint)
	$(call run-complexity-check)
	@echo "✅ Quality checks completed!"

# Monthly maintenance workflow (update deps + run essential checks + test)
.PHONY: maintenance
maintenance:
	$(call run-deps-update)
	$(call run-security)
	$(call run-fmt-check)
	$(call run-vet)
	$(call run-lint-maintenance)
	@$(MAKE) test
	@echo "✅ Monthly maintenance completed!"
	@echo "📋 Summary:"
	@echo "  - Dependencies updated to latest versions"
	@echo "  - Security vulnerabilities checked"
	@echo "  - Essential code quality validated"
	@echo "  - All tests passing"
	@echo ""
	@echo "💡 Consider running 'git status' to review dependency changes"

# Full CI pipeline (clean + deps + quality + coverage + build)
.PHONY: ci
ci: clean deps quality coverage build
	@echo "✅ CI pipeline completed!"

# Build release binary with version info
.PHONY: release
release: clean
	@echo "Building release binary..."
	@$(MAKE) build
	@echo "Release binary created: $(BINARY_PATH)"
	@echo "Version: $(VERSION)"

# =============================================================================
# Individual Targets (Available for Granular Control)
# =============================================================================

.PHONY: fmt fmt-check vet lint complexity-check gosec vuln-check mod-verify security deps-update deps-audit

# Code formatting
fmt:
	$(call run-fmt)

fmt-check:
	$(call run-fmt-check)

# Code analysis
vet:
	$(call run-vet)

lint:
	$(call run-lint)

complexity-check:
	$(call run-complexity-check)

# Security analysis
gosec:
	$(call run-gosec)

vuln-check:
	$(call run-vuln-check)

mod-verify:
	$(call run-mod-verify)

security:
	$(call run-security)

# Dependency management
deps-update:
	$(call run-deps-update)

deps-audit:
	$(call run-deps-audit)

# =============================================================================
# Help Documentation
# =============================================================================

# Show available targets
.PHONY: help
help:
	@echo "slack-butler Makefile (Go $(GO_VERSION))"
	@echo "Version: $(VERSION) | Commit: $(GIT_COMMIT)"
	@echo ""
	@echo "🚀 Main workflows (recommended):"
	@echo "  make dev         - Quick development cycle (format + vet + test + build)"
	@echo "  make quality     - Complete quality validation (security + format + vet + lint + complexity)"
	@echo "  make maintenance - Monthly maintenance (deps-update + quality + test)"
	@echo "  make ci          - Full CI pipeline (clean + deps + quality + coverage + build)"
	@echo ""
	@echo "📦 Core targets:"
	@echo "  build            - Build binary with version info"
	@echo "  test             - Run tests with race detection"
	@echo "  coverage         - Generate test coverage report"
	@echo "  clean            - Clean build artifacts and coverage files"
	@echo "  deps             - Install and tidy dependencies"
	@echo "  install-tools    - Install dev tools (golangci-lint, gocyclo, gosec, govulncheck)"
	@echo "  release          - Build release binary (clean + build)"
	@echo ""
	@echo "🔒 Security & Dependencies:"
	@echo "  security         - Complete security analysis (gosec + vuln-check + mod-verify)"
	@echo "  deps-audit       - Audit dependencies for vulnerabilities"
	@echo "  deps-update      - Update all dependencies to latest versions"
	@echo ""
	@echo "🔧 Individual targets (for granular control):"
	@echo "  fmt              - Format code with gofmt -s"
	@echo "  fmt-check        - Check code formatting (CI-friendly)"
	@echo "  vet              - Run go vet analysis"
	@echo "  lint             - Run golangci-lint"
	@echo "  complexity-check - Check cyclomatic complexity (threshold: 15)"
	@echo "  gosec            - Static security analysis"
	@echo "  vuln-check       - Check for known vulnerabilities"
	@echo "  mod-verify       - Verify module integrity"
	@echo ""
	@echo "💡 Note: Quality/security targets require 'make install-tools' first"
	@echo "💡 Add Go tools to PATH: export PATH=\$$PATH:~/go/bin"