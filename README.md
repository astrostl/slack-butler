# Slack Buddy AI

[![Go Report Card](https://goreportcard.com/badge/github.com/astrostl/slack-buddy-ai)](https://goreportcard.com/report/github.com/astrostl/slack-buddy-ai)

A powerful Go CLI tool designed to help Slack workspaces become more useful, organized, and tidy through intelligent automation and monitoring.

**Version 1.0.5-dev - Post-Release Development** ✅

> **⚠️ Disclaimer**: This software is "vibe coded" (developed entirely using generative AI tools like Claude Code) and provided as-is without any warranties, guarantees, or official support. Use at your own risk in production environments.

## Features

- **🔍 Channel Detection**: Automatically detect new channels created within specified time periods
- **📁 Channel Archival**: Detect inactive channels, warn about upcoming archival, and automatically archive channels after grace period
- **📢 Smart Announcements**: Announce new channels to designated channels with rich formatting
- **🩺 Health Checks**: Diagnostic command to verify configuration, permissions, and connectivity
- **⏰ Flexible Time Filtering**: Support for days-based filtering (1, 7, 30, etc.)
- **🔐 Secure Configuration**: Environment-based token management with git-safe storage
- **🛡️ Security Features**: Basic security scanning and dependency monitoring (community-maintained)
- **💡 Intelligent Error Handling**: Clear, actionable error messages for missing permissions and configuration issues
- **✅ Well Tested**: Successfully tested with real Slack workspaces with good test coverage

## Installation

### Prerequisites
- Go 1.24.4
- Slack Bot Token with appropriate permissions

### Install via Go
```bash
go install github.com/astrostl/slack-buddy-ai@latest
```

**Note:** This will install the binary as `slack-buddy-ai`. You may want to create an alias:
```bash
# Add to your shell profile (.bashrc, .zshrc, etc.)
alias slack-buddy='slack-buddy-ai'
```

### Build from Source
```bash
git clone https://github.com/astrostl/slack-buddy-ai.git
cd slack-buddy-ai
go build -o slack-buddy
```

## Setup

### 1. Create a Slack App
1. Go to [Slack API](https://api.slack.com/apps)
2. Create a new app for your workspace
3. Add the following OAuth scopes:
   - `channels:read` - To list public channels (**required**)
   - `channels:join` - To join public channels for warnings (**required for archive**)
   - `channels:manage` - To archive channels (**required for archive**)
   - `chat:write` - To post announcements and warnings (**required**)
   - `groups:read` - To list private channels (*optional*)
   - `users:read` - To resolve user names in messages (*optional*)
4. Install the app to your workspace and copy the Bot User OAuth Token

### 2. Configure Token
Create a `.env` file in the project directory:
```bash
cp .env.example .env
# Edit .env and add your token, then:
source .env
```

Or set the environment variable directly:
```bash
export SLACK_TOKEN=xoxb-your-bot-token-here
```

**Note**: The `.env.example` file uses `export` statements to ensure environment variables are available to the slack-buddy binary when you `source .env`.

**Note**: If you get permission errors, the tool will tell you exactly which OAuth scopes to add in your Slack app settings. The `groups:read` scope is only needed if you want to detect private channels - without it, the tool will only detect public channels.

## Important Limitations

### Slack API Rate Limits
**Duplicate Detection Limitations**: Due to Slack API restrictions for non-Marketplace apps:
- `GetConversationHistory` is limited to **1 request per minute**
- Maximum **15 messages** returned per request
- The duplicate detection feature only checks the **last 15 messages** in the announcement channel

**Impact**: If more than 15 messages are posted to your announcement channel between runs, some duplicate announcements may not be detected. For high-traffic channels, consider running the tool more frequently or using a dedicated low-traffic channel for announcements.

## Usage

### Health Check
```bash
# Basic health check
./bin/slack-buddy health

# Detailed health check with verbose output
./bin/slack-buddy health --verbose
```

### Basic Channel Detection
```bash
# Load environment variables
source .env

# Detect new channels from the last 8 days (default)
./bin/slack-buddy channels detect

# Detect from last week
./bin/slack-buddy channels detect --since=7
```

### Channel Archival Management
```bash
# Dry run what channels would be warned/archived (default mode)
./bin/slack-buddy channels archive

# Actually warn channels inactive for 5 minutes, archive after 1 minute grace period
./bin/slack-buddy channels archive --warn-seconds=300 --archive-seconds=60 --commit

# Exclude important channels from archival
./bin/slack-buddy channels archive --exclude-channels="general,announcements" --exclude-prefixes="prod-,admin-" --commit
```

### With Announcements
```bash
# Detect and announce to #general
./bin/slack-buddy channels detect --since=1 --announce-to=#general

# Detect from last 3 days and announce to #announcements
./bin/slack-buddy channels detect --since=3 --announce-to=#announcements
```

### Dry Run vs Commit Mode
```bash
# Dry run what would be announced without posting (default)
./bin/slack-buddy channels detect --since=7 --announce-to=#general

# Actually post announcements
./bin/slack-buddy channels detect --since=7 --announce-to=#general --commit

# The default behavior is safe dry run mode
```

### Using Token Flag
```bash
# Use token directly without .env file
./bin/slack-buddy channels detect --token=xoxb-your-token --since=7
```

### Time Format Examples
- `1` - Last 1 day (24 hours)
- `7` - Last 7 days (1 week)
- `2` - Last 2 days (48 hours)
- `0.5` - Last 12 hours (half day)
- `30` - Last 30 days (1 month)

## Commands

### `health`
Check Slack connectivity and validate configuration.

**Purpose:**
- Verify token validity and format
- Test Slack API connectivity
- Check required OAuth scopes and permissions
- Validate bot user information
- Test basic API functionality

**Flags:**
- `-v, --verbose` - Show detailed health check information

**Examples:**
```bash
./bin/slack-buddy health
./bin/slack-buddy health --verbose
```

### `channels detect`
Detect new channels created within a specified time period.

**Flags:**
- `--since` - Number of days to look back (default: "8")
- `--announce-to` - Channel to announce new channels to
- `--commit` - Actually post messages (default is dry run mode)
- `--token` - Slack bot token (can also use SLACK_TOKEN env var)

**Examples:**
```bash
./bin/slack-buddy channels detect --since=7 --announce-to=#general
./bin/slack-buddy channels detect --since=1 --announce-to=#general --commit
```

### `channels archive`
Manage inactive channel archival with automated warnings and grace periods.

**Flags:**
- `--warn-seconds` - Seconds of inactivity before warning (default: 300)
- `--archive-seconds` - Seconds after warning before archiving (default: 60)
- `--exclude-channels` - Comma-separated list of channels to exclude
- `--exclude-prefixes` - Comma-separated list of prefixes to exclude
- `--commit` - Actually warn and archive channels (default is dry run mode)
- `--token` - Slack bot token (can also use SLACK_TOKEN env var)

**Note:** Currently using seconds for testing purposes.

**Examples:**
```bash
./bin/slack-buddy channels archive
./bin/slack-buddy channels archive --warn-seconds=1800 --archive-seconds=300 --commit
./bin/slack-buddy channels archive --exclude-channels="general,random" --commit
```

**Required OAuth Scopes for Archive:**
- `channels:read` (to list channels)
- `channels:join` (to join public channels)
- `chat:write` (to post warnings)
- `channels:manage` (to archive channels)
- `users:read` (optional, for user name resolution)

## Development

### Project Structure
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

### Building
```bash
go build -o slack-buddy
```

### Testing
```bash
# Run all tests with race detection
make test

# Generate coverage report
make coverage
```

### Development Commands
```bash
# Install development and security tools
make install-tools

# Quick development cycle (format + vet + test + build)
make dev

# Complete quality validation (security + format + vet + lint + complexity)
make quality

# Monthly maintenance (update dependencies + quality checks + test)
make maintenance

# Full CI pipeline (clean + deps + quality + coverage + build)
make ci

# Full release workflow with quality checks
make release-full

# Install dependencies
make deps

# Clean build artifacts
make clean
```

### Individual Quality Checks
```bash
# Format code
make fmt

# Check formatting (CI-friendly)
make fmt-check

# Vet code
make vet

# Lint code
make lint

# Check cyclomatic complexity
make complexity-check

# Static security analysis
make gosec

# Check for vulnerabilities
make vuln-check

# Verify module integrity
make mod-verify

# Complete security analysis
make security
```

### Release Management

#### Local Release Process
```bash
# Install development tools (includes GoReleaser)
make install-tools

# Create full release with checksums (includes quality checks)
make release-full

# Standalone release creation
make release
```

#### Release Features
- **Multi-platform builds**: Linux, macOS, Windows (amd64, arm64)
- **Checksums**: SHA256 checksums for all artifacts
- **Archives**: tar.gz (Unix) and zip (Windows) with documentation
- **Local-only**: No GitHub integration (release.disable: true)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Security

### Built-in Security Features
- **Token Validation**: All Slack tokens are validated and sanitized
- **Rate Limiting**: Respects Slack's native rate limiting responses
- **Input Sanitization**: All user inputs are validated
- **No Token Logging**: Tokens never appear in logs or error messages

### Security Scanning (Best Effort)
- **Static Analysis**: gosec security scanner integration available
- **License Compliance**: Basic license scanning to avoid GPL dependencies

> **Note**: Security scanning tools are available but require manual setup. No guarantees provided.

### Best Practices
- Never commit your `.env` file or Slack tokens
- Use bot tokens (`xoxb-`) with minimal required OAuth scopes
- Store tokens in environment variables, never in code
- Test `make security` locally (may require tool installation)

See [SECURITY.md](SECURITY.md) for vulnerability reporting and detailed security information.

## Roadmap

- [x] Channel cleanup detection (inactive channels) - **Implemented**
- [ ] Bulk channel operations
- [ ] Multi-workspace support
- [ ] Configurable warning message templates

## License

MIT License - see LICENSE file for details

## Disclaimer & Warranty

**This software is "vibe coded" and comes with no warranties or guarantees.**

- ❌ **No official support** - Community-driven, best-effort maintenance
- ❌ **No production guarantees** - Use at your own risk
- ❌ **No liability** - Authors not responsible for any damages or issues
- ❌ **No SLA or uptime commitments** - Volunteer-maintained project

**Before using in production:**
- Thoroughly test in your environment
- Implement your own monitoring and error handling
- Have backup plans for critical operations
- Review all code and dependencies yourself

## Support

This is a community project with **no official support**. For issues, feature requests, or questions, please open an issue on GitHub with the understanding that responses are provided on a best-effort basis.
