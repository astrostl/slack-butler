package slack

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateSlackToken performs basic validation on Slack bot tokens
func ValidateSlackToken(token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Allow test tokens for testing purposes
	if strings.HasPrefix(token, "MOCK-") || strings.Contains(token, "TESTING-ONLY") {
		return nil
	}

	// Slack bot tokens should start with "xoxb-"
	if !strings.HasPrefix(token, "xoxb-") {
		return fmt.Errorf("invalid token format: bot tokens must start with 'xoxb-'")
	}

	// Basic format validation (xoxb-XXXXXXXXXXXX-XXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXX)
	botTokenPattern := regexp.MustCompile(`^xoxb-\d+-\d+-[a-zA-Z0-9]+$`)
	if !botTokenPattern.MatchString(token) {
		return fmt.Errorf("invalid token format: token does not match expected Slack bot token pattern")
	}

	// Check for minimum length (Slack tokens are typically much longer)
	if len(token) < 50 {
		return fmt.Errorf("invalid token: token appears too short")
	}

	return nil
}

// SanitizeForLogging removes sensitive information from strings for safe logging
func SanitizeForLogging(input string) string {
	// Replace any token-like patterns with [REDACTED]
	tokenPattern := regexp.MustCompile(`xoxb-[a-zA-Z0-9-]+`)
	result := tokenPattern.ReplaceAllString(input, "[REDACTED]")

	// Also redact test tokens
	testTokenPattern := regexp.MustCompile(`MOCK-[A-Z0-9-]+`)
	result = testTokenPattern.ReplaceAllString(result, "[REDACTED]")

	return result
}

// ValidateChannelName performs basic validation on channel names
func ValidateChannelName(channelName string) error {
	if channelName == "" {
		return fmt.Errorf("channel name cannot be empty")
	}

	// Remove # prefix for validation
	name := strings.TrimPrefix(channelName, "#")

	// Check if name is empty after removing prefix
	if name == "" {
		return fmt.Errorf("channel name cannot be empty")
	}

	// Basic channel name validation (alphanumeric, hyphens, underscores)
	channelPattern := regexp.MustCompile(`^[a-z0-9_-]+$`)
	if !channelPattern.MatchString(name) {
		return fmt.Errorf("invalid channel name: must contain only lowercase letters, numbers, hyphens, and underscores")
	}

	if len(name) > 80 {
		return fmt.Errorf("invalid channel name: too long (max 80 characters)")
	}

	return nil
}
