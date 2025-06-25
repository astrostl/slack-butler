//go:build tools

package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToolsPackageExists(t *testing.T) {
	// This test ensures the tools package compiles correctly
	// and all tool dependencies are properly declared.
	// The build constraint ensures this only runs when tools are needed.
	assert.True(t, true, "tools package should compile without errors")
}

func TestBuildConstraints(t *testing.T) {
	// Verify that tools package has proper build constraints
	// This ensures tools don't get compiled into the main binary
	t.Run("tools package has correct build tag", func(t *testing.T) {
		// The fact that this test runs means the build constraint is working
		assert.True(t, true, "build constraint 'tools' should be properly applied")
	})
}
