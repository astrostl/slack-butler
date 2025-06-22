package main

import (
	"testing"
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