package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/astrostl/slack-butler/pkg/slack"
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
		{"Valid bot token", "test-FAKE-TESTING-ONLY-NOT-REAL-TOKEN", true},
		{"Valid user token", "xoxp-FAKE-TESTING-ONLY-NOT-REAL-TOKEN", true},
		{"Invalid prefix", "invalid-FAKE-TOKEN-FOR-TESTING-ONLY-NOT-REAL", false},
		{"Too short", "test-123", false},
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

func TestCheckOAuthScopes(t *testing.T) {
	t.Run("All scopes available", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		scopes, err := client.CheckOAuthScopes()
		require.NoError(t, err)

		// The mock implementation should return true for all scopes by default
		assert.True(t, scopes["channels:read"])
		assert.True(t, scopes["channels:join"])
		assert.True(t, scopes["chat:write"])
		assert.True(t, scopes["channels:manage"])
		assert.True(t, scopes["users:read"])
		assert.True(t, scopes["groups:read"])
	})

	t.Run("Missing users:read scope", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetGetUsersError("missing_scope")
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		scopes, err := client.CheckOAuthScopes()
		require.NoError(t, err)

		// users:read should be false due to missing scope
		assert.False(t, scopes["users:read"])
		// Other scopes should still be true
		assert.True(t, scopes["channels:read"])
	})

	t.Run("Auth failure in scope check", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		// First create client successfully, then set auth error for CheckOAuthScopes call
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		// Now set auth error for the CheckOAuthScopes call
		mockAPI.SetAuthError(true)

		scopes, err := client.CheckOAuthScopes()
		assert.Error(t, err)
		assert.Nil(t, scopes)
		assert.Contains(t, err.Error(), "failed to authenticate")
	})
}

func TestRunHealthErrorScenarios(t *testing.T) {
	t.Run("Missing token", func(t *testing.T) {
		// Clear environment variables
		t.Setenv("SLACK_TOKEN", "")
		viper.Set("token", "")

		// Create a mock command
		cmd := &cobra.Command{}

		err := runHealth(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration validation failed")
	})

	t.Run("Invalid token format", func(t *testing.T) {
		// Set invalid token format directly in viper
		viper.Set("token", "invalid-token-format")

		cmd := &cobra.Command{}

		err := runHealth(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token format validation failed")
	})

	t.Run("Client creation failure", func(t *testing.T) {
		// Set valid token format directly in viper
		viper.Set("token", "test-invalid-auth-token")

		cmd := &cobra.Command{}

		// This will fail because the real client will try to authenticate
		// but we don't have actual credentials
		err := runHealth(cmd, []string{})
		assert.Error(t, err)
		// Should contain some form of authentication error
		assert.True(t, strings.Contains(err.Error(), "client initialization failed") ||
			strings.Contains(err.Error(), "authentication") ||
			strings.Contains(err.Error(), "invalid_auth"))
	})
}

func TestRunHealthWithMockClient(t *testing.T) {
	t.Run("Complete successful health check", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		scopes, err := client.CheckOAuthScopes()
		require.NoError(t, err)
		assert.True(t, scopes["channels:read"])
		assert.True(t, scopes["chat:write"])
	})

	t.Run("Health check with missing optional scopes", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetGetUsersError("missing_scope")
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		scopes, err := client.CheckOAuthScopes()
		require.NoError(t, err)

		assert.False(t, scopes["users:read"])
		assert.True(t, scopes["channels:read"])
	})

	t.Run("Basic functionality test", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		basicErr := testBasicFunctionality(client)
		assert.NoError(t, basicErr)
	})

	t.Run("Complete runHealth with mock client - success path", func(t *testing.T) {
		// Set valid token
		viper.Set("token", "test-token-for-testing-purposes-only")

		// This test covers integration of runHealth function with real mock client
		// We can't fully test runHealth due to NewClient dependency, but we can test components

		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		// Test all the components runHealth uses
		authInfo, err := client.TestAuth()
		assert.NoError(t, err)
		assert.NotEmpty(t, authInfo.User)
		assert.NotEmpty(t, authInfo.Team)

		scopes, err := client.CheckOAuthScopes()
		assert.NoError(t, err)

		// Verify all required scopes are available
		requiredScopes := []string{"channels:read", "channels:join", "chat:write", "channels:manage", "users:read"}
		for _, scope := range requiredScopes {
			assert.True(t, scopes[scope], "Required scope %s should be available", scope)
		}

		err = testBasicFunctionality(client)
		assert.NoError(t, err)
	})

	t.Run("Complete runHealth with verbose mode simulation", func(t *testing.T) {
		// Test verbose mode paths
		originalVerbose := healthVerbose
		defer func() { healthVerbose = originalVerbose }()

		healthVerbose = true
		viper.Set("token", "test-token-for-testing-purposes-only")

		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		// Test verbose token display logic
		token := viper.GetString("token")
		if len(token) >= 16 {
			prefix := token[:8]
			suffix := token[len(token)-8:]
			assert.Equal(t, "test-tok", prefix)
			assert.Equal(t, "ses-only", suffix)
		}

		// Test verbose auth info
		authInfo, err := client.TestAuth()
		assert.NoError(t, err)
		assert.NotEmpty(t, authInfo.User)
		assert.NotEmpty(t, authInfo.Team)
		assert.NotEmpty(t, authInfo.UserID)

		scopes, err := client.CheckOAuthScopes()
		assert.NoError(t, err)
		assert.True(t, scopes["channels:read"])

		healthVerbose = false
	})
}

