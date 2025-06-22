# Slack Buddy AI

[![Go Report Card](https://goreportcard.com/badge/github.com/astrostl/slack-buddy-ai)](https://goreportcard.com/report/github.com/astrostl/slack-buddy-ai)

A powerful Go CLI tool designed to help Slack workspaces become more useful, organized, and tidy through intelligent automation and monitoring.

**Version 1.0.3 - Stable Release** ✅

> **⚠️ Disclaimer**: This software is "vibe coded" (developed entirely using generative AI tools like Claude Code) and provided as-is without any warranties, guarantees, or official support. Use at your own risk in production environments.

## Features

- **🔍 Channel Detection**: Automatically detect new channels created within specified time periods
- **📢 Smart Announcements**: Announce new channels to designated channels with rich formatting
- **🚫 Idempotency**: Automatically prevents re-announcing channels that have already been announced
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
   - `channels:read` - To list public channels
   - `groups:read` - To list private channels
   - `chat:write` - To post announcements
   - `channels:history` - To prevent duplicate announcements (idempotency)
4. Install the app to your workspace and copy the Bot User OAuth Token

### 2. Configure Token
Create a `.env` file in the project directory:
```bash
SLACK_TOKEN=xoxb-your-bot-token-here
```

Or set the environment variable directly:
```bash
export SLACK_TOKEN=xoxb-your-bot-token-here
```

**Note**: If you get permission errors, the tool will tell you exactly which OAuth scopes to add in your Slack app settings. The `groups:read` scope is only needed if you want to detect private channels.

## Usage

### Health Check
```bash
# Basic health check
./slack-buddy health

# Detailed health check with verbose output
./slack-buddy health --verbose
```

### Basic Channel Detection
```bash
# Load environment variables
source .env

# Detect new channels from the last day
./slack-buddy channels detect

# Detect from last week
./slack-buddy channels detect --since=7
```

### With Announcements
```bash
# Detect and announce to #general
./slack-buddy channels detect --since=1 --announce-to=#general

# Detect from last 3 days and announce to #announcements
./slack-buddy channels detect --since=3 --announce-to=#announcements
```

### Dry Run Mode
```bash
# Preview what would be announced without posting
./slack-buddy channels detect --since=1 --announce-to=#general --dry-run

# Test announcement formatting and channel detection safely
./slack-buddy channels detect --since=7 --announce-to=#announcements --dry-run
```

### Using Token Flag
```bash
# Use token directly without .env file
./slack-buddy channels detect --token=xoxb-your-token --since=7
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
./slack-buddy health
./slack-buddy health --verbose
```

### `channels detect`
Detect new channels created within a specified time period.

**Flags:**
- `--since` - Number of days to look back (default: "1")
- `--announce-to` - Channel to announce new channels to
- `--dry-run` - Preview what would be announced without posting messages
- `--token` - Slack bot token (can also use SLACK_TOKEN env var)

**Examples:**
```bash
./slack-buddy channels detect --since=7 --announce-to=#general
./slack-buddy channels detect --since=1 --announce-to=#general --dry-run
```

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
- **Rate Limiting**: Exponential backoff prevents API abuse
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

- [ ] Channel cleanup detection (inactive channels)

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