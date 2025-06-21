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
- **Status**: Initialized with 2 commits
- **Security**: Token and binaries excluded via .gitignore
- **Branch**: master (ready for GitHub)

## Next Features (Ideas)
- Channel cleanup detection (inactive channels)
- User activity monitoring
- Automated channel archiving
- Integration with other workspace tools