# Slack Buddy AI

A powerful Go CLI tool designed to help Slack workspaces become more useful, organized, and tidy through intelligent automation and monitoring.

## Features

- **🔍 Channel Detection**: Automatically detect new channels created within specified time periods
- **📢 Smart Announcements**: Announce new channels to designated channels with rich formatting
- **⏰ Flexible Time Filtering**: Support for various time formats (24h, 7d, 1w, etc.)
- **🔐 Secure Configuration**: Environment-based token management with git-safe storage
- **💡 Intelligent Error Handling**: Clear, actionable error messages for missing permissions and configuration issues
- **✅ Fully Tested**: Successfully tested with real Slack workspaces

## Installation

### Prerequisites
- Go 1.19 or higher
- Slack Bot Token with appropriate permissions

### Build from Source
```bash
git clone https://github.com/yourusername/slack-buddy-ai.git
cd slack-buddy-ai
go build -o slack-buddy
```

## Setup

### 1. Create a Slack App
1. Go to [Slack API](https://api.slack.com/apps)
2. Create a new app for your workspace
3. Add the following OAuth scopes:
   - `channels:read` - To list channels
   - `chat:write` - To post announcements
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

**Note**: If you get permission errors, the tool will tell you exactly which OAuth scopes to add in your Slack app settings.

## Usage

### Basic Channel Detection
```bash
# Load environment variables
source .env

# Detect new channels from the last 24 hours
./slack-buddy channels detect

# Detect from last week
./slack-buddy channels detect --since=7d
```

### With Announcements
```bash
# Detect and announce to #general
./slack-buddy channels detect --since=24h --announce-to=#general

# Detect from last 3 days and announce to #announcements
./slack-buddy channels detect --since=3d --announce-to=#announcements
```

### Using Token Flag
```bash
# Use token directly without .env file
./slack-buddy channels detect --token=xoxb-your-token --since=1w
```

### Time Format Examples
- `24h` - Last 24 hours
- `7d` - Last 7 days
- `1w` - Last week
- `48h` - Last 48 hours
- `30m` - Last 30 minutes

## Commands

### `channels detect`
Detect new channels created within a specified time period.

**Flags:**
- `--since` - Time period to look back (default: "24h")
- `--announce-to` - Channel to announce new channels to
- `--token` - Slack bot token (can also use SLACK_TOKEN env var)

**Examples:**
```bash
./slack-buddy channels detect --since=7d --announce-to=#general
```

## Development

### Project Structure
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
└── go.mod              # Dependencies
```

### Building
```bash
go build -o slack-buddy
```

### Testing
```bash
# Test help output
./slack-buddy --help
./slack-buddy channels detect --help

# Test with dry run (no announcements)
./slack-buddy channels detect --since=24h
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Security

- Never commit your `.env` file or Slack tokens
- Use environment variables for sensitive configuration
- The `.gitignore` file excludes all sensitive files by default

## Roadmap

- [ ] Channel cleanup detection (inactive channels)
- [ ] User activity monitoring
- [ ] Automated channel archiving
- [ ] Integration with other workspace tools
- [ ] Web dashboard for insights
- [ ] Scheduled automation

## License

MIT License - see LICENSE file for details

## Support

For issues, feature requests, or questions, please open an issue on GitHub.