# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.4.0] - 2026-01-18

### Added
- **Warn-Only Mode**: New `--warn-only` flag for archive command
  - Send inactivity warnings without archiving channels
  - Useful for extending grace periods or gradually introducing channel hygiene
  - Warning message does not include archive timeline (softer messaging)
  - Requires `--commit` to actually send warnings (safe dry-run by default)
- **Re-warn Support**: New `--rewarn-days` flag for archive command
  - Re-warn channels whose last warning is older than specified days
  - Default: 0 (disabled, no rewarning)
  - Only applies in warn-only mode
  - Helps refresh stale warnings after a break from running the tool

### Improved
- **Code Quality**: Refactored channel analysis for lower cognitive complexity
  - Extracted `categorizeChannel`, `categorizeChannelWarnOnly`, `categorizeChannelNormal` helpers
  - Added `channelAnalysisParams` struct for cleaner parameter passing
  - Made `#meta` channel reference a constant (`metaChannelText`)

### Documentation
- **README.md**: Added warn-only mode documentation and examples
- **CLAUDE.md**: Added warn-only mode to features list and usage examples

## [1.3.2] - 2025-10-18

### Added
- **Default Channel Diagnostic Mode**: New `--default-channel-check` flag for archive command
  - Shows which users are sampled for default channel detection
  - Displays detected default channels before running archival
  - Helps users verify and tune detection parameters
  - Provides clear guidance on adjusting threshold and sample size
- **Enhanced Test Coverage**: Comprehensive test suite for default channel check functionality
  - 141 new lines of test code covering all diagnostic scenarios
  - Tests for user display, API errors, empty names, and parameter variations
  - 100% coverage of new diagnostic mode code paths

### Improved
- **Code Quality**: Refactored archive command for better maintainability
  - Extracted `validateArchiveDays` helper function for input validation
  - Extracted `resolveArchiveConfig` helper for configuration resolution
  - Cleaner separation of concerns and improved testability
- **Default Channel Detection**: Enhanced user selection logic
  - Clarified that "last N users" refers to most recently joined members
  - Improved documentation explaining Slack API returns users oldest-first
  - More accurate detection by focusing on recent workspace members
- **API Client**: New `GetDefaultChannelsWithUsers` method
  - Returns both default channels and sampled user details
  - Enables diagnostic reporting with user context
  - Better transparency into detection algorithm

### Updated
- **Dependencies**: Updated to latest versions for security and compatibility
  - `github.com/securego/gosec/v2` v2.22.9 → v2.22.10
  - `github.com/anthropics/anthropic-sdk-go` v1.12.0 → v1.13.0
  - `golang.org/x/crypto` v0.42.0 → v0.43.0
  - `golang.org/x/mod` v0.28.0 → v0.29.0
  - `golang.org/x/net` v0.44.0 → v0.46.0
  - `golang.org/x/sys` v0.36.0 → v0.37.0
  - `golang.org/x/tools` v0.37.0 → v0.38.0
  - Various other minor dependency updates

### Documentation
- **CLAUDE.md**: Enhanced default channel protection documentation
  - Added usage examples for `--default-channel-check` flag
  - Clarified user selection algorithm (last N users = most recent)
  - Added verification guidance using diagnostic flag
  - Documented Go module proxy caching behavior for releases
- **README.md**: Will be updated with diagnostic flag documentation

## [1.3.1] - 2025-10-18

### Fixed
- **Version Display**: Fixed `go install` to show correct version number instead of "dev"
  - Uses `runtime/debug.ReadBuildInfo()` to extract version from module metadata
  - Maintains compatibility with `make build` which uses ldflags
  - Updated tests to handle both version patterns

## [1.3.0] - 2025-10-18 [YANKED - use 1.3.1]

**Note**: This version was yanked due to go install version display issue. Use v1.3.1 instead.

