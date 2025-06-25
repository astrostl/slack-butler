# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.2] - 2025-06-25

### Fixed
- **Development Tools**: Fixed tools.go imports and Makefile tool installation for non-importable packages
- **Documentation**: Removed temporary "testing" language from archive command help text
- **Dependencies**: Major cleanup of go.mod removing ~2,400 lines of unnecessary dependencies

### Improved
- **Build System**: Enhanced Makefile to handle standalone tool installation properly
- **User Experience**: Improved clarity of archive command timing configuration messaging
- **Project Maintenance**: Streamlined dependency management for better maintainability

## [1.1.1] - 2025-06-25

### Fixed
- **Code Quality**: Refactored test functions to reduce cyclomatic complexity below 15 threshold
- **Test Reliability**: Fixed test expectations for duration formatting (1 hour vs 60 minutes)
- **Code Structure**: Improved test organization with smaller, focused helper functions

### Improved
- **Test Maintenance**: Extracted complex test setup into reusable helper functions
- **Code Readability**: Better separation of concerns in test validation and setup
- **Quality Compliance**: Achieved 100% quality gate compliance with all checks passing

## [1.1.0] - 2025-06-25

### Added
- **Channel Archival System**: Complete implementation of channel archival functionality with warnings and grace periods
- **Enhanced Testing**: Comprehensive test improvements including archive functionality tests
- **Security Enhancements**: Additional security features and code quality improvements

### Fixed
- **Test Coverage**: Fixed API signature issues in inactive channel tests (missing isDebug parameter)
- **Rate Limiting**: Fixed missing rate limiter call in GetNewChannels method for consistent API throttling  
- **Documentation Accuracy**: Updated README.md to reflect actual command defaults and flag behavior
  - Corrected `--since` default from "1" to "8" days
  - Updated flag documentation to use `--commit` instead of `--dry-run` for consistency

### Improved  
- **Issue Templates**: Enhanced GitHub issue templates with CLI-specific fields for better bug reporting
- **Makefile Documentation**: Clarified maintenance target description for accuracy
- **Code Quality**: Improved maintenance lint configuration to keep critical error checking while relaxing style requirements


## [1.0.4] - 2025-06-23

### Fixed
- **OAuth Scope Documentation**: Removed incorrect `channels:history` requirement from .env.example 
- **Dependabot Configuration**: Fixed empty package-ecosystem field in .github/dependabot.yml
- **Installation Documentation**: Clarified go install method availability in README.md
- **Development Tools**: Improved Makefile install-tools reliability using go list instead of awk

### Improved
- **Documentation Accuracy**: All configuration examples now reflect actual requirements
- **Development Workflow**: More reliable tool installation for consistent development environment

## [1.0.3] - 2025-06-22

### Fixed
- **Documentation Accuracy**: Updated all documentation to reflect current Makefile targets and functionality
- **CLAUDE.md**: Removed outdated references and test coverage percentages
- **README.md**: Fixed reference to non-existent `make security-full` target
- **Code Quality**: Fixed unhandled error warnings in cmd/root.go for better gosec compliance

### Improved
- **Documentation Consistency**: All documentation now accurately reflects current project state
- **Makefile Integration**: Optimized development workflow with proper target composition

## [1.0.2] - 2025-06-22

### Added
- **CONTRIBUTING.md**: Comprehensive contribution guidelines for community project
- **Extended Test Suite**: Added main_test.go and additional integration tests
- **Enhanced Makefile**: Complete rewrite with organized target structure and comprehensive workflow management
  - **Main Workflows**: `dev`, `quality`, `ci`, `release-full` for different development stages
  - **Helper Functions**: Centralized logic for formatting, linting, security checks with consistent error reporting
  - **Tool Management**: Automated installation of development, linting, security, and release tools via `install-tools`
  - **Security Integration**: Complete security pipeline with `gosec`, `govulncheck`, and module verification
  - **Build Optimization**: Enhanced build process with embedded version info, build time, and git commit
  - **Release Automation**: GoReleaser integration for multi-platform releases with checksums
  - **Quality Gates**: Comprehensive quality checks including formatting, vetting, linting, and complexity analysis
  - **CI Pipeline**: Full continuous integration workflow with dependency management and artifact cleanup

### Improved
- **Developer Experience**: Streamlined workflows reduce common development tasks to single commands
- **Code Quality**: Automated formatting, linting, and complexity checking prevents technical debt
- **Security**: Integrated vulnerability scanning and static security analysis in development workflow
- **Build Consistency**: Standardized build process with embedded metadata for better version tracking
- **Documentation**: Enhanced help system with organized target categories and clear usage examples

### Fixed  
- **Import Paths**: Updated all import paths to use full GitHub module path for proper dependency resolution

