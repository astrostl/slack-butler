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
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Execute version command run function directly
	versionCmd.Run(versionCmd, []string{})

	// Restore stdout and read output
	if err := w.Close(); err != nil {
		t.Logf("Warning: failed to close pipe writer: %v", err)
	}
	os.Stdout = oldStdout

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Check output contains expected information
	if !strings.Contains(output, "slack-butler version 1.0.0-test") {
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
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	// Execute version command run function directly
	versionCmd.Run(versionCmd, []string{})

	// Restore stdout and read output
	if err := w.Close(); err != nil {
		t.Logf("Warning: failed to close pipe writer: %v", err)
	}
	os.Stdout = oldStdout

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Check output contains version but not unknown values
	if !strings.Contains(output, "slack-butler version dev") {
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
