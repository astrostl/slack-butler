# CLAUDE.md - Development Notes

## Project Overview
**slack-butler** - A Go CLI tool to help Slack workspaces be more useful and tidy.

- **Repository Name**: `slack-butler`
- **CLI Tool Name**: `slack-butler`
- **Language**: Go
- **Purpose**: Workspace management and automation for Slack

## Current Features
- **Channel Detection**: Detect new channels created during a specified time period
- **Channel Archival**: Detect inactive channels, warn about upcoming archival, and automatically archive channels after grace period (✅ **IMPLEMENTED**)
- **Warn-Only Mode**: Send inactivity warnings without archiving, with optional rewarn for stale warnings (`--warn-only`, `--rewarn-days`) (✅ **IMPLEMENTED**)
- **Default Channel Protection**: Automatically detects and protects workspace default channels from archival using user intersection heuristics (✅ **IMPLEMENTED**)
- **Default Channel Check**: Diagnostic flag (`--default-channel-check`) for archive command to see which channels are detected as defaults and which users are sampled (✅ **IMPLEMENTED**)
- **Random Channel Highlights**: Randomly select and highlight active channels to encourage discovery and participation (✅ **IMPLEMENTED**)
- **Health Checks**: Diagnostic command to verify configuration, permissions, and connectivity
- **Announcement System**: Optionally announce new channels to a target channel
- **Duplicate Prevention**: Scans last 15 messages to prevent duplicate announcements (✅ **IMPLEMENTED**)
- **Time-based Filtering**: Support for days-based time filtering (8 days default, configurable)
- **API Resilience**: Relies on Slack's native rate limiting responses rather than client-side rate limiting (✅ **IMPLEMENTED**)

## Configuration
- **Token Storage**: Uses `.env` file (git-ignored) with `SLACK_TOKEN` environment variable
- **CLI Framework**: Built with Cobra for professional command structure
- **Configuration Management**: Viper for environment variables and flags

## Usage Examples

### Development Usage (from source)
```bash
# Set token via environment
source .env

# Health check with basic output
./bin/slack-butler health

# Health check with detailed information
./bin/slack-butler health --verbose

# Basic channel detection (last 8 days - default)
./bin/slack-butler channels detect

# Custom time period with announcement (last 7 days)
./bin/slack-butler channels detect --since=7 --announce-to=#general

# Using token flag directly (last 3 days)
./bin/slack-butler channels detect --token=xoxb-your-token --since=3

# Dry run announcements without posting (default)
./bin/slack-butler channels detect --since=7 --announce-to=#general

# Actually post announcements
./bin/slack-butler channels detect --since=7 --announce-to=#general --commit

# Channel archival management (dry run mode by default)
# Automatically protects default channels (detected via user intersection)
./bin/slack-butler channels archive
./bin/slack-butler channels archive --warn-days=45 --archive-days=30 --commit
./bin/slack-butler channels archive --exclude-channels="general,announcements" --commit

# Override default channel protection (allows archiving of default channels)
./bin/slack-butler channels archive --include-default-channels --commit

# Customize default channel detection sample size (default: 10 users)
./bin/slack-butler channels archive --default-channel-sample-size=20 --commit

# Use environment variables for configuration
export SLACK_INCLUDE_DEFAULT_CHANNELS=true
export SLACK_DEFAULT_CHANNEL_SAMPLE_SIZE=15
export SLACK_DEFAULT_CHANNEL_THRESHOLD=0.85
./bin/slack-butler channels archive --commit

# Adjust detection threshold (default: 0.9 = 90%)
./bin/slack-butler channels archive --default-channel-threshold=0.95 --commit

# Check which channels are detected as defaults (diagnostic flag)
./bin/slack-butler channels archive --default-channel-check
./bin/slack-butler channels archive --default-channel-check --default-channel-sample-size=20
./bin/slack-butler channels archive --default-channel-check --default-channel-threshold=0.95

# Warn-only mode: send warnings without archiving (dry run preview)
./bin/slack-butler channels archive --warn-only

# Warn-only mode: actually send warnings (use with --commit)
./bin/slack-butler channels archive --warn-only --commit

# Warn-only mode with rewarn: re-warn channels whose warning is older than 30 days
./bin/slack-butler channels archive --warn-only --rewarn-days=30 --commit

# Random channel highlights (dry run mode by default)
./bin/slack-butler channels highlight
./bin/slack-butler channels highlight --count=5 --announce-to=#general
./bin/slack-butler channels highlight --count=1 --announce-to=#general --commit
```

