package cmd

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestExecuteFunction(t *testing.T) {
	t.Run("Execute function exists", func(t *testing.T) {
		assert.NotNil(t, Execute)
	})

	t.Run("Execute with help flag", func(t *testing.T) {
		// Temporarily redirect stdout to capture output
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		os.Args = []string{"test", "--help"}

		// Execute should not panic
		assert.NotPanics(t, func() {
			Execute("dev", "unknown", "unknown")
		})
	})
}

func TestRootCommand(t *testing.T) {
	t.Run("Root command exists", func(t *testing.T) {
		assert.NotNil(t, rootCmd)
		assert.Equal(t, "slack-buddy", rootCmd.Use)
	})
}

func TestCommandStructure(t *testing.T) {
	t.Run("Commands are properly registered", func(t *testing.T) {
		commands := rootCmd.Commands()
		assert.Greater(t, len(commands), 0)

		// Check if channels command exists
		var channelsCmd *cobra.Command
		for _, cmd := range commands {
			if cmd.Use == "channels" {
				channelsCmd = cmd
				break
			}
		}
		assert.NotNil(t, channelsCmd)
	})
}

func TestInitConfig(t *testing.T) {
	t.Run("InitConfig sets up viper correctly", func(t *testing.T) {
		// initConfig should not panic
		assert.NotPanics(t, func() {
			initConfig()
		})
	})

	t.Run("Environment variables are properly bound", func(t *testing.T) {
		// Test that SLACK_TOKEN environment variable is read by viper
		expectedToken := "xoxb-test-environment-token-12345678901234567890"
		t.Setenv("SLACK_TOKEN", expectedToken)

		// Initialize config to bind environment variables
		initConfig()

		// Verify viper can read the environment variable
		actualToken := viper.GetString("token")
		assert.Equal(t, expectedToken, actualToken, "viper should read SLACK_TOKEN environment variable")
	})

	t.Run("Environment variables take precedence over empty flags", func(t *testing.T) {
		// Test that environment variables work when flags are not set
		expectedToken := "xoxb-env-precedence-token-12345678901234567890"
		t.Setenv("SLACK_TOKEN", expectedToken)

		// Initialize config to rebind environment variables
		initConfig()

		// Environment variable should be picked up
		actualToken := viper.GetString("token")
		assert.Equal(t, expectedToken, actualToken, "environment variable should take precedence when flag is empty")
	})
}

func TestRootCommandVersionFlag(t *testing.T) {
	t.Run("Version flag with build info", func(t *testing.T) {
		// Set version info like it would be set during build
		Execute("v1.0.0", "2025-06-22_12:00:00", "abc1234")

		// Test version flag handling
		versionFlag := rootCmd.Flags().Lookup("version")
		assert.NotNil(t, versionFlag)
		assert.Equal(t, "v", versionFlag.Shorthand)

		// Set version flag
		err := rootCmd.Flags().Set("version", "true")
		assert.NoError(t, err)

		// Verify flag was set
		versionSet, err := rootCmd.Flags().GetBool("version")
		assert.NoError(t, err)
		assert.True(t, versionSet)
	})

	t.Run("Root command Run function exists", func(t *testing.T) {
		// Verify that the root command has a Run function defined
		assert.NotNil(t, rootCmd.Run)

		// Test that the function can be called without panicking
		// (it will show help output, which is expected behavior)
		assert.NotPanics(t, func() {
			rootCmd.Run(rootCmd, []string{})
		})
	})
}
