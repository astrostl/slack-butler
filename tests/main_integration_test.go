package main_test

import (
	"os"
	"testing"

	"github.com/astrostl/slack-buddy-ai/cmd"
)

// TestMainFunctionExists verifies that the main function integration works
func TestMainFunctionExists(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "Main with version flag",
			args: []string{"slack-buddy", "--version"},
		},
		{
			name: "Main with help flag",
			args: []string{"slack-buddy", "--help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			// Set test args
			os.Args = tt.args

			// Test that Execute function can be called
			// We don't test main() directly as it would exit,
			// but verify the Execute path works
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Execute panicked with: %v", r)
				}
			}()

			// Test Execute function with build info
			cmd.Execute("test-version", "test-time", "test-commit")
		})
	}
}