## [1.0.1] - 2025-06-22

### Fixed
- **Module Path**: Fixed Go module path to match GitHub repository URL

## [1.0.0] - 2025-06-22

### Added - Stable Release
- **Logger Test Coverage**: Comprehensive test suite for pkg/logger package with strong coverage
- **Documentation Accuracy**: Fixed incorrect flag syntax in .env.example
- **Test Quality**: All tests passing with legitimate coverage (no artificial shortcuts)

### Security Infrastructure (Community-Maintained)
- **Security Documentation**: Basic SECURITY.md with community vulnerability reporting process
- **Code Quality Configuration**: golangci-lint with security and quality linters

### Security Tools Integration
- **govulncheck Integration**: Go vulnerability scanner integration (requires working setup)
- **gosec Static Analysis**: Security-focused static analysis tool integration
- **License Compliance**: Basic license scanning to avoid GPL dependencies
- **Hardcoded Secrets Detection**: Pattern-based detection of potential secrets in code

### Enhanced Build System
- **Enhanced Makefile**: New security targets (`security`, `vuln-check`, `gosec`, `mod-verify`)
- **Tool Installation**: Security tool installation helpers

### Documentation Updates
- **Realistic Disclaimers**: Added "vibe coded" disclaimers and warranty sections
- **Accurate Information**: Updated Go version requirements and repository URLs
- **Community Focus**: Aligned documentation with volunteer/community project reality

### Fixed
- **Unhandled Error**: Fixed gosec G104 warning in cmd/root.go by properly handling viper.BindPFlag error
- **Repository Organization**: Moved build artifacts and documentation to organized directories
- **Flag Documentation**: Corrected time flag examples in .env.example (removed incorrect 'd' suffix)

## [1.0.1-beta] - 2025-06-21

### Enhanced
- **Test Coverage**: Achieved good test coverage with comprehensive test scenarios
- **Testing Framework**: Solid test quality with unit, integration, and CLI tests
- **Business Logic Coverage**: Good coverage of core business logic and error paths
- **Edge Case Testing**: Rigorous boundary testing for time precision and API failure scenarios
- **Mock Infrastructure**: Enhanced mock Slack API with complete error simulation capabilities
- **Error Path Validation**: Complete coverage of all error scenarios including API failures and announcement errors
- **Race Condition Safety**: All tests pass with race detection enabled (-race flag)
- **Quality Focus**: Good error test to error code ratio for solid error handling

### Improved
- **Time Boundary Testing**: Added precise boundary condition tests for channel filtering logic
- **API Error Handling**: Enhanced test coverage for GetNewChannels error scenarios
- **Mock State Management**: Improved mock error state clearing and edge case handling
- **Integration Testing**: Added comprehensive workflow testing without artificial shortcuts

## [1.0.0-beta] - 2025-06-21

### Added
- **Beta release** of Slack Buddy AI CLI tool
- **Channel Detection Feature**: Detect new channels created within specified time periods
- **Flexible Time Filtering**: Support for various time formats (24h, 7d, 1w, etc.)
- **Smart Announcements**: Post formatted announcements to designated channels
- **Secure Configuration**: Environment-based token management with `.env` support
- **CLI Framework**: Built with Cobra for professional command structure
- **Slack API Integration**: Full integration with Slack API using official Go SDK
- **Intelligent Error Handling**: Detailed error messages for missing OAuth scopes and permissions
- **User-Friendly Feedback**: Clear guidance on how to fix common configuration issues

### Enhanced
- **Error Messages**: Now shows exactly which OAuth scopes are missing (channels:read, groups:read, chat:write)
- **Authentication Feedback**: Displays connected user and team information
- **Channel Access Validation**: Specific messages for bot membership requirements

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

### Testing
- **Successfully tested** with real Slack workspace (Vibe Coding, Inc.)
- **Channel detection** verified with 4 new channels
- **Announcement feature** confirmed working with #announcements channel
- **Error handling** validated for missing permissions and channel access

### Documentation
- Comprehensive README with installation and usage instructions
- CLAUDE.md for development notes and project tracking
- Inline help documentation for all commands
- TODO.md for task tracking and project management

## [Future Plans]

### Planned Features
- **Bulk Channel Operations**: Mass operations on multiple channels simultaneously  
- **Multi-workspace Support**: Manage multiple Slack workspaces from a single configuration
- **Configurable Message Templates**: Customizable warning and announcement message templates
- **Scheduled Archival Policies**: Cron-like scheduling for automated archival operations

### Planned Enhancements  
- **Enhanced Analytics**: Channel usage statistics and reporting
- **Integration Improvements**: Better error handling and respect for Slack's native rate limiting
- **Performance Optimizations**: Batch API operations for large workspaces