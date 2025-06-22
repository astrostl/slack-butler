package slack

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMockSlackAPI(t *testing.T) {
	t.Run("Creates mock with defaults", func(t *testing.T) {
		mock := NewMockSlackAPI()
		assert.NotNil(t, mock)
		assert.NotNil(t, mock.AuthTestResponse)
		assert.Equal(t, "test-bot", mock.AuthTestResponse.User)
		assert.Equal(t, "Test Team", mock.AuthTestResponse.Team)
		assert.Empty(t, mock.Channels)
		assert.Empty(t, mock.PostedMessages)
	})
}

func TestMockAuthTest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := NewMockSlackAPI()
		resp, err := mock.AuthTest()
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "test-bot", resp.User)
	})

	t.Run("Error", func(t *testing.T) {
		mock := NewMockSlackAPI()
		mock.SetAuthError(true)

		resp, err := mock.AuthTest()
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestMockGetConversations(t *testing.T) {
	t.Run("Success with channels", func(t *testing.T) {
		mock := NewMockSlackAPI()

		// Add a test channel
		createdTime := time.Now()
		mock.AddChannel("C1234567890", "test-channel", createdTime, "Test purpose")

		channels, cursor, err := mock.GetConversations(nil)
		assert.NoError(t, err)
		assert.Empty(t, cursor)
		assert.Len(t, channels, 1)
		assert.Equal(t, "test-channel", channels[0].Name)
		assert.Equal(t, "Test purpose", channels[0].Purpose.Value)
	})

	t.Run("Error", func(t *testing.T) {
		mock := NewMockSlackAPI()
		mock.SetGetConversationsError(true)

		channels, cursor, err := mock.GetConversations(nil)
		assert.Error(t, err)
		assert.Empty(t, cursor)
		assert.Nil(t, channels)
	})
}

func TestMockPostMessage(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := NewMockSlackAPI()

		channelID, timestamp, err := mock.PostMessage("general", nil)
		assert.NoError(t, err)
		assert.Equal(t, "mock-channel-id", channelID)
		assert.Equal(t, "mock-timestamp", timestamp)

		messages := mock.GetPostedMessages()
		assert.Len(t, messages, 1)
		assert.Equal(t, "general", messages[0].ChannelID)
	})

	t.Run("Error", func(t *testing.T) {
		mock := NewMockSlackAPI()
		mock.SetPostMessageError("generic_error")

		channelID, timestamp, err := mock.PostMessage("general", nil)
		assert.Error(t, err)
		assert.Empty(t, channelID)
		assert.Empty(t, timestamp)
	})
}

func TestMockHelperMethods(t *testing.T) {
	t.Run("AddChannel", func(t *testing.T) {
		mock := NewMockSlackAPI()
		createdTime := time.Now()

		mock.AddChannel("C1234567890", "test-channel", createdTime, "Test purpose")

		assert.Len(t, mock.Channels, 1)
		assert.Equal(t, "C1234567890", mock.Channels[0].ID)
		assert.Equal(t, "test-channel", mock.Channels[0].Name)
	})

	t.Run("ClearPostedMessages", func(t *testing.T) {
		mock := NewMockSlackAPI()

		// Post a message
		mock.PostMessage("test", nil)
		assert.Len(t, mock.GetPostedMessages(), 1)

		// Clear messages
		mock.ClearPostedMessages()
		assert.Empty(t, mock.GetPostedMessages())
	})

	t.Run("Error simulation methods", func(t *testing.T) {
		mock := NewMockSlackAPI()

		mock.SimulateMissingScopeError()
		assert.NotNil(t, mock.GetConversationsError)

		mock.SimulateInvalidAuthError()
		assert.NotNil(t, mock.AuthTestError)

		mock.SimulateChannelNotFoundError()
		assert.NotNil(t, mock.PostMessageError)

		mock.SimulateNotInChannelError()
		assert.NotNil(t, mock.PostMessageError)
	})
}

func TestMockErrorSettersEdgeCases(t *testing.T) {
	t.Run("SetAuthError with enable false", func(t *testing.T) {
		mock := NewMockSlackAPI()
		mock.SetAuthError(true) // Enable first
		assert.NotNil(t, mock.AuthTestError)

		mock.SetAuthError(false) // Disable
		assert.Nil(t, mock.AuthTestError)
	})

	t.Run("SetGetConversationsError with enable false", func(t *testing.T) {
		mock := NewMockSlackAPI()
		mock.SetGetConversationsError(true) // Enable first
		assert.NotNil(t, mock.GetConversationsError)

		mock.SetGetConversationsError(false) // Disable
		assert.Nil(t, mock.GetConversationsError)
	})

	t.Run("SetMissingScopeError with enable false", func(t *testing.T) {
		mock := NewMockSlackAPI()
		mock.SetMissingScopeError(true) // Enable first
		assert.NotNil(t, mock.GetConversationsError)

		mock.SetMissingScopeError(false) // Disable
		assert.Nil(t, mock.GetConversationsError)
	})

	t.Run("SetInvalidAuthError with enable false", func(t *testing.T) {
		mock := NewMockSlackAPI()
		mock.SetInvalidAuthError(true) // Enable first
		assert.NotNil(t, mock.GetConversationsError)

		mock.SetInvalidAuthError(false) // Disable
		assert.Nil(t, mock.GetConversationsError)
	})

	t.Run("SetPostMessageError with empty string", func(t *testing.T) {
		mock := NewMockSlackAPI()
		mock.SetPostMessageError("some_error") // Enable first
		assert.NotNil(t, mock.PostMessageError)

		mock.SetPostMessageError("") // Clear with empty string
		assert.Nil(t, mock.PostMessageError)
	})
}
