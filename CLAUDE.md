# CLAUDE.md - Development Notes

## Project Overview
**slack-buddy-ai** - A Go CLI tool to help Slack workspaces be more useful and tidy.

- **Repository Name**: `slack-buddy-ai`
- **CLI Tool Name**: `slack-buddy`
- **Language**: Go
- **Purpose**: Workspace management and automation for Slack

## Current Features
- **Channel Detection**: Detect new channels created during a specified time period
- **Announcement System**: Optionally announce new channels to a target channel
- **Time-based Filtering**: Support for days-based time filtering (1, 7, 30, etc.)
- **Idempotency**: Automatically prevents re-announcing channels that have already been announced by reading message history

## Configuration
- **Token Storage**: Uses `.env` file (git-ignored) with `SLACK_TOKEN` environment variable
- **CLI Framework**: Built with Cobra for professional command structure
- **Configuration Management**: Viper for environment variables and flags

## Usage Examples
```bash
# Set token via environment
source .env

# Basic channel detection (last 1 day)
./slack-buddy channels detect

# Custom time period with announcement (last 7 days)
./slack-buddy channels detect --since=7 --announce-to=#general

# Using token flag directly (last 3 days)
./slack-buddy channels detect --token=xoxb-your-token --since=3

# Dry run to preview announcements without posting
./slack-buddy channels detect --since=7 --announce-to=#general --dry-run
```

## Required Slack Permissions
- `channels:read` - To list public channels
- `groups:read` - To list private channels
- `channels:history` - To read announcement channel history for idempotency
- `chat:write` - To post announcements

## Project Structure
```
slack-buddy-ai/
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
```bash
# Build the tool
go build -o slack-buddy

# Code quality checks
make fmt             # Format code with gofmt -s (required for commits)
make fmt-check       # Check formatting with simplify flag (CI-friendly)
make complexity-check # Check cyclomatic complexity
make quality         # Run all quality checks (fmt-check + complexity-check)

# Development workflow
make dev             # Quick cycle: fmt, vet, test, build
make ci              # Full CI checks including complexity

# Test help output
./slack-buddy --help

# Test channel detection
./slack-buddy channels detect --help
```

## Git Repository
- **Version**: 1.0.2 - Current Stable Release
- **Status**: ✅ **STABLE RELEASE** - Production-ready with comprehensive testing and security features
- **Security**: ✅ **COMMUNITY SECURITY** - Security tools available, community-maintained
- **Recent Updates**: Module path fixes and additional test coverage improvements (v1.0.1, v1.0.2)
- **Branches**: 
  - `main` - Stable release branch

## Testing Results
- **Workspace**: Successfully tested with "Vibe Coding, Inc." Slack workspace
- **Authentication**: ✅ Connected as `slack_buddy` bot user
- **Channel Detection**: ✅ Found 4 new channels in last 24h
- **Announcement Feature**: ✅ Posted formatted message to #announcements
- **Error Handling**: ✅ Provides clear, actionable error messages for:
  - Missing OAuth scopes (channels:read, groups:read, chat:write, channels:history)
  - Invalid tokens
  - Bot not in channel
  - Channel not found

## Idempotency Feature
- **Smart Duplicate Prevention**: Reads announcement channel message history to identify previously announced channels
- **Bot-Only Filtering**: Only processes messages posted by the bot itself, ignoring user messages for accuracy
- **Automatic Filtering**: Only announces channels that haven't been announced before
- **History Analysis**: Parses up to 1000 recent messages in the announcement channel to extract channel IDs
- **Permission Enforcement**: Requires proper permissions and fails fast if `channels:history` scope is missing to prevent duplicate announcements
- **Channel ID Extraction**: Uses regex pattern matching to find `<#CHANNEL_ID>` mentions in messages
- **Rate Limiting**: Respects Slack API rate limits when reading message history
- **Efficient Processing**: Skips messages from other users, focusing only on bot's own announcements

## Test Coverage
- **Comprehensive Test Coverage** - Strong test coverage across all packages and core functionality
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
- **Logging** - Add appropriate logging (debug level for verbose information)

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
- Good test suite covering core functionality
- Unit tests for business logic and message formatting
- Integration tests with mock Slack API interactions
- CLI tests for end-to-end command execution
- Error path testing for key scenarios
- Race detection enabled for all tests
- Realistic mock framework for external dependencies

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
- **Build System Integration**: Makefile with security targets (security-full, vuln-check, install-security)

### Security Status
- **Code Quality**: ✅ **IMPROVED** - Fixed gosec G104 unhandled error warning
- **Compliance**: ✅ **BASIC** - No known GPL dependencies

## Next Features (Ideas)
- Channel cleanup detection (inactive channels)
- GoReleaser for automated releases