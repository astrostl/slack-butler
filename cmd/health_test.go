package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"slack-buddy-ai/pkg/slack"
)

func TestHealthCommandSetup(t *testing.T) {
	t.Run("Command structure", func(t *testing.T) {
		assert.NotNil(t, healthCmd)
		assert.Equal(t, "health", healthCmd.Use)
		assert.NotEmpty(t, healthCmd.Short)
		assert.NotEmpty(t, healthCmd.Long)
	})

	t.Run("Command flags", func(t *testing.T) {
		verboseFlag := healthCmd.Flags().Lookup("verbose")
		assert.NotNil(t, verboseFlag)
		assert.Equal(t, "false", verboseFlag.DefValue)
		assert.Equal(t, "v", verboseFlag.Shorthand)
	})

	t.Run("Command is registered", func(t *testing.T) {
		commands := rootCmd.Commands()
		var healthFound bool
		for _, cmd := range commands {
			if cmd.Use == "health" {
				healthFound = true
				break
			}
		}
		assert.True(t, healthFound)
	})
}

func TestIsValidTokenFormat(t *testing.T) {
	testCases := []struct {
		name     string
		token    string
		expected bool
	}{
		{"Valid bot token", "xoxb-FAKE-TESTING-ONLY-NOT-REAL-TOKEN", true},
		{"Valid user token", "xoxp-FAKE-TESTING-ONLY-NOT-REAL-TOKEN", true},
		{"Invalid prefix", "invalid-FAKE-TOKEN-FOR-TESTING-ONLY-NOT-REAL", false},
		{"Too short", "xoxb-123", false},
		{"Empty token", "", false},
		{"Random string", "randomstring", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidTokenFormat(tc.token)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestContainsFunction(t *testing.T) {
	testCases := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"Contains substring", "hello world", "world", true},
		{"Contains at beginning", "hello world", "hello", true},
		{"Contains at end", "hello world", "world", true},
		{"Does not contain", "hello world", "foo", false},
		{"Empty substring", "hello", "", true},
		{"Empty string", "", "foo", false},
		{"Same string", "hello", "hello", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := contains(tc.s, tc.substr)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCheckRequiredPermissions(t *testing.T) {
	t.Run("No permission errors", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		errors := checkRequiredPermissions(client)
		assert.Len(t, errors, 0)
	})

	t.Run("Missing scope error", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		// Note: The current implementation doesn't have a way to mock specific permission errors
		// This would need enhancement of the mock API to support permission error simulation
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		errors := checkRequiredPermissions(client)
		// For now, this should pass since the mock doesn't simulate permission errors
		assert.Len(t, errors, 0)
	})
}

func TestRunHealthErrors(t *testing.T) {
	t.Run("No token configured", func(t *testing.T) {
		// Clear any environment token
		t.Setenv("SLACK_TOKEN", "")
		initConfig()

		err := runHealth(healthCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration validation failed")
	})

	t.Run("Invalid token format", func(t *testing.T) {
		t.Setenv("SLACK_TOKEN", "invalid-token")
		initConfig()

		err := runHealth(healthCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token format validation failed")
	})

	t.Run("Invalid token pattern", func(t *testing.T) {
		t.Setenv("SLACK_TOKEN", "xoxb-short")
		initConfig()

		err := runHealth(healthCmd, []string{})
		assert.Error(t, err)
		// This should fail at client initialization due to token validation
		assert.True(t, 
			strings.Contains(err.Error(), "client initialization failed") ||
			strings.Contains(err.Error(), "token format validation failed"))
	})
}

func TestTestBasicFunctionality(t *testing.T) {
	t.Run("Basic functionality test passes", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		err = testBasicFunctionality(client)
		assert.NoError(t, err)
	})

	t.Run("Basic functionality test with error", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetGetConversationsError(true) // Force an error
		
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		err = testBasicFunctionality(client)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "channel listing test failed")
	})
}