### Added
- **Default Channel Protection**: Automatic detection and protection of workspace default channels from archival
  - Intelligent heuristic-based detection using user membership intersection analysis
  - Configurable sample size (default: 10 users) via `--default-channel-sample-size` or `SLACK_DEFAULT_CHANNEL_SAMPLE_SIZE`
  - Adjustable membership threshold (default: 90%) via `--default-channel-threshold` or `SLACK_DEFAULT_CHANNEL_THRESHOLD`
  - Override protection with `--include-default-channels` flag when needed
  - Works across all Slack plans without requiring admin API access
  - Protects critical workspace channels (#general, #announcements, etc.) from accidental archival
- **Enhanced Archive Command**: New configuration options for default channel detection
  - Three new flags for fine-tuning protection behavior
  - Environment variable support for all new configuration options
  - Clear user feedback showing detected default channels and protection status

### Enhanced
- **Code Quality**: Reduced complexity and improved maintainability
  - Extracted helper functions to reduce cognitive complexity below threshold
  - Refactored nested logic into focused, single-purpose functions
  - Improved code organization with better separation of concerns
- **API Resilience**: Centralized rate limit retry logic
  - New `shouldRetryRateLimit()` helper for consistent retry behavior
  - Reduced code duplication across API functions
  - More robust error handling for rate limiting scenarios
- **Test Coverage**: Expanded test suite with comprehensive coverage
  - Added tests for all new helper functions (100% coverage on key helpers)
  - Comprehensive default channel detection tests
  - Enhanced test coverage from 83.4% to 85.2% in cmd package
  - Added 5 new test functions covering edge cases and error paths

### Improved
- **User Experience**: Clear feedback about channel protection
  - Displays detected default channels with count and names
  - Shows protection status and override instructions
  - Separates manual exclusions from auto-detected defaults in output
- **Performance**: Optimized memory usage with pre-allocated slices
- **Code Maintainability**: All linting and complexity checks passing (100% quality gates)
- **Documentation**: Enhanced inline documentation and code comments

## [1.2.1] - 2025-09-27

### Fixed
- **File Upload Detection**: Fixed critical bug where file-only messages (images, documents) were not counted as channel activity
  - Channels with recent file uploads (like memes, screenshots) are now properly recognized as active
  - Prevents incorrect "inactive channel" warnings when users share files without accompanying text
  - Enhanced `isRealMessage` function to check for file attachments (`msg.Files`) in addition to text content
  - Added comprehensive test coverage for file uploads with empty text, minimal text, and multiple files

### Improved
- **Test Coverage**: Added extensive test cases for file upload scenarios to prevent regressions
- **Activity Detection**: More accurate channel activity analysis that considers all forms of user engagement

## [1.2.0] - 2025-08-11

### Enhanced
- **Channel Joining Efficiency**: Major performance improvement to auto-join functionality for inactive channel detection
  - Leverages `IsMember` field from Slack API to skip channels bot is already a member of
  - Reduces API calls dramatically (from 149 join attempts to typically 0-10 actual joins needed)
  - Prevents unnecessary rate limiting during channel analysis operations
  
### Improved  
- **User Experience**: Enhanced auto-join reporting with clearer status messages
  - Shows exactly how many channels need joining vs already member of (e.g., "Joining 5 channels (already member of 144)")
  - Displays "Already member of all X channels, no joining needed" when no joins are required
  - Provides "No channel joining required" confirmation when appropriate
  - Better progress feedback reduces user confusion about join operations

### Fixed
- **API Efficiency**: Eliminated redundant join attempts that were causing rate limit issues
- **Status Reporting**: Fixed confusing messages that showed "Joined 0 channels successfully" for large workspaces

## [1.1.13] - 2025-07-05

### Enhanced
- **Archive Command UX**: Significantly improved user experience with comprehensive status reporting
  - Added configuration status display showing warning/archive thresholds and execution mode at startup
  - Enhanced progress reporting with "[X/Y]" counters for all channel operations
  - Smart time formatting that displays whole numbers cleanly and sub-day intervals with human-readable durations
  - Better dry run summaries with clearer indication of planned actions
- **Default Configuration**: Changed default warn-days from 30 to 45 days for more conservative archival approach

### Improved
- **Progress Tracking**: All channel operations now show detailed progress indicators (e.g., "[1/15] #channel-name")
- **Time Display**: Intelligent formatting that shows "45" instead of "45.0000" for whole numbers, with automatic precision for fractional values
- **Debug Experience**: Enhanced debug logs with channel index and total count context for better troubleshooting
- **Code Organization**: Improved function signatures to pass configuration values through for better reporting

### Fixed
- **Status Reporting**: Enhanced archive command status display with proper parameter passing and cleaner output formatting
- **Test Coverage**: Updated all test files to match enhanced function signatures and improved error handling

## [1.1.12] - 2025-07-01

### Enhanced
- **Channel Archival Messages**: Improved archival message formatting with cleaner, more user-friendly text
  - Removed excessive bold formatting for better readability
  - Enhanced meta channel linking with proper Slack channel link format when available
  - Updated warning messages to be more actionable and less judgmental
  - Added fallback to plain text "#meta" when channel ID lookup fails

### Fixed
- **Message Formatting**: Fixed archival message function signature to accept meta channel ID parameter
- **API Integration**: Enhanced meta channel ID resolution to reduce redundant API calls
- **Error Handling**: Improved error handling in archival process with better fallback behavior

### Improved
- **User Experience**: More helpful and encouraging archival messages
- **Code Quality**: Better separation of concerns in message formatting functions
- **Test Coverage**: Added comprehensive tests for archival message formatting with and without meta channel links

## [1.1.11] - 2025-07-01

### Added
- **Random Channel Highlight Feature**: New `slack-butler channels highlight` command to randomly select and showcase active channels
  - Configurable channel count with `--count` flag (default: 3)
  - Dry run mode by default with `--commit` flag for actual posting
  - Requires `--announce-to` when using `--commit` for safety
  - Comprehensive test coverage with 150+ test cases

### Enhanced
- **Channel Announcements**: Improved announcement message formatting with creation timestamps and clearer time references
  - Enhanced time display showing "days ago" for better context
  - Changed "Purpose" to "Description" for clearer labeling
  - Better singular/plural handling for time references ("1 day" vs "days")
- **Archive Command Display**: Enhanced inactive channel display with days of inactivity calculation
- **User Experience**: Updated Go version requirement from "1.24.4 or later" for clearer compatibility
- **Message Formatting**: Improved announcement and dry-run message consistency across all commands

### Fixed
- **Test Message Expectations**: Updated test cases to match new announcement message formats
- **Duplicate Detection**: Enhanced duplicate announcement detection with updated message patterns

### Improved
- **Code Organization**: Better separation of announcement formatting between commit and dry-run modes
- **Test Coverage**: Expanded test suite with comprehensive highlight command testing
- **Documentation**: Added random channel highlight feature to roadmap

## [1.1.10] - 2025-06-27

### Enhanced
- **Archive Command Testing**: Added comprehensive test coverage for --commit flag validation requiring --announce-to parameter
- **User Experience**: Improved archive command feedback with clearer progress messages for channel joining
- **Code Quality**: Reduced log noise by converting Info-level logs to Debug-level in channel analysis functions

### Fixed
- **Archive Scope**: Restricted archive operations to public channels only for improved safety and performance
- **Command Validation**: Added proper validation ensuring --announce-to is required when using --commit flag
- **Logging Consistency**: Fixed excessive INFO logging during channel analysis operations

### Improved
- **Archive Performance**: Optimized channel filtering to focus on public channels only
- **User Feedback**: Enhanced progress messaging during channel joining operations
- **Code Organization**: Better separation between debug and user-facing output

## [1.1.9] - 2025-06-27

### Enhanced
- **Test Coverage**: Major expansion of test suite with comprehensive `isRealMessage` testing covering all message types, system events, and edge cases
- **Documentation Quality**: Removed development disclaimer and improved Go version specification (Go 1.23+ with Go 1.24.4 testing)
- **User Experience**: Enhanced test coverage with comprehensive testing framework

### Fixed  
- **Test Accuracy**: Fixed error message expectations in archive tests (warn-seconds → warn-days)
- **Health Testing**: Added comprehensive health command test coverage
- **Integration Testing**: Enhanced integration test reliability and coverage

### Removed
- **Build Dependencies**: Completely removed GoReleaser configuration and dependencies for simplified release process

### Improved
- **Build System**: Streamlined Makefile with improved organization and documentation
- **Code Quality**: Enhanced testing framework with better mock coverage and edge case validation
- **Project Maintenance**: Simplified dependency management and build process

## [1.1.8] - 2025-06-26

### Fixed
- **Documentation**: Minor documentation consistency improvements and cleanup
- **Version References**: Ensured all documentation files accurately reflect current stable release status

### Improved
- **Documentation Quality**: Maintained accurate and up-to-date project documentation

## [1.1.7] - 2025-06-26

### Enhanced
- **Command Usability**: Switched archive command from seconds to days with decimal precision support for more intuitive time configuration
- **Documentation Clarity**: Simplified project disclaimer in README for better readability
- **Project Branding**: Fixed remaining references from 'Slack Buddy AI' to 'Slack Butler' in documentation

### Improved
- **User Experience**: Archive command now accepts practical day-based timing (e.g., --warn-days=7.5 for 7.5 days)
- **Documentation Consistency**: All documentation now consistently uses the correct project name

## [1.1.6] - 2025-06-26

### Fixed
- **OAuth Scope Validation**: Added missing channels:history scope to health check validation to match README requirements
- **Documentation Consistency**: Synchronized OAuth scope requirements between README.md and health check command

### Improved
- **Setup Requirements**: Simplified OAuth scope documentation by removing redundant (required) annotations
- **Scope Coverage**: Health check now validates all 6 required scopes: channels:read, channels:join, channels:manage, channels:history, chat:write, users:read
- **Documentation Clarity**: Updated channels:join description to include message checks and announcements

### Removed
- **Optional Scopes**: Removed groups:read from OAuth requirements as it's not needed for core functionality

## [1.1.5] - 2025-06-26

### Fixed
- **Code Quality**: Fixed gofmt -s formatting issues in cmd/root.go to ensure 100% quality compliance
- **Documentation**: Updated README.md for better clarity and consistency
  - Changed "basic channel detection" to "new channel detection" 
  - Changed "channel archival management" to "warning and archiving inactive channels"
  - Updated archive examples to use days for practical usage
  - Removed redundant sections and duplicate OAuth scope documentation
  - Added security warning about --token flag exposure in shell history
  - Removed "bulk channel operations" from roadmap (not planned)

### Added
- **Roadmap**: Added new channel detection as completed feature in roadmap

## [1.1.4] - 2025-06-26

### Fixed
- **Module Path**: Fixed go install compatibility after repository rename from slack-buddy-ai to slack-butler
- **Documentation**: Updated all version references to maintain consistency across documentation files
- **Repository Cleanup**: Removed coverage artifacts from version control

### Added
- **CLI Enhancement**: Added generative AI development acknowledgment to help output

## [1.1.3] - 2025-06-25

### Fixed
- **Documentation**: Fixed incorrect flag names in .env.example archive command example (--warn-seconds/--archive-seconds → --warn-days/--archive-days)

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
- **Beta release** of Slack Butler CLI tool
- **Channel Detection Feature**: Detect new channels created within specified time periods
- **Flexible Time Filtering**: Support for various time formats (24h, 7d, 1w, etc.)
- **Smart Announcements**: Post formatted announcements to designated channels
- **Secure Configuration**: Environment-based token management with `.env` support
- **CLI Framework**: Built with Cobra for professional command structure
- **Slack API Integration**: Full integration with Slack API using official Go SDK
- **Intelligent Error Handling**: Detailed error messages for missing OAuth scopes and permissions
- **User-Friendly Feedback**: Clear guidance on how to fix common configuration issues

### Enhanced
- **Error Messages**: Now shows exactly which OAuth scopes are missing (channels:read, chat:write)
- **Authentication Feedback**: Displays connected user and team information
- **Channel Access Validation**: Specific messages for bot membership requirements

### Features
- `slack-butler channels detect` command with the following options:
  - `--since` flag for time period specification (default: 24h)
  - `--announce-to` flag for target announcement channel
  - `--token` flag for direct token specification
- Environment variable support via `SLACK_TOKEN`
- Rich message formatting for channel announcements
- Error handling for API failures and authentication issues

### Technical Details
- Go module: `slack-butler`
- CLI tool name: `slack-butler`
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

