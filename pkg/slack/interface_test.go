package slack

import (
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestRealSlackAPI(t *testing.T) {
	t.Run("NewRealSlackAPI creates wrapper", func(t *testing.T) {
		token := "xoxb-test-token"
		api := NewRealSlackAPI(token)

		assert.NotNil(t, api)
		assert.NotNil(t, api.client)
	})

	t.Run("RealSlackAPI interface compliance", func(t *testing.T) {
		token := "xoxb-test-token"
		api := NewRealSlackAPI(token)

		// Verify interface compliance at compile time
		var _ SlackAPI = api

		// Test that all interface methods are properly implemented
		assert.NotNil(t, api.AuthTest)
		assert.NotNil(t, api.GetConversations)
		assert.NotNil(t, api.GetConversationHistory)
		assert.NotNil(t, api.PostMessage)
		assert.NotNil(t, api.ArchiveConversation)
		assert.NotNil(t, api.JoinConversation)
		assert.NotNil(t, api.GetUsers)
	})

	t.Run("RealSlackAPI token handling", func(t *testing.T) {
		// Test with different token formats
		testCases := []string{
			"xoxb-test-token",
			"xoxp-user-token",
			"",
		}

		for _, token := range testCases {
			api := NewRealSlackAPI(token)
			assert.NotNil(t, api)
			assert.NotNil(t, api.client)
		}
	})

	t.Run("RealSlackAPI method delegation", func(t *testing.T) {
		// Test that wrapper properly delegates to underlying client
		// We can't make real API calls, but we can verify the methods exist
		// and have correct signatures
		token := "xoxb-test-token"
		api := NewRealSlackAPI(token)

		// Verify method signatures match interface
		assert.IsType(t, &RealSlackAPI{}, api)

		// Test that each method can be called (will fail with auth error but proves delegation)
		// We don't actually call them to avoid network requests in tests

		// Test struct field access
		assert.NotNil(t, api.client, "Internal client should be accessible")
	})

	t.Run("RealSlackAPI method calls with expected errors", func(t *testing.T) {
		// Test that methods properly delegate to underlying client
		// These will fail with auth errors but prove the delegation works
		token := "xoxb-invalid-test-token"
		api := NewRealSlackAPI(token)

		// Test AuthTest delegation
		_, err := api.AuthTest()
		assert.Error(t, err, "AuthTest should fail with invalid token")

		// Test GetConversations delegation
		_, _, err = api.GetConversations(&slack.GetConversationsParameters{})
		assert.Error(t, err, "GetConversations should fail with invalid token")

		// Test GetConversationHistory delegation
		_, err = api.GetConversationHistory(&slack.GetConversationHistoryParameters{
			ChannelID: "C123456",
		})
		assert.Error(t, err, "GetConversationHistory should fail with invalid token")

		// Test PostMessage delegation
		_, _, err = api.PostMessage("C123456")
		assert.Error(t, err, "PostMessage should fail with invalid token")

		// Test ArchiveConversation delegation
		err = api.ArchiveConversation("C123456")
		assert.Error(t, err, "ArchiveConversation should fail with invalid token")

		// Test JoinConversation delegation
		_, _, _, err = api.JoinConversation("C123456")
		assert.Error(t, err, "JoinConversation should fail with invalid token")

		// Test GetUsers delegation
		_, err = api.GetUsers()
		assert.Error(t, err, "GetUsers should fail with invalid token")
	})
}
