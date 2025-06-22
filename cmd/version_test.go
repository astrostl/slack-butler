package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	// Set test version info
	version = "1.0.0-test"
	buildTime = "2023-01-01T00:00:00Z"
	gitCommit = "abc123def456"

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute version command run function directly
	versionCmd.Run(versionCmd, []string{})

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	output := buf.String()

	// Check output contains expected information
	if !strings.Contains(output, "slack-buddy version 1.0.0-test") {
		t.Errorf("Expected version in output, got: %s", output)
	}

	if !strings.Contains(output, "Built: 2023-01-01T00:00:00Z") {
		t.Errorf("Expected build time in output, got: %s", output)
	}

	if !strings.Contains(output, "Commit: abc123def456") {
		t.Errorf("Expected commit in output, got: %s", output)
	}
}

func TestVersionCommandDefaults(t *testing.T) {
	// Set test version info with defaults
	version = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute version command run function directly
	versionCmd.Run(versionCmd, []string{})

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	output := buf.String()

	// Check output contains version but not unknown values
	if !strings.Contains(output, "slack-buddy version dev") {
		t.Errorf("Expected version in output, got: %s", output)
	}

	// Should not show "unknown" values
	if strings.Contains(output, "Built: unknown") {
		t.Errorf("Should not show unknown build time, got: %s", output)
	}

	if strings.Contains(output, "Commit: unknown") {
		t.Errorf("Should not show unknown commit, got: %s", output)
	}
}
