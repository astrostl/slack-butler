# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Channel cleanup detection for inactive channels
- User activity monitoring
- Automated channel archiving
- Web dashboard for workspace insights
- Scheduled automation features

## [0.1.0] - 2025-06-21

### Added
- Initial release of Slack Buddy AI CLI tool
- **Channel Detection Feature**: Detect new channels created within specified time periods
- **Flexible Time Filtering**: Support for various time formats (24h, 7d, 1w, etc.)
- **Smart Announcements**: Post formatted announcements to designated channels
- **Secure Configuration**: Environment-based token management with `.env` support
- **CLI Framework**: Built with Cobra for professional command structure
- **Slack API Integration**: Full integration with Slack API using official Go SDK

### Features
- `slack-buddy channels detect` command with the following options:
  - `--since` flag for time period specification (default: 24h)
  - `--announce-to` flag for target announcement channel
  - `--token` flag for direct token specification
- Environment variable support via `SLACK_TOKEN`
- Rich message formatting for channel announcements
- Error handling for API failures and authentication issues

### Technical Details
- Go module: `slack-buddy-ai`
- CLI tool name: `slack-buddy`
- Dependencies:
  - `github.com/spf13/cobra` - CLI framework
  - `github.com/slack-go/slack` - Slack API client
  - `github.com/spf13/viper` - Configuration management
- Security: `.gitignore` configured to exclude tokens and binaries

### Documentation
- Comprehensive README with installation and usage instructions
- CLAUDE.md for development notes and project tracking
- Inline help documentation for all commands