package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionVariables(t *testing.T) {
	// Test that version variables exist and have default values
	if Version == "" {
		t.Error("Version variable should not be empty")
	}
	if BuildTime == "" {
		t.Error("BuildTime variable should not be empty")
	}
	if GitCommit == "" {
		t.Error("GitCommit variable should not be empty")
	}
}

func TestMainIntegration(t *testing.T) {
	t.Run("CLI binary builds and runs", func(t *testing.T) {
		// Build the binary
		buildCmd := exec.Command("go", "build", "-o", "slack-buddy-test", ".")
		buildCmd.Dir = "."
		err := buildCmd.Run()
		require.NoError(t, err, "Failed to build CLI binary")

		// Clean up binary after test
		defer func() {
			os.Remove("./slack-buddy-test")
		}()

		// Test version flag
		versionCmd := exec.Command("./slack-buddy-test", "--version")
		output, err := versionCmd.Output()
		require.NoError(t, err, "Version command failed")

		outputStr := strings.TrimSpace(string(output))
		assert.Contains(t, outputStr, "slack-buddy", "Version output should contain tool name")
		assert.Contains(t, outputStr, "dev", "Version output should contain version")
	})

	t.Run("Help command works", func(t *testing.T) {
		// Build the binary
		buildCmd := exec.Command("go", "build", "-o", "slack-buddy-test", ".")
		buildCmd.Dir = "."
		err := buildCmd.Run()
		require.NoError(t, err, "Failed to build CLI binary")

		// Clean up binary after test
		defer func() {
			os.Remove("./slack-buddy-test")
		}()

		// Test help flag
		helpCmd := exec.Command("./slack-buddy-test", "--help")
		output, err := helpCmd.Output()
		require.NoError(t, err, "Help command failed")

		outputStr := string(output)
		assert.Contains(t, outputStr, "Usage:", "Help should contain usage info")
		assert.Contains(t, outputStr, "Available Commands:", "Help should list commands")
		assert.Contains(t, outputStr, "channels", "Help should mention channels command")
		assert.Contains(t, outputStr, "health", "Help should mention health command")
	})

	t.Run("Invalid command shows error", func(t *testing.T) {
		// Build the binary
		buildCmd := exec.Command("go", "build", "-o", "slack-buddy-test", ".")
		buildCmd.Dir = "."
		err := buildCmd.Run()
		require.NoError(t, err, "Failed to build CLI binary")

		// Clean up binary after test
		defer func() {
			os.Remove("./slack-buddy-test")
		}()

		// Test invalid command
		invalidCmd := exec.Command("./slack-buddy-test", "nonexistent-command")
		output, err := invalidCmd.CombinedOutput()

		// Should exit with non-zero code
		assert.Error(t, err, "Invalid command should return error")

		outputStr := string(output)
		assert.Contains(t, outputStr, "unknown command", "Should indicate unknown command")
	})
}
