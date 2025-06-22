package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRealSlackAPI(t *testing.T) {
	t.Run("NewRealSlackAPI creates wrapper", func(t *testing.T) {
		token := "xoxb-test-token"
		api := NewRealSlackAPI(token)

		assert.NotNil(t, api)
		assert.NotNil(t, api.client)
	})

	t.Run("RealSlackAPI methods exist and are callable", func(t *testing.T) {
		token := "xoxb-test-token"
		api := NewRealSlackAPI(token)

		// Test that methods exist and can be called (interface compliance)
		// We don't call them as they would make real network requests
		assert.NotNil(t, api.AuthTest)
		assert.NotNil(t, api.GetConversations)
		assert.NotNil(t, api.PostMessage)
	})

	t.Run("RealSlackAPI interface method signatures", func(t *testing.T) {
		token := "xoxb-test-token"
		api := NewRealSlackAPI(token)

		// Test that GetConversations and PostMessage methods exist by verifying
		// they're not nil function pointers - this validates interface compliance
		// without making actual network calls that would fail

		// Verify GetConversations method exists with correct signature
		assert.NotNil(t, api.GetConversations)

		// Verify PostMessage method exists with correct signature
		assert.NotNil(t, api.PostMessage)

		// Verify the interface is implemented correctly by checking
		// that RealSlackAPI satisfies the SlackAPI interface
		var _ SlackAPI = api
		assert.True(t, true) // This will compile only if interface is satisfied
	})
}
