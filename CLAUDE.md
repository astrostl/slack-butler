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
│   ├── config/         # Configuration management
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

# Test help output
./slack-buddy --help

# Test channel detection
./slack-buddy channels detect --help
```

## Git Repository
- **Version**: 1.0.0 - Stable Release
- **Status**: ✅ **STABLE RELEASE** - Production-ready with comprehensive testing and security features
- **Security**: ✅ **COMMUNITY SECURITY** - Security tools available, community-maintained
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
- **Excellent Test Coverage** - Comprehensive test coverage across all packages
  - `cmd` package: 69.9% coverage
  - `pkg/slack` package: 89.2% coverage  
  - `pkg/logger` package: 75.0% coverage
- **Comprehensive Test Suite** - Legitimate test scenarios covering business logic
- **All Tests Passing** - Tests pass with race detection enabled
- **Error Path Coverage** - Key error scenarios and edge cases validated
- **Boundary Testing** - Time precision and API failure testing
- **Mock Infrastructure** - Realistic Slack API mocking for testing

## Development Guidelines

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
- Shell completions (bash, zsh, fish)