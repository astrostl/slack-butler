# Slack Butler

[![Go Report Card](https://goreportcard.com/badge/github.com/astrostl/slack-butler?v=2)](https://goreportcard.com/report/github.com/astrostl/slack-butler)

A powerful Go CLI tool designed to help Slack workspaces become more useful, organized, and tidy through intelligent automation and monitoring.

> **⚠️ Disclaimer**: This software is "vibe coded" (developed entirely using generative AI tools like Claude Code) and provided as-is without any warranties, guarantees, or official support. Use at your own risk.

**Version 1.1.12** ✅

## Features

- **🔍 Channel Detection**: Automatically detect new channels created within specified time periods
- **📁 Channel Archival**: Detect inactive channels, warn about upcoming archival, and automatically archive channels after grace period
- **🧭 Random Channel Highlights**: Randomly select and highlight active channels to encourage discovery and participation
- **📢 Smart Announcements**: Announce new channels to designated channels with rich formatting
- **🩺 Health Checks**: Diagnostic command to verify configuration, permissions, and connectivity
- **⏰ Flexible Time Filtering**: Support for days-based filtering (1, 7, 30, etc.)
- **🔐 Secure Configuration**: Environment-based token management with git-safe storage
- **🛡️ Security Features**: Basic security scanning and dependency monitoring (community-maintained)
- **💡 Intelligent Error Handling**: Clear, actionable error messages for missing permissions and configuration issues
- **✅ Well Tested**: Comprehensive test coverage with API validation

## Installation

### Prerequisites
- Go 1.24.4 or later
- Slack Bot Token with appropriate permissions

### Install via Go
```bash
go install github.com/astrostl/slack-butler@latest
```

**Note:** This installs the binary as `slack-butler` in your `~/go/bin` directory. Make sure `~/go/bin` is in your PATH:
```bash
export PATH=$PATH:~/go/bin
```

### Build from Source
```bash
git clone https://github.com/astrostl/slack-butler.git
cd slack-butler
go build -o slack-butler
```

## Setup

### 1. Create a Slack App
1. Go to [Slack API](https://api.slack.com/apps)
2. Create a new app for your workspace
3. Add the following OAuth scopes:
   - `channels:read` - To list public channels
   - `channels:join` - To join public channels for message checks and announcements
   - `channels:manage` - To archive channels
   - `channels:history` - To check for activity and announcements
   - `chat:write` - To post announcements and warnings
   - `users:read` - To resolve user names in messages
4. Install the app to your workspace and copy the Bot User OAuth Token

### 2. Configure Token
Create a `.env` file in the project directory using the provided template:
```bash
cp .env.example .env
# Edit .env and replace the example token with your actual token, then:
source .env
```

Or set the environment variable directly:
```bash
export SLACK_TOKEN=xoxb-your-bot-token-here
```

**Note**: The `.env.example` file uses `export` statements to ensure environment variables are available to the slack-butler binary when you `source .env`.

**Note**: If you get permission errors, the tool will tell you exactly which OAuth scopes to add in your Slack app settings.

**Important**: All listed OAuth scopes are required. The tool will fail without proper permissions.

## Important Limitations

### Slack API Rate Limits
**Duplicate Detection Limitations**: Due to Slack API restrictions:
- Duplicate detection checks recent messages in the announcement channel
- API rate limits may apply to frequent requests

**Impact**: For high-traffic announcement channels, consider running the tool more frequently or using a dedicated low-traffic channel for announcements to ensure optimal duplicate detection.

## Usage

### Health Check
```bash
# Basic health check
slack-butler health

# Detailed health check with verbose output
slack-butler health --verbose
```

### New Channel Detection
```bash
# Load environment variables
source .env

# Detect new channels from the last 8 days (default)
slack-butler channels detect

# Detect from last week
slack-butler channels detect --since=7
```

### Warning and Archiving Inactive Channels
```bash
# Dry run what channels would be warned/archived (default mode)
slack-butler channels archive

# Actually warn channels inactive for 30 days, archive after 30 days grace period
slack-butler channels archive --warn-days=30 --archive-days=30 --commit

# Exclude important channels from archival
slack-butler channels archive --exclude-channels="general,announcements" --exclude-prefixes="prod-,admin-" --commit
```

### Dry Run vs Commit Mode
```bash
# Dry run what would be announced without posting (default)
slack-butler channels detect --since=7 --announce-to=#general

# Actually post announcements
slack-butler channels detect --since=7 --announce-to=#general --commit

# The default behavior is safe dry run mode
```

### Using Token Flag
```bash
# Use token directly without .env file
slack-butler channels detect --token=xoxb-your-token --since=7
```

**⚠️ Security Warning**: Using `--token` directly in commands may expose your token in shell history. Use environment variables or `.env` files for better security.

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
slack-butler health
slack-butler health --verbose
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
slack-butler channels detect --since=7 --announce-to=#general
slack-butler channels detect --since=1 --announce-to=#general --commit
```

### `channels archive`
Manage inactive channel archival with automated warnings and grace periods.

**Flags:**
- `--warn-days` - Days of inactivity before warning (default: 30)
- `--archive-days` - Days after warning before archiving (default: 30)
- `--exclude-channels` - Comma-separated list of channels to exclude
- `--exclude-prefixes` - Comma-separated list of prefixes to exclude
- `--commit` - Actually warn and archive channels (default is dry run mode)
- `--token` - Slack bot token (can also use SLACK_TOKEN env var)

**Note:** Archive timing supports decimal precision (e.g., 0.5 = 12 hours, 7.5 = 7.5 days). While sub-day precision is available, day-based values are recommended for practical channel management.

**Examples:**
```bash
slack-butler channels archive
slack-butler channels archive --warn-days=60 --archive-days=14 --commit
slack-butler channels archive --exclude-channels="general,random" --commit
```

### `channels highlight`
Randomly select and highlight active channels to encourage discovery and participation.

**Flags:**
- `--count` - Number of random channels to highlight (default: 3)
- `--announce-to` - Channel to announce highlights to (required when using --commit)
- `--commit` - Actually post messages (default is dry run mode)
- `--token` - Slack bot token (can also use SLACK_TOKEN env var)

**Examples:**
```bash
slack-butler channels highlight
slack-butler channels highlight --count=5 --announce-to=#general
slack-butler channels highlight --count=1 --announce-to=#general --commit
```


## Development

### Project Structure
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
├── bin/                # Build outputs (git-ignored)
├── build/              # Build artifacts (git-ignored)
├── .env.example        # Configuration template
├── .env                # Token storage (git-ignored)
├── Makefile            # Build automation
└── go.mod              # Dependencies
```

### Building
```bash
go build -o slack-butler
```

### Testing
```bash
# Run all tests with race detection
make test

# Generate coverage report
make coverage
```

### Development Tool Setup (Required First)
```bash
# REQUIRED: Install development and security tools first
make install-tools

# REQUIRED: Add Go tools to PATH for quality/security targets
export PATH=$PATH:~/go/bin
```

### Main Workflows
```bash
# Quick development cycle (format + vet + test + build)
make dev

# Complete quality validation (security + format + vet + lint + complexity)
make quality

# Monthly maintenance (deps-update + quality + test)
make maintenance

# Full CI pipeline (clean + deps + quality + coverage + build)
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

# Install dev tools (golangci-lint, gocyclo, gosec, govulncheck)
make install-tools

# Build release binary (clean + build)
make release
```

### Security & Dependencies
```bash
# Complete security analysis (gosec + vuln-check + mod-verify)
make security

# Audit dependencies for vulnerabilities
make deps-audit

# Update all dependencies to latest versions
make deps-update
```

### Individual Targets (Granular Control)
```bash
# Format code with gofmt -s
make fmt

# Check code formatting (CI-friendly)
make fmt-check

# Run go vet analysis
make vet

# Run golangci-lint
make lint

# Check cyclomatic complexity (threshold: 15)
make complexity-check

# Static security analysis
make gosec

# Check for known vulnerabilities
make vuln-check

# Verify module integrity
make mod-verify
```

**Important:** Many targets require development tools. Run `make install-tools` and `export PATH=$PATH:~/go/bin` first.

### Version Management

This project uses git tags for version tracking but does NOT use GitHub releases.

#### Installing Specific Versions
```bash
# Install latest version
go install github.com/astrostl/slack-butler@latest

# Install specific version
go install github.com/astrostl/slack-butler@v1.1.11

# Install development version
go install github.com/astrostl/slack-butler@main

# Build from specific version
git clone https://github.com/astrostl/slack-butler.git
cd slack-butler
git checkout v1.1.11
go build -o slack-butler
```

#### For Developers: Version Tagging
```bash
# Tag new versions (maintainers only)
git tag v1.x.x
git push origin main --tags
```

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

- [x] New channel detection - **Implemented**
- [x] Channel cleanup detection (inactive channels) - **Implemented**
- [x] Random channel highlight feature - **Implemented**
- [ ] Interactive setup wizard (`slack-butler init`)
- [ ] Multi-workspace support
- [ ] Configurable warning message templates

## License

MIT License - see LICENSE file for details

## Disclaimer & Warranty

**This software comes with no warranties or guarantees.**

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