### Installed Usage
Replace `./bin/slack-butler` with `slack-butler` when using the installed version.

### Docker Usage
```bash
# Pull the latest image
docker pull astrostl/slack-butler:latest

# Health check
docker run -e SLACK_TOKEN=$SLACK_TOKEN astrostl/slack-butler:latest health --verbose

# Detect new channels
docker run -e SLACK_TOKEN=$SLACK_TOKEN astrostl/slack-butler:latest channels detect --since=7

# Archive inactive channels (dry run)
docker run -e SLACK_TOKEN=$SLACK_TOKEN astrostl/slack-butler:latest channels archive --warn-days=45

# With commit flag to actually perform actions
docker run -e SLACK_TOKEN=$SLACK_TOKEN astrostl/slack-butler:latest channels detect --since=7 --announce-to=#general --commit
```

## Default Channel Protection

The archive command automatically detects and protects workspace default channels (channels that new members automatically join) from archival.

### How Detection Works
- **Membership Threshold Heuristic**: Samples recently joined workspace users and finds channels shared by a configurable percentage of sampled users
- **User Selection**: Takes the last N users from Slack's API response (Slack returns users oldest-first, so the last users are the most recently joined)
- **Sample Size**: Default 10 users (configurable via `--default-channel-sample-size` or `SLACK_DEFAULT_CHANNEL_SAMPLE_SIZE`)
- **Threshold**: Default 90% membership requirement (e.g., 9 out of 10 users) - configurable via `--default-channel-threshold` or `SLACK_DEFAULT_CHANNEL_THRESHOLD`
- **Automatic Protection**: Detected default channels are automatically excluded from archival consideration
- **Override Available**: Use `--include-default-channels` flag to disable protection
- **Verification**: Use `slack-butler channels archive --default-channel-check` to see which users are sampled and which channels are detected

### Why This Matters
Slack's public API does not expose workspace-level "Default Channels" configuration. This heuristic approach:
- Protects critical workspace channels without requiring admin API access
- Works on all Slack plans (Free, Pro, Business+, Enterprise Grid)
- Provides a reliable approximation by analyzing actual user membership patterns
- Resilient to edge cases where users leave default channels

### Example Output
```
🔍 Detecting default channels (sampling 10 recent users, 90% threshold)...
✅ Detected 9 default channels: #tech, #meta, #company-events, #company-general, #support, #company-announcements, #kudos, #company-all-hands, #random
   These channels will be excluded from archival.
   Use --include-default-channels to override this protection.
```

### Configuration Examples

| Threshold | Sample Size | Behavior |
|-----------|-------------|----------|
| **0.9** (default) | 10 | Requires 9/10 users - good balance |
| **0.95** | 10 | Requires 10/10 users - very strict |
| **0.8** | 10 | Requires 8/10 users - more permissive |
| **0.9** | 20 | Requires 18/20 users - higher confidence |

### API Efficiency
- **Moderate API Calls**: With 10 users sampled, typically makes 20-30 API calls total
- **Fast Detection**: Completes in seconds, even with Slack rate limits
- **Configurable**: Increase sample size for higher confidence in large workspaces

## Required Slack Permissions
- `channels:read` - To list public channels (**required**)
- `channels:history` - To read channel messages for activity detection (**required**)
- `chat:write` - To post announcements and warnings (**required**)
- `channels:join` - To join public channels for warnings (**required for archive**)
- `channels:manage` - To archive channels (**required for archive**)
- `users:read` - To resolve user names in messages and detect default channels (**required for enhanced features and default channel detection**)

## Project Structure
```
slack-butler/
├── main.go              # Entry point
├── cmd/                 # CLI commands and tests
│   ├── root.go         # Root command and configuration
│   ├── channels.go     # Channel management commands
│   └── *_test.go       # Command tests
├── pkg/                 # Core packages
│   ├── logger/         # Structured logging
│   └── slack/          # Slack API wrapper and client
├── bin/                # Build outputs
├── build/              # Build artifacts (coverage, reports)
├── .env                # Token storage (git-ignored)
├── .env.example        # Configuration template
├── Dockerfile          # Multi-stage Docker build
├── .dockerignore       # Docker build context exclusions
├── Makefile            # Build automation
└── go.mod              # Dependencies
```

