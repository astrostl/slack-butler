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
./bin/slack-butler channels archive
./bin/slack-butler channels archive --warn-days=30 --archive-days=30 --commit
./bin/slack-butler channels archive --exclude-channels="general,announcements" --commit
```

## Required Slack Permissions
- `channels:read` - To list public channels (**required**)
- `chat:write` - To post announcements and warnings (**required**)
- `channels:join` - To join public channels for warnings (**required for archive**)
- `channels:manage` - To archive channels (**required for archive**)
- `users:read` - To resolve user names in messages (**required for enhanced features**)
- `groups:read` - To list private channels (*optional*)

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
├── docs/               # Documentation
├── bin/                # Build outputs
├── build/              # Build artifacts (coverage, reports)
├── .env                # Token storage (git-ignored)
├── .env.example        # Configuration template
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

# Full release workflow with quality checks
# NOTE: Requires Go tools in PATH (export PATH=$PATH:~/go/bin)
make release-full
```

### Setup and Tools
```bash
# Install development tools (versions pinned in go.mod)
make install-tools

# Install dependencies and tidy modules
make deps

# Clean build artifacts and coverage files
make clean

# NOTE: Security and quality targets require Go tools in PATH:
export PATH=$PATH:~/go/bin
# This applies to: security, gosec, vuln-check, quality, maintenance, ci targets
```

### Individual Quality Checks
```bash
# Code formatting
make fmt             # Format code with gofmt -s (required for commits)
make fmt-check       # Check formatting with simplify flag (CI-friendly)

# Code analysis
make vet             # Go vet analysis
make lint            # golangci-lint analysis
make complexity-check # Check cyclomatic complexity (threshold: 15)

# Security analysis
# NOTE: Security targets require Go tools in PATH (export PATH=$PATH:~/go/bin)
make gosec           # Static security analysis with gosec
make vuln-check      # Check for known vulnerabilities with govulncheck
make mod-verify      # Verify module integrity
make security        # Complete security analysis (all above)
```

### Security & Dependency Management
```bash
# Monthly security maintenance (RECOMMENDED)
make maintenance     # Update deps + run quality checks + test

# Dependency management
make deps-update     # Update all dependencies to latest versions
make deps-audit      # Audit dependencies for security vulnerabilities

# Security workflows
make security-update # Security checks with dependency updates
```

### Testing and Coverage
```bash
# Run tests with race detection
make test

# Generate test coverage report
make coverage
```

### Build and Release
```bash
# Build the tool
make build           # Build binary with embedded version info

# Release management
make release         # Create release with GoReleaser (standalone)
```

### Help and Documentation
```bash
# Show all available targets
make help

# Test CLI help output
./bin/slack-butler --help
./bin/slack-butler channels detect --help
```

## Git Repository
- **Version**: 1.1.6 - Stable Release
- **Status**: ✅ **STABLE RELEASE** - Production-ready with comprehensive testing and security features
- **Security**: ✅ **COMMUNITY SECURITY** - Security tools available, community-maintained
- **Recent Updates**: Enhanced OAuth scope validation and documentation consistency, added channels:history scope support (v1.1.6)
- **Branches**: 
  - `main` - Stable release branch

## Testing Results
- **Workspace**: Successfully tested with "Vibe Coding, Inc." Slack workspace
- **Authentication**: ✅ Connected as `slack_buddy` bot user
- **Channel Detection**: ✅ Found 4 new channels in last 24h
- **Announcement Feature**: ✅ Posted formatted message to #announcements
- **Error Handling**: ✅ Provides clear, actionable error messages for:
  - Missing OAuth scopes (channels:read, chat:write; groups:read optional)
  - Invalid tokens
  - Bot not in channel
  - Channel not found


## Test Coverage
- **Comprehensive Test Suite** - Legitimate test scenarios covering business logic
- **All Tests Passing** - Tests pass with race detection enabled  
- **Error Path Coverage** - Key error scenarios and edge cases validated
- **Boundary Testing** - Time precision and API failure testing
- **Mock Infrastructure** - Realistic Slack API mocking for testing

