package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSlackToken(t *testing.T) {
	t.Run("Valid bot token", func(t *testing.T) {
		validToken := "xoxb-test-token-for-validation-only"
		err := ValidateSlackToken(validToken)
		assert.NoError(t, err)
	})

	t.Run("Empty token", func(t *testing.T) {
		err := ValidateSlackToken("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token cannot be empty")
	})

	t.Run("Invalid prefix", func(t *testing.T) {
		err := ValidateSlackToken("xoxa-1234567890123-1234567890123-abcdefghijklmnopqrstuvwxyz123456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must start with 'xoxb-'")
	})

	t.Run("Invalid format", func(t *testing.T) {
		err := ValidateSlackToken("xoxb-invalid-format")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not match expected")
	})

	t.Run("Token too short", func(t *testing.T) {
		err := ValidateSlackToken("xoxb-123-456-abc")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "appears too short")
	})
}

func TestSanitizeForLogging(t *testing.T) {
	t.Run("Replace token in string", func(t *testing.T) {
		input := "Error with token xoxb-test-token-for-validation-only"
		result := SanitizeForLogging(input)
		assert.Equal(t, "Error with token [REDACTED]", result)
		assert.NotContains(t, result, "xoxb-")
	})

	t.Run("No token in string", func(t *testing.T) {
		input := "This is a normal log message"
		result := SanitizeForLogging(input)
		assert.Equal(t, input, result)
	})

	t.Run("Multiple tokens", func(t *testing.T) {
		input := "Token1: xoxb-111-222-abc Token2: xoxb-333-444-def"
		result := SanitizeForLogging(input)
		assert.Equal(t, "Token1: [REDACTED] Token2: [REDACTED]", result)
		assert.NotContains(t, result, "xoxb-")
	})
}

func TestValidateChannelName(t *testing.T) {
	t.Run("Valid channel with #", func(t *testing.T) {
		err := ValidateChannelName("#general")
		assert.NoError(t, err)
	})

	t.Run("Valid channel without #", func(t *testing.T) {
		err := ValidateChannelName("general")
		assert.NoError(t, err)
	})

	t.Run("Valid channel with hyphens", func(t *testing.T) {
		err := ValidateChannelName("#team-updates")
		assert.NoError(t, err)
	})

	t.Run("Valid channel with underscores", func(t *testing.T) {
		err := ValidateChannelName("dev_team")
		assert.NoError(t, err)
	})

	t.Run("Empty channel name", func(t *testing.T) {
		err := ValidateChannelName("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("Only # character", func(t *testing.T) {
		err := ValidateChannelName("#")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("Invalid characters", func(t *testing.T) {
		err := ValidateChannelName("#Invalid Channel!")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must contain only lowercase")
	})

	t.Run("Uppercase characters", func(t *testing.T) {
		err := ValidateChannelName("#General")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must contain only lowercase")
	})

	t.Run("Too long channel name", func(t *testing.T) {
		longName := "#" + string(make([]byte, 81))
		for i := range longName[1:] {
			longName = longName[:i+1] + "a" + longName[i+2:]
		}
		err := ValidateChannelName(longName)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too long")
	})
}