## Development Commands

### Main Workflows
```bash
# Quick development cycle (format + vet + test + build)
make dev

# Complete quality validation (security + format + vet + lint + complexity)
# NOTE: Requires Go tools in PATH (export PATH=$PATH:~/go/bin)
make quality

# Monthly maintenance workflow (deps-update + quality + test) - RECOMMENDED
# NOTE: Requires Go tools in PATH (export PATH=$PATH:~/go/bin)
make maintenance

# Full CI pipeline (clean + deps + quality + coverage + build)
# NOTE: Requires Go tools in PATH (export PATH=$PATH:~/go/bin)
make ci
```

### Core Targets
```bash
# Build binary with version info
make build

# Run tests with race detection
make test

# Generate test coverage report
make coverage

# Clean build artifacts and coverage files
make clean

# Install and tidy dependencies
make deps

# Install development tools (versions pinned in go.mod)
make install-tools

# Build release binary (clean + build)
make release

# NOTE: Quality and security targets require Go tools in PATH:
export PATH=$PATH:~/go/bin
# This applies to: quality, maintenance, ci, security, lint, complexity-check, gosec targets
```

### Security & Dependencies
```bash
# Complete security analysis (gosec + vuln-check + mod-verify)
# NOTE: Requires Go tools in PATH (export PATH=$PATH:~/go/bin)
make security

# Audit dependencies for vulnerabilities
make deps-audit

# Update all dependencies to latest versions
make deps-update
```

### Individual Targets (Granular Control)
```bash
# Code formatting
make fmt             # Format code with gofmt -s (required for commits)
make fmt-check       # Check code formatting (CI-friendly)

# Code analysis
make vet             # Run go vet analysis
make lint            # Run golangci-lint (requires tools + PATH setup)
make complexity-check # Check cyclomatic complexity (threshold: 15, requires tools + PATH setup)

# Security analysis
# NOTE: Security targets require Go tools in PATH (export PATH=$PATH:~/go/bin)
make gosec           # Static security analysis
make vuln-check      # Check for known vulnerabilities
make mod-verify      # Verify module integrity
```

### Homebrew Release Targets
```bash
# Build macOS binaries for Homebrew (amd64 + arm64)
make build-macos-binaries

# Package macOS binaries into tar.gz archives
make package-macos-binaries

# Generate SHA256 checksums for macOS packages
make generate-macos-checksums

# Update Homebrew formula with version and checksums
# WARNING: Only use during development - NOT during actual releases
make update-homebrew-formula
```

### Docker Release Targets
```bash
# Build Docker image locally
make docker-build

# Tag Docker image for release
make docker-tag

# Build and push multi-platform images (linux/amd64, linux/arm64)
make docker-push

# Push single-platform images (for testing only)
make docker-push-single

# Create multi-platform manifests using manifest-tool
make docker-manifest

# Complete Docker release workflow (build + push)
make docker-release
```

### Version Management & Releases

For complete release instructions, see **[RELEASE.md](RELEASE.md)**.

### Help and Documentation
```bash
# Show all available targets
make help

# Test CLI help output (development)
./bin/slack-butler --help
./bin/slack-butler channels detect --help

# Test CLI help output (installed)
slack-butler --help
slack-butler channels detect --help
```

## Git Repository
- **Version**: 1.4.1 - Current stable release
- **Status**: ✅ **STABLE** - Homebrew tap + go install + Docker distribution
- **Security**: ✅ **COMMUNITY SECURITY** - Security tools available, community-maintained
- **Distribution**:
  - Homebrew: `brew install astrostl/slack-butler/slack-butler` (macOS)
  - Docker: `docker pull astrostl/slack-butler:latest` (all platforms, multi-arch)
  - Go Install: `go install github.com/astrostl/slack-butler@latest` (cross-platform)
  - Build from source: See README.md
- **Branches**:
  - `main` - Stable release branch

## Testing Results
- **Workspace**: Successfully tested with real Slack workspaces
- **Authentication**: ✅ Proper bot authentication and connection
- **Channel Detection**: ✅ Reliable detection of new channels across time periods
- **Announcement Feature**: ✅ Formatted message posting with duplicate prevention
- **Error Handling**: ✅ Clear, actionable error messages for common issues:
  - Missing OAuth scopes
  - Invalid tokens
  - Bot permissions
  - Channel access issues


## Test Coverage
- **Comprehensive Test Suite** - Legitimate test scenarios covering business logic
- **All Tests Passing** - Tests pass with race detection enabled  
- **Error Path Coverage** - Key error scenarios and edge cases validated
- **Boundary Testing** - Time precision and API failure testing
- **Mock Infrastructure** - Realistic Slack API mocking for testing

## Development Guidelines

### Code Review and Verification Principles
**CRITICAL**: Always verify technical claims before making changes:

- **Dependency Verification**: Before claiming a dependency or feature doesn't exist, verify by checking documentation, running commands, or testing functionality.
- **Fact Checking**: When identifying "issues" in configuration or code, verify the issue actually exists rather than making assumptions based on patterns or expectations.
- **Version Validation**: Use `go version`, `go mod verify`, or other appropriate tools to check current state before suggesting changes.
- **Version Pinning**: Pinning to exact Go versions (e.g., "1.23.4") in documentation is preferred over ranges (e.g., "1.24+") for clarity and reproducibility. Avoid over-analyzing version specifications.

### Code Quality Standards

- **Code Formatting**: Always run `make fmt` before committing - uses `gofmt -s` for strict formatting with simplification
- **Format Verification**: Use `make fmt-check` to verify formatting in CI/automated checks
- **Cyclomatic Complexity**: Keep functions under 15 complexity threshold - use `make complexity-check` to verify
- **Quality Gate**: Use `make quality` to check both formatting and complexity before commits
- **Follow existing patterns** - Look at surrounding code for style and conventions
- **Error Handling** - Handle errors properly with meaningful, actionable messages
- **Logging Standards** - 
  - **CRITICAL**: NEVER emit INFO level structured logging unless the CLI was passed a `--debug` flag
  - Use `logger.Debug()` for internal diagnostics - only shows when debug is enabled
  - Use `logger.Info()` ONLY when `--debug` flag is present 
  - NEVER use stdout `fmt.Printf` for INFO or DEBUG logging unless a `--debug` flag is passed
  - Use structured logging via the logger package for internal diagnostics
  - Keep stdout clean for user-facing output only
  - Default CLI output should be clean and minimal without debug noise

### API Integration Standards
**MANDATORY**: All Slack API functions must use standardized robust patterns:

- **Retry Logic**: Use `maxRetries = 3` pattern for all API calls
- **Rate Limiting**: Reactive approach - handle Slack's rate limit responses, parse retry-after directives
- **Timeout Handling**: Use context with appropriate timeouts (30s for quick ops, 2min for complex)
- **Error Detection**: Check for "rate_limited" errors and apply Slack's retry-after timing
- **Progress Feedback**: Show progress bars during rate limit waits using `showProgressBar()`
- **Graceful Degradation**: Functions should continue safely even if non-critical API calls fail
- **Consistent Patterns**: Follow established patterns from archiving functions for all new API integrations

**Example Pattern:**
```go
const maxRetries = 3
for attempt := 1; attempt <= maxRetries; attempt++ {
    // Make API call
    result, err := c.api.SomeAPICall(params)
    if err != nil {
        errStr := err.Error()
        if strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "rate limit") {
            if attempt < maxRetries {
                // Parse Slack's retry directive and wait
                waitDuration := parseSlackRetryAfter(errStr)
                if waitDuration > 0 {
                    logger.WithFields(logrus.Fields{
                        "attempt": attempt,
                        "max_tries": maxRetries,
                        "wait_duration": waitDuration,
                    }).Debug("Rate limited, waiting before retry")
                    showProgressBar(waitDuration)
                }
                continue
            }
        }
        return nil, fmt.Errorf("API call failed after %d attempts: %w", maxRetries, err)
    }
    return result, nil
}
```

### Release Process

**All release instructions are documented in [RELEASE.md](RELEASE.md).**

#### Go Module Proxy and Version Caching