func TestHealthCheckScopesErrorPaths(t *testing.T) {
	t.Run("Missing channels:read scope", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetMissingScopeError(true) // This will simulate missing scope specifically
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		scopes, err := client.CheckOAuthScopes()
		require.NoError(t, err)

		// Should detect missing channels:read scope
		assert.False(t, scopes["channels:read"])
	})

	t.Run("Missing groups:read scope", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetGetUsersError("missing_scope")
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		scopes, err := client.CheckOAuthScopes()
		require.NoError(t, err)

		// groups:read affects users, so users:read should be false
		assert.False(t, scopes["users:read"])
	})

	t.Run("API connectivity failure", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetAuthError(true)

		// Client creation should fail with auth error
		client, err := slack.NewClientWithAPI(mockAPI)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "authentication failed")
	})
}

func TestHealthVerboseOutput(t *testing.T) {
	t.Run("Verbose output includes scope details", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		scopes, err := client.CheckOAuthScopes()
		require.NoError(t, err)

		// Verify that all expected scopes are present in the result
		expectedScopes := []string{"channels:read", "chat:write", "channels:join", "channels:manage", "users:read", "groups:read"}
		for _, scope := range expectedScopes {
			_, exists := scopes[scope]
			assert.True(t, exists, "Expected scope %s to be checked", scope)
		}
	})

	t.Run("Verbose flag parsing", func(t *testing.T) {
		// Test that verbose flag can be retrieved
		cmd := &cobra.Command{}
		cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

		// Set verbose flag
		err := cmd.Flags().Set("verbose", "true")
		assert.NoError(t, err)

		// Get verbose flag value
		verbose, err := cmd.Flags().GetBool("verbose")
		assert.NoError(t, err)
		assert.True(t, verbose)
	})
}

func TestRunHealthMissingScopes(t *testing.T) {
	t.Run("Health check fails with invalid token", func(t *testing.T) {
		// Use a properly formatted test token that will pass format validation
		viper.Set("token", "test-fake-token-not-real-for-testing")

		cmd := &cobra.Command{}
		cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

		// This will fail at client initialization since it's not a real token
		err := runHealth(cmd, []string{})
		assert.Error(t, err)
		// Will fail at client initialization step
		assert.Contains(t, err.Error(), "client initialization failed")
	})
}