## Development Guidelines

### Code Review and Verification Principles
**CRITICAL**: Always verify technical claims before making changes:

- **Go Version Verification**: Never assume a Go version doesn't exist without checking. Go releases follow their own schedule and may have versions that seem unusual but are legitimate.
- **Dependency Verification**: Before claiming a dependency or feature doesn't exist, verify by checking documentation, running commands, or testing functionality.
- **Fact Checking**: When identifying "issues" in configuration or code, verify the issue actually exists rather than making assumptions based on patterns or expectations.
- **Version Validation**: Use `go version`, `go mod verify`, or other appropriate tools to check current state before suggesting changes.

**Example**: Go 1.24.4 is a legitimate version that exists and should not be "corrected" to an earlier version without explicit user request.

### Code Quality Standards

- **Code Formatting**: Always run `make fmt` before committing - uses `gofmt -s` for strict formatting with simplification
- **Format Verification**: Use `make fmt-check` to verify formatting in CI/automated checks
- **Cyclomatic Complexity**: Keep functions under 15 complexity threshold - use `make complexity-check` to verify
- **Quality Gate**: Use `make quality` to check both formatting and complexity before commits
- **Follow existing patterns** - Look at surrounding code for style and conventions
- **Error Handling** - Handle errors properly with meaningful, actionable messages
- **Logging Standards** - NEVER use stdout `fmt.Printf` for INFO or DEBUG logging unless a `--debug` flag is passed. Use structured logging via the logger package for internal diagnostics. Keep stdout clean for user-facing output only.

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
                    }).Info("Rate limited, waiting before retry")
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
**IMPORTANT**: Always maintain documentation with every release (stable or beta):

1. **CHANGELOG.md** - Update with new version, features, and changes
   - Follow semantic versioning (MAJOR.MINOR.PATCH)
   - Use `-beta`, `-alpha` suffixes for pre-releases
   - Document all breaking changes, new features, and bug fixes

2. **README.md** - Keep synchronized with current features
   - Update usage examples if commands change
   - Add new features to feature list
   - Update installation instructions if needed
   - Ensure roadmap reflects current plans

3. **CLAUDE.md** - Update development notes
   - Record version changes and release status
   - Update project structure if modified
   - Add new development commands or processes

### Version Strategy
- **Beta releases**: `1.x.x-beta` for feature-complete testing
- **Stable releases**: `1.x.x` for production-ready versions
- **Major versions**: Breaking changes or significant feature additions

### Release Quality Gates
**CRITICAL**: NEVER push releases to GitHub without passing ALL quality checks:

**MANDATORY Pre-Release Requirements:**
1. **Complete Test Suite**: Run `make test` - ALL tests must pass with 100% success rate
2. **Quality Checks**: Run `make quality` - ALL linting, security, and complexity checks must pass
3. **Coverage Validation**: Run `make coverage` - Ensure test coverage remains comprehensive
4. **Build Verification**: Run `make build` - Binary must compile successfully
5. **Race Condition Testing**: All tests must pass with race detection enabled

**ABSOLUTE PROHIBITIONS:**
- ❌ **NEVER** push releases with failing tests
- ❌ **NEVER** push releases with linting errors
- ❌ **NEVER** push releases with security issues (gosec failures)
- ❌ **NEVER** push releases with build failures
- ❌ **NEVER** skip quality checks "just this once"

**Release Command Sequence (MANDATORY):**
```bash
# Required sequence before ANY release push
make clean && make deps && make quality && make test && make coverage && make build

# Only proceed to release if ALL commands succeed
make release
```

**Rationale**: Quality is non-negotiable. Users depend on stable, secure releases. Every release represents the project's commitment to excellence.

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
- ✅ **Comprehensive test suite** covering core functionality
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

## Next Features (Ideas)
- Bulk channel operations
- Multi-workspace support
- Configurable warning message templates
- Scheduled archival policies