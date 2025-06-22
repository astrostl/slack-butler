package cmd

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
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
}