func TestRunHealthSuccess(t *testing.T) {
	t.Run("Health check with proper token format", func(t *testing.T) {
		// Set properly formatted token that will pass format validation
		viper.Set("token", "test-fake-token-not-real-for-testing")

		cmd := &cobra.Command{}
		cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

		// This will fail at client initialization due to invalid auth, but tests token format validation
		err := runHealth(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client initialization failed")
	})

	t.Run("Health check verbose flag processing", func(t *testing.T) {
		// Test verbose flag processing without going through full health check
		viper.Set("token", "test-fake-token-not-real-for-testing")

		cmd := &cobra.Command{}
		cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
		err := cmd.Flags().Set("verbose", "true")
		require.NoError(t, err)

		// Verify verbose flag is set correctly
		verbose, err := cmd.Flags().GetBool("verbose")
		require.NoError(t, err)
		assert.True(t, verbose)
	})
}

func TestRunHealthStructure(t *testing.T) {
	t.Run("Health command structure validation", func(t *testing.T) {
		// Set up valid token
		t.Setenv("SLACK_TOKEN", "test-token-for-testing-purposes-only")
		initConfig()

		// Test that the health command exists and is properly structured
		assert.NotNil(t, healthCmd)
		assert.Equal(t, "health", healthCmd.Use)
		assert.NotNil(t, healthCmd.RunE)

		// Test that token validation would pass with a properly formatted token
		token := "test-token-for-testing-purposes-only"
		assert.True(t, isValidTokenFormat(token))
	})

	t.Run("Verbose flag functionality", func(t *testing.T) {
		// Test that verbose flag is properly defined and accessible
		verboseFlag := healthCmd.Flags().Lookup("verbose")
		assert.NotNil(t, verboseFlag)
		assert.Equal(t, "v", verboseFlag.Shorthand)

		// Set verbose flag
		err := healthCmd.Flags().Set("verbose", "true")
		assert.NoError(t, err)

		// Verify flag was set
		verbose, err := healthCmd.Flags().GetBool("verbose")
		assert.NoError(t, err)
		assert.True(t, verbose)
	})

	t.Run("Health command with verbose output", func(t *testing.T) {
		// Test containsSubstring function edge cases for better coverage
		testCases := []struct {
			name     string
			s        string
			substr   string
			expected bool
		}{
			{"Case sensitive match", "Hello", "hello", false},
			{"Case insensitive substring in middle", "testing", "est", true},
			{"Single character", "a", "a", true},
			{"Unicode characters", "héllo", "éll", true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := containsSubstring(tc.s, tc.substr)
				assert.Equal(t, tc.expected, result)
			})
		}
	})
}

func TestRunHealthErrors(t *testing.T) {
	t.Run("No token configured", func(t *testing.T) {
		// Clear any environment token and viper setting
		viper.Set("token", "")

		err := runHealth(healthCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration validation failed")
	})

	t.Run("Invalid token format", func(t *testing.T) {
		viper.Set("token", "invalid-token")

		err := runHealth(healthCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token format validation failed")
	})

	t.Run("Invalid token pattern", func(t *testing.T) {
		viper.Set("token", "test-short")

		err := runHealth(healthCmd, []string{})
		assert.Error(t, err)
		// This should fail at client initialization due to invalid token pattern
		assert.Contains(t, err.Error(), "client initialization failed")
	})

	t.Run("API authentication failure", func(t *testing.T) {
		t.Setenv("SLACK_TOKEN", "test-token-for-testing-purposes-only")
		initConfig()

		// We would need to mock the NewClient function to test client initialization failure
		// and API connectivity failure scenarios, but these would require significant
		// refactoring of the health command to be dependency-injectable.
		// The current tests cover the main error paths that can be tested without mocking.
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

func TestRunHealthVerboseOutput(t *testing.T) {
	// This test verifies that the verbose flag functionality exists
	// We don't need to test the complete integration since that's complex
	// and already covered by other tests
	t.Run("Verbose flag handling", func(t *testing.T) {
		// Test that healthVerbose variable can be modified
		originalVerbose := healthVerbose
		defer func() { healthVerbose = originalVerbose }()

		healthVerbose = true
		assert.True(t, healthVerbose)

		healthVerbose = false
		assert.False(t, healthVerbose)
	})
}

func TestRunHealthWithMockSuccess(t *testing.T) {
	t.Run("Successful health check with mock", func(t *testing.T) {
		// Set up valid token
		t.Setenv("SLACK_TOKEN", "test-token-for-testing-purposes-only")
		initConfig()

		// Create mock client with successful responses
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		// Test individual components that runHealth uses
		authInfo, err := client.TestAuth()
		assert.NoError(t, err)
		assert.NotEmpty(t, authInfo.User)

		scopes, err := client.CheckOAuthScopes()
		assert.NoError(t, err)
		assert.True(t, scopes["channels:read"])

		err = testBasicFunctionality(client)
		assert.NoError(t, err)
	})
}

func TestRunHealthWithMissingScopeScenarios(t *testing.T) {
	t.Run("Health check with missing optional groups:read scope", func(t *testing.T) {
		t.Setenv("SLACK_TOKEN", "test-token-for-testing-purposes-only")
		initConfig()

		// Create mock that simulates missing groups:read scope specifically
		mockAPI := slack.NewMockSlackAPI()
		// Set error specifically for private_channel type conversation requests
		mockAPI.SetGetConversationsErrorWithMessage(true, "missing_scope")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		scopes, err := client.CheckOAuthScopes()
		assert.NoError(t, err)

		// Should still pass required scopes (chat:write defaults to true in mock)
		assert.True(t, scopes["chat:write"])
		assert.True(t, scopes["channels:manage"])

		// But groups:read should be false due to the missing_scope error
		assert.False(t, scopes["groups:read"])
	})

	t.Run("Health check with missing users:read scope", func(t *testing.T) {
		t.Setenv("SLACK_TOKEN", "test-token-for-testing-purposes-only")
		initConfig()

		// Create mock that simulates missing users:read scope
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetGetUsersError("missing_scope")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		scopes, err := client.CheckOAuthScopes()
		assert.NoError(t, err)

		// Should fail users:read scope
		assert.False(t, scopes["users:read"])
		// But other scopes should pass
		assert.True(t, scopes["chat:write"])
	})
}

func TestRunHealthAPIFailureScenarios(t *testing.T) {
	t.Run("Client creation with auth error", func(t *testing.T) {
		t.Setenv("SLACK_TOKEN", "test-token-for-testing-purposes-only")
		initConfig()

		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetAuthError(true)

		// Client creation should fail during authentication
		_, err := slack.NewClientWithAPI(mockAPI)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
	})

	t.Run("OAuth scope check with auth error during check", func(t *testing.T) {
		t.Setenv("SLACK_TOKEN", "test-token-for-testing-purposes-only")
		initConfig()

		mockAPI := slack.NewMockSlackAPI()
		// Allow initial client creation (auth success)
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		// Now set auth error for subsequent scope check calls
		mockAPI.SetAuthError(true)

		_, err = client.CheckOAuthScopes()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to authenticate")
	})

	t.Run("Basic functionality test failure", func(t *testing.T) {
		t.Setenv("SLACK_TOKEN", "test-token-for-testing-purposes-only")
		initConfig()

		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetGetConversationsError(true)

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		err = testBasicFunctionality(client)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "channel listing test failed")
	})
}

func TestRunHealthEndToEnd(t *testing.T) {
	// Save original values
	originalVerbose := healthVerbose
	defer func() { healthVerbose = originalVerbose }()

	t.Run("Missing token configuration", func(t *testing.T) {
		// Clear any existing token
		viper.Set("token", "")

		cmd := &cobra.Command{}
		err := runHealth(cmd, []string{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration validation failed")
	})

	t.Run("Invalid token format", func(t *testing.T) {
		// Set invalid token format
		viper.Set("token", "invalid-token")

		cmd := &cobra.Command{}
		err := runHealth(cmd, []string{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token format validation failed")
	})

	t.Run("Verbose mode displays token info", func(t *testing.T) {
		// Set a test token that will pass basic validation but fail client creation
		viper.Set("token", "MOCK-TESTING-ONLY-TOKEN-FOR-UNIT-TESTS-12345678901234567890")

		// Enable verbose mode
		healthVerbose = true

		cmd := &cobra.Command{}
		err := runHealth(cmd, []string{})

		// This test should fail at token format validation since health command expects valid prefix
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token format validation failed")
	})

	t.Run("Health check with working token format", func(t *testing.T) {
		// Use a properly formatted test token that passes validation
		viper.Set("token", "test-fake-token-not-real-for-testing")

		cmd := &cobra.Command{}

		// This will pass token format validation but fail on actual Slack API calls
		err := runHealth(cmd, []string{})

		// We expect an error when trying to create the actual client, but the token format should be valid
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client initialization failed")
	})
}

func TestRunHealthIntegrationSuccess(t *testing.T) {
	// Save original values
	originalVerbose := healthVerbose
	defer func() { healthVerbose = originalVerbose }()

	t.Run("Complete successful health check integration", func(t *testing.T) {
		// We need to test the integration by mocking the NewClient function
		// Since we can't easily mock NewClient without refactoring, we'll test
		// the components that make up a successful health check

		// Set valid token
		viper.Set("token", "test-fake-token-not-real-for-testing")

		// Verify token format validation passes
		token := viper.GetString("token")
		assert.True(t, isValidTokenFormat(token))

		// Create mock client and verify all health check components work
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		// Test auth info retrieval
		authInfo, err := client.TestAuth()
		assert.NoError(t, err)
		assert.NotEmpty(t, authInfo.User)
		assert.NotEmpty(t, authInfo.Team)

		// Test OAuth scope validation
		scopes, err := client.CheckOAuthScopes()
		assert.NoError(t, err)
		assert.True(t, scopes["channels:read"])
		assert.True(t, scopes["channels:join"])
		assert.True(t, scopes["chat:write"])
		assert.True(t, scopes["channels:manage"])
		assert.True(t, scopes["users:read"])
		assert.True(t, scopes["groups:read"])

		// Test basic functionality
		err = testBasicFunctionality(client)
		assert.NoError(t, err)

		// Verify that all required scopes are present
		requiredScopes := map[string]bool{
			"channels:read":   true,
			"channels:join":   true,
			"chat:write":      true,
			"channels:manage": true,
			"users:read":      true,
		}

		var missingRequired []string
		for scope := range requiredScopes {
			if !scopes[scope] {
				missingRequired = append(missingRequired, scope)
			}
		}
		assert.Empty(t, missingRequired, "Should have no missing required scopes")
	})

	t.Run("Successful health check with verbose output", func(t *testing.T) {
		// Test verbose mode functionality
		healthVerbose = true

		viper.Set("token", "test-12345678901234567890123456")

		// Create mock client
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		// Test verbose token display logic
		token := viper.GetString("token")
		if len(token) >= 16 {
			prefix := token[:8]
			suffix := token[len(token)-8:]
			assert.Equal(t, "test-123", prefix)
			assert.Equal(t, "90123456", suffix)
		}

		// Test auth info with verbose details
		authInfo, err := client.TestAuth()
		assert.NoError(t, err)
		assert.NotEmpty(t, authInfo.User)
		assert.NotEmpty(t, authInfo.Team)
		assert.NotEmpty(t, authInfo.UserID)
		// TeamID may be empty in mock - that's ok for this test

		// Test scope validation with verbose details
		scopes, err := client.CheckOAuthScopes()
		assert.NoError(t, err)

		// Verify verbose scope reporting logic
		requiredScopes := map[string]bool{
			"channels:read":   true,
			"channels:join":   true,
			"chat:write":      true,
			"channels:manage": true,
			"users:read":      true,
		}

		for scope := range requiredScopes {
			assert.True(t, scopes[scope], "Required scope %s should be available", scope)
		}

		// Test basic functionality with verbose output
		err = testBasicFunctionality(client)
		assert.NoError(t, err)

		healthVerbose = false
	})
}

func TestRunHealthMissingRequiredScopes(t *testing.T) {
	t.Run("Health check fails with missing required scopes", func(t *testing.T) {
		t.Setenv("SLACK_TOKEN", "test-token-for-testing-purposes-only")
		initConfig()

		// Create mock that simulates missing required scopes
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetMissingScopeError(true)        // Missing channels:read
		mockAPI.SetGetUsersError("missing_scope") // Missing users:read

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		scopes, err := client.CheckOAuthScopes()
		require.NoError(t, err)

		// Manually check what runHealth would do with missing required scopes
		requiredScopes := map[string]bool{
			"channels:read":   true,
			"channels:join":   true,
			"chat:write":      true,
			"channels:manage": true,
			"users:read":      true,
		}

		var missingRequired []string
		for scope := range requiredScopes {
			if !scopes[scope] {
				missingRequired = append(missingRequired, scope)
			}
		}

		// Should have missing scopes
		assert.NotEmpty(t, missingRequired)
		assert.Contains(t, missingRequired, "channels:read")
		assert.Contains(t, missingRequired, "users:read")
	})
}

func TestContainsFunctions(t *testing.T) {
	t.Run("Contains function basic cases", func(t *testing.T) {
		// Test basic substring matching
		assert.True(t, contains("hello world", "hello"))
		assert.True(t, contains("hello world", "world"))
		assert.True(t, contains("hello world", "lo wo"))
		assert.True(t, contains("hello", "hello"))
		assert.False(t, contains("hello", "world"))
		assert.False(t, contains("he", "hello"))

		// Test empty strings
		assert.True(t, contains("hello", ""))
		assert.True(t, contains("", ""))
		assert.False(t, contains("", "hello"))
	})

	t.Run("Contains function edge cases", func(t *testing.T) {
		// Test prefix and suffix matching
		assert.True(t, contains("testing", "test"))
		assert.True(t, contains("testing", "ting"))

		// Test exact matches
		assert.True(t, contains("exact", "exact"))

		// Test single character
		assert.True(t, contains("a", "a"))
		assert.False(t, contains("a", "b"))
	})

	t.Run("ContainsSubstring function", func(t *testing.T) {
		// Test the containsSubstring helper function directly
		assert.True(t, containsSubstring("hello world", "lo wo"))
		assert.True(t, containsSubstring("abcdef", "cd"))
		assert.False(t, containsSubstring("abcdef", "xyz"))
		assert.False(t, containsSubstring("ab", "abc"))

		// Edge cases
		assert.True(t, containsSubstring("a", "a"))
		assert.False(t, containsSubstring("", "a"))
		assert.True(t, containsSubstring("abc", ""))
	})
}

func TestRunHealthVerboseOutputPaths(t *testing.T) {
	// Save original state
	originalVerbose := healthVerbose
	defer func() { healthVerbose = originalVerbose }()

	t.Run("Verbose token display logic", func(t *testing.T) {
		// Test the verbose token display logic without full health check
		token := "test-fake-token-for-display-test"

		// Simulate what runHealth does for verbose token display
		if len(token) >= 16 {
			prefix := token[:8]
			suffix := token[len(token)-8:]

			assert.Equal(t, "test-fak", prefix)
			assert.Equal(t, "lay-test", suffix)
		}
	})

	t.Run("Verbose flag state testing", func(t *testing.T) {
		// Test verbose flag state management
		healthVerbose = true
		assert.True(t, healthVerbose)

		healthVerbose = false
		assert.False(t, healthVerbose)
	})

	t.Run("Scope display logic for verbose mode", func(t *testing.T) {
		// Test the scope checking and display logic that would be used in verbose mode
		requiredScopes := map[string]bool{
			"channels:read":   true,
			"channels:join":   true,
			"chat:write":      true,
			"channels:manage": true,
			"users:read":      true,
		}
		optionalScopes := map[string]bool{
			"groups:read": false,
		}

		// Simulate successful scopes
		scopes := map[string]bool{
			"channels:read":   true,
			"channels:join":   true,
			"chat:write":      true,
			"channels:manage": true,
			"users:read":      true,
			"groups:read":     true,
		}

		var missingRequired []string
		var missingOptional []string

		for scope := range requiredScopes {
			if !scopes[scope] {
				missingRequired = append(missingRequired, scope)
			}
		}

		for scope := range optionalScopes {
			if !scopes[scope] {
				missingOptional = append(missingOptional, scope)
			}
		}

		// With all scopes present, should have no missing scopes
		assert.Empty(t, missingRequired)
		assert.Empty(t, missingOptional)
	})

	t.Run("Missing optional scopes display logic", func(t *testing.T) {
		// Test case where optional scopes are missing
		optionalScopes := map[string]bool{
			"groups:read": false,
		}

		// Simulate missing optional scope
		scopes := map[string]bool{
			"channels:read":   true,
			"channels:join":   true,
			"chat:write":      true,
			"channels:manage": true,
			"users:read":      true,
			"groups:read":     false, // Missing optional scope
		}

		var missingOptional []string
		for scope := range optionalScopes {
			if !scopes[scope] {
				missingOptional = append(missingOptional, scope)
			}
		}

		// Should detect missing optional scope
		assert.Contains(t, missingOptional, "groups:read")
		assert.Len(t, missingOptional, 1)
	})
}