**IMPORTANT**: Be aware of Go module proxy caching behavior when managing releases:

- **Module Proxy Caching**: The Go module proxy (proxy.golang.org) caches version information, including the @latest version
- **Deleted Tags**: If you delete a tag from GitHub after it's been cached by the proxy, the proxy will continue to serve the deleted version
- **Cache Duration**: The proxy can take 24-48 hours to naturally refresh and detect new versions
- **Impact on Tools**: Tools like Go Report Card use the proxy's version information, so they may show outdated versions during the cache period
- **No Manual Refresh**: There is no official way to force the proxy to immediately refresh its @latest cache
- **Best Practice**: Avoid deleting tags after release. Instead, create a new version (e.g., v1.3.1) and use the `retract` directive in go.mod if needed
- **Retract Directive**: Use `retract v1.x.x` in go.mod to officially mark versions as withdrawn (requires releasing in a subsequent version)

**Example Issue**: If you release v1.3.0, then delete the tag and create v1.3.1, the proxy may still report v1.3.0 as @latest for up to 48 hours, causing tools like goreportcard.com to display the old version even after refresh.

#### Testing Excellence
**MANDATORY**: Maintain comprehensive testing without artificial shortcuts.

**Testing Principles:**
- **No Skipped Tests**: Every test must pass legitimately or be fixed/removed
- **No Artificial Passes**: Never use `t.Skip()`, empty tests, or placeholder assertions to inflate numbers
- **Real Testing**: Tests must validate actual functionality, not just exercise code paths
- **Comprehensive Coverage**: Unit tests + integration tests + CLI tests
- **Mock Properly**: Use mocks for external dependencies, but test real logic
- **Test Real Scenarios**: Integration tests should simulate realistic user workflows

**Current Test Status:**
- ✅ **Comprehensive test coverage** covering core functionality
- ✅ **Unit tests** for business logic and message formatting
- ✅ **Integration tests** with mock Slack API interactions
- ✅ **CLI tests** for end-to-end command execution
- ✅ **Error path testing** for key scenarios including API failures and rate limiting
- ✅ **Race detection** enabled for all tests
- ✅ **Realistic mock framework** for external dependencies
- ✅ **API resilience testing** validates retry logic and graceful degradation
- ✅ **Duplicate detection testing** covers all edge cases and error scenarios

**Test Quality Standards:**
- Tests must validate actual functionality and behavior
- Integration tests use realistic mock data
- Error paths are tested along with happy paths
- Tests are maintainable and readable
- No artificial test inflation or empty placeholders
- Solid coverage of core business logic
- External dependencies properly mocked

## Security Infrastructure (Community-Maintained)

### Security Tools Available
- **Vulnerability Scanning**: govulncheck integration available (requires manual setup)
- **Static Security Analysis**: gosec security scanner integration available
- **License Compliance**: Basic GPL/copyleft license detection
- **Code Quality**: golangci-lint with security and quality linters

### Security Tools & Configuration
- **Security Documentation**: Basic SECURITY.md with community vulnerability reporting process
- **Build System Integration**: Makefile with security targets (security, vuln-check, gosec, mod-verify)

### Security Status
- **Code Quality**: ✅ **IMPROVED** - Fixed gosec G104 unhandled error warning
- **Compliance**: ✅ **BASIC** - No known GPL dependencies

## Development Policy

### Feature Planning
- **Roadmap Location**: Maintain project roadmap ONLY in README.md - do not create separate TODO files or duplicate roadmaps
- **Single Source**: README.md roadmap section is the authoritative source for planned features and development priorities
- **No Duplicate Planning**: Avoid creating separate TODO.md, BACKLOG.md, or other planning files that would diverge from README.md

### Documentation Standards
- **Development Disclaimer**: ALWAYS preserve the development disclaimer at the top of README.md after the main description. This disclaimer identifies the software as developed using generative AI tools and sets appropriate user expectations.
- **Format**: `> **⚠️ Disclaimer**: This software is "vibe coded" (developed entirely using generative AI tools like Claude Code) and provided as-is without any warranties, guarantees, or official support. Use at your own risk.`
- **Placement**: Must appear between the main project description and the version badge

### Next Features (Ideas)
See [README.md roadmap](README.md#roadmap) for planned features and development roadmap.