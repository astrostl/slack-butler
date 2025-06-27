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
./bin/slack-butler channels archive
./bin/slack-butler channels archive --warn-days=30 --archive-days=30 --commit
./bin/slack-butler channels archive --exclude-channels="general,announcements" --commit
```

### Installed Usage
Replace `./bin/slack-butler` with `slack-butler` when using the installed version.

## Required Slack Permissions
- `channels:read` - To list public channels (**required**)
- `channels:history` - To read channel messages for activity detection (**required**)
- `chat:write` - To post announcements and warnings (**required**)
- `channels:join` - To join public channels for warnings (**required for archive**)
- `channels:manage` - To archive channels (**required for archive**)
- `users:read` - To resolve user names in messages (**required for enhanced features**)

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

### Version Management
```bash
# Version management (git tags only - no GitHub releases)
git tag v1.x.x       # Create version tag for reference
git push origin main --tags  # Push tags to remote
```

**Note**: This project uses git tags for version tracking but does NOT use GitHub releases.

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
- **Version**: 1.1.10 - Current stable release
- **Status**: ✅ **STABLE** - GoReleaser configuration removed, using git tags only
- **Security**: ✅ **COMMUNITY SECURITY** - Security tools available, community-maintained
- **Recent Updates**: Improved command usability with days-based timing, simplified documentation, consistent project branding
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

**Version Tagging Sequence (MANDATORY):**
```bash
# Required sequence before ANY version tag
make clean && make deps && make quality && make test && make coverage && make build

# Only proceed to tag if ALL commands succeed
git tag v1.x.x && git push origin main --tags
```

**Distribution Policy**: This project does NOT use GitHub releases. Users should:
- Install via `go install github.com/astrostl/slack-butler@latest`
- Build from source using git tags: `git checkout v1.x.x && go build`

**Rationale**: Quality is non-negotiable. Users depend on stable, secure code. Every version tag represents the project's commitment to excellence.

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

### Next Features (Ideas)
See [README.md roadmap](README.md#roadmap) for planned features and development roadmap.