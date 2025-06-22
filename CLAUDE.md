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
- **Time-based Filtering**: Support for various time formats (24h, 7d, 1w, etc.)

## Configuration
- **Token Storage**: Uses `.env` file (git-ignored) with `SLACK_TOKEN` environment variable
- **CLI Framework**: Built with Cobra for professional command structure
- **Configuration Management**: Viper for environment variables and flags

## Usage Examples
```bash
# Set token via environment
source .env

# Basic channel detection (last 24 hours)
./slack-buddy channels detect

# Custom time period with announcement
./slack-buddy channels detect --since=7d --announce-to=#general

# Using token flag directly
./slack-buddy channels detect --token=xoxb-your-token --since=3d
```

## Required Slack Permissions
- `channels:read` - To list channels
- `chat:write` - To post announcements

## Project Structure
```
slack-buddy-ai/
├── main.go              # Entry point
├── cmd/
│   ├── root.go         # Root command and configuration
│   └── channels.go     # Channel management commands
├── pkg/
│   └── slack/
│       └── client.go   # Slack API wrapper
├── .env                # Token storage (git-ignored)
├── .gitignore          # Excludes sensitive files
└── go.mod              # Go module definition
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
- **Version**: In Development (post v1.0.1-beta)
- **Status**: ✅ **COMMUNITY PROJECT** - Well-tested with basic security features
- **Security**: ✅ **BASIC SECURITY** - Security workflows configured, community-maintained
- **Branches**: 
  - `master` - Main development branch
  - `beta` - Current beta release branch (well-tested community project)
- **Tags**: v1.0.0-beta, v1.0.1-beta

## Testing Results
- **Workspace**: Successfully tested with "Vibe Coding, Inc." Slack workspace
- **Authentication**: ✅ Connected as `slack_buddy` bot user
- **Channel Detection**: ✅ Found 4 new channels in last 24h
- **Announcement Feature**: ✅ Posted formatted message to #announcements
- **Error Handling**: ✅ Provides clear, actionable error messages for:
  - Missing OAuth scopes (channels:read, groups:read, chat:write)
  - Invalid tokens
  - Bot not in channel
  - Channel not found

## Test Coverage (v1.0.1-beta)
- **Good Test Coverage** - Solid test coverage across core functionality
- **Comprehensive Test Suite** - Good test scenarios covering business logic
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

**Current Test Stats:**
- **383+ test cases** across comprehensive unit and integration suites
- **87.3% code coverage** achieved with legitimate, valuable tests
- **Unit Tests**: Core functionality, message formatting, time parsing, edge cases
- **Integration Tests**: Mock Slack API interactions, complete error handling scenarios
- **CLI Tests**: End-to-end command execution with dependency injection
- **Security Tests**: Token validation, input sanitization, error path coverage
- **Race Detection**: All tests pass with `-race` flag enabled
- **Mock Framework**: Robust Slack API mocking without artificial shortcuts

**Test Quality Standards:**
- Each test must have clear purpose and validate specific behavior
- Integration tests must use realistic mock data
- Error paths must be tested, not just happy paths
- Tests must be maintainable and readable
- No tests that always pass regardless of implementation
- **Good Coverage**: Solid coverage across core packages
- **Appropriate Gaps**: Remaining uncovered lines are external dependencies (network, OS)
- **No Artificial Inflation**: All test cases validate real functionality and scenarios

## Security Infrastructure (Community-Maintained)

### Security Workflows
- **GitHub Actions Security Workflow**: Security scanning workflows configured (may require fixes)
- **Vulnerability Scanning**: govulncheck integration available (requires working Go version)
- **Static Security Analysis**: gosec security scanner integration available
- **Dependency Monitoring**: Basic dependency monitoring via Dependabot
- **License Compliance**: Basic GPL/copyleft license detection
- **Code Quality**: golangci-lint with security and quality linters

### Security Tools & Configuration
- **Dependabot**: Weekly dependency updates for Go modules and GitHub Actions
- **Security Documentation**: Basic SECURITY.md with community vulnerability reporting process
- **Build System Integration**: Makefile with security targets (security-full, vuln-check, install-security)

### Security Status
- **Workflow Status**: ⚠️ **COMMUNITY-MAINTAINED** - Security workflows exist but may need fixes
- **Go Version**: ✅ **FIXED** - Updated GitHub Actions to use correct Go 1.24.4
- **Code Quality**: ✅ **IMPROVED** - Fixed gosec G104 unhandled error warning
- **Compliance**: ✅ **BASIC** - No known GPL dependencies

## Next Features (Ideas)
- Channel cleanup detection (inactive channels)
- User activity monitoring
- Automated channel archiving
- Integration with other workspace tools
- CI/CD pipeline with GoReleaser for automated releases
- Shell completions (bash, zsh, fish)