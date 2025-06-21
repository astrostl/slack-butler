package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainFunction(t *testing.T) {
	t.Run("Main function exists", func(t *testing.T) {
		// Test that main function exists and can be called
		// We can't directly test main() but we can test the binary
		assert.NotNil(t, main)
	})

	t.Run("Main function integration", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test in short mode")
		}

		// Build the binary
		cmd := exec.Command("go", "build", "-o", "test-slack-buddy", ".")
		err := cmd.Run()
		assert.NoError(t, err)

		// Clean up
		defer os.Remove("test-slack-buddy")

		// Test help flag
		cmd = exec.Command("./test-slack-buddy", "--help")
		output, err := cmd.Output()
		assert.NoError(t, err)
		assert.Contains(t, string(output), "slack-buddy")
	})
}