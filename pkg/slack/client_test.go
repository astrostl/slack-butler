package slack

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClientWithAPI(t *testing.T) {
	t.Run("Success with mock API", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, err := NewClientWithAPI(mockAPI)
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("Nil API error", func(t *testing.T) {
		client, err := NewClientWithAPI(nil)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "API cannot be nil")
	})

	t.Run("Auth failure", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.SetAuthError(true)
		
		client, err := NewClientWithAPI(mockAPI)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "authentication failed")
	})
}

func TestGetNewChannels(t *testing.T) {
	t.Run("Success with new channels", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)
		
		// Add a channel created 1 hour ago
		createdTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C1234567890", "test-channel", createdTime, "Test channel")
		
		// Look for channels created in the last 2 hours
		since := time.Now().Add(-2 * time.Hour)
		channels, err := client.GetNewChannels(since)
		
		assert.NoError(t, err)
		assert.Len(t, channels, 1)
		assert.Equal(t, "test-channel", channels[0].Name)
	})

	t.Run("No new channels", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)
		
		// Add a channel created 3 hours ago
		createdTime := time.Now().Add(-3 * time.Hour)
		mockAPI.AddChannel("C1234567890", "old-channel", createdTime, "Old channel")
		
		// Look for channels created in the last 2 hours
		since := time.Now().Add(-2 * time.Hour)
		channels, err := client.GetNewChannels(since)
		
		assert.NoError(t, err)
		assert.Len(t, channels, 0)
	})

	t.Run("Channel created exactly at boundary", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)
		
		// Create a specific time for boundary testing
		boundaryTime := time.Now().Add(-2 * time.Hour)
		
		// Add a channel created exactly at the boundary time
		mockAPI.AddChannel("C1234567890", "boundary-channel", boundaryTime, "Boundary channel")
		
		// Look for channels created after the boundary (should not include the boundary channel)
		channels, err := client.GetNewChannels(boundaryTime)
		
		assert.NoError(t, err)
		assert.Len(t, channels, 0) // Channel created AT boundary time should not be included
	})

	t.Run("Channel created one second after boundary", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)
		
		// Create a specific time for boundary testing
		boundaryTime := time.Now().Add(-2 * time.Hour)
		
		// Add a channel created one second after the boundary
		oneSecondAfter := boundaryTime.Add(1 * time.Second)
		mockAPI.AddChannel("C1234567890", "after-boundary-channel", oneSecondAfter, "After boundary")
		
		// Look for channels created after the boundary (should include this channel)
		channels, err := client.GetNewChannels(boundaryTime)
		
		assert.NoError(t, err)
		assert.Len(t, channels, 1)
		assert.Equal(t, "after-boundary-channel", channels[0].Name)
	})

	t.Run("API error handling", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)
		
		mockAPI.SetGetConversationsError(true)
		
		since := time.Now().Add(-2 * time.Hour)
		_, err := client.GetNewChannels(since)
		
		assert.Error(t, err)
	})

	t.Run("Missing scope error", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)
		
		mockAPI.SetMissingScopeError(true)
		
		since := time.Now().Add(-2 * time.Hour)
		_, err := client.GetNewChannels(since)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required permissions")
		assert.Contains(t, err.Error(), "channels:read")
		assert.Contains(t, err.Error(), "groups:read")
	})

	t.Run("Invalid auth error", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)
		
		mockAPI.SetInvalidAuthError(true)
		
		since := time.Now().Add(-2 * time.Hour)
		_, err := client.GetNewChannels(since)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token")
	})
}

func TestFormatNewChannelAnnouncement(t *testing.T) {
	mockAPI := NewMockSlackAPI()
	client, _ := NewClientWithAPI(mockAPI)

	t.Run("Single channel", func(t *testing.T) {
		channels := []Channel{
			{
				ID:      "C1234567890",
				Name:    "test-channel",
				Created: time.Now(),
				Purpose: "Test purpose",
				Creator: "U1234567",
			},
		}
		
		since := time.Now().Add(-1 * time.Hour)
		message := client.FormatNewChannelAnnouncement(channels, since)
		
		assert.Contains(t, message, "New channel alert!")
		assert.Contains(t, message, "C1234567890") // Channel ID in link format
		assert.Contains(t, message, "Test purpose")
		assert.Contains(t, message, "by <@U1234567>")
		assert.Contains(t, message, "created")
		// Check for natural date format
		expectedDate := time.Now().Format("January 2, 2006")
		assert.Contains(t, message, expectedDate)
	})

	t.Run("Multiple channels", func(t *testing.T) {
		channels := []Channel{
			{ID: "C1234567890", Name: "channel1", Created: time.Now(), Creator: "U1111111"},
			{ID: "C0987654321", Name: "channel2", Created: time.Now(), Creator: "U2222222"},
		}
		
		since := time.Now().Add(-1 * time.Hour)
		message := client.FormatNewChannelAnnouncement(channels, since)
		
		assert.Contains(t, message, "2 new channels created!")
		assert.Contains(t, message, "C1234567890") // First channel ID
		assert.Contains(t, message, "C0987654321") // Second channel ID
		assert.Contains(t, message, "by <@U1111111>")
		assert.Contains(t, message, "by <@U2222222>")
		// Check for spacing between channels (should have double newlines)
		assert.Contains(t, message, "\n\n•")
	})

	t.Run("Channel without purpose", func(t *testing.T) {
		channels := []Channel{
			{
				ID:      "C1234567890",
				Name:    "test-channel",
				Created: time.Now(),
				Purpose: "",
				Creator: "U3333333",
			},
		}
		
		since := time.Now().Add(-1 * time.Hour)
		message := client.FormatNewChannelAnnouncement(channels, since)
		
		assert.Contains(t, message, "C1234567890") // Channel ID in link format
		assert.Contains(t, message, "by <@U3333333>")
		assert.NotContains(t, message, "Purpose:")
	})

	t.Run("Channel without creator", func(t *testing.T) {
		channels := []Channel{
			{
				ID:      "C1234567890",
				Name:    "test-channel",
				Created: time.Now(),
				Purpose: "Test purpose",
				Creator: "",
			},
		}
		
		since := time.Now().Add(-1 * time.Hour)
		message := client.FormatNewChannelAnnouncement(channels, since)
		
		assert.Contains(t, message, "C1234567890") // Channel ID in link format
		assert.Contains(t, message, "Purpose: Test purpose")
		assert.NotContains(t, message, " by <@")
	})
}

func TestPostMessage(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		// Add the general channel that will be used for posting
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")
		client, _ := NewClientWithAPI(mockAPI)
		
		err := client.PostMessage("#general", "Test message")
		assert.NoError(t, err)
		
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 1)
		assert.Equal(t, "CGENERAL", messages[0].ChannelID)
	})

	t.Run("Channel name validation", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)
		
		err := client.PostMessage("", "Test message")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid channel name")
	})

	t.Run("Missing scope error", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		// Add the general channel that will be used for posting
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")
		client, _ := NewClientWithAPI(mockAPI)
		
		mockAPI.SetPostMessageError("missing_scope")
		
		err := client.PostMessage("#general", "Test message")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required permission")
		assert.Contains(t, err.Error(), "chat:write")
	})

	t.Run("Channel not found error", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)
		
		mockAPI.SetPostMessageError("channel_not_found")
		
		err := client.PostMessage("#nonexistent", "Test message")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "channel '#nonexistent' not found")
	})

	t.Run("Not in channel error", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		// Add the private channel that will be used for posting
		mockAPI.AddChannel("CPRIVATE", "private", time.Now().Add(-24*time.Hour), "Private channel")
		client, _ := NewClientWithAPI(mockAPI)
		
		mockAPI.SetPostMessageError("not_in_channel")
		
		err := client.PostMessage("#private", "Test message")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bot is not a member")
	})
	
	t.Run("Generic error", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		// Add the general channel that will be used for posting
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")
		client, _ := NewClientWithAPI(mockAPI)
		
		mockAPI.SetPostMessageError("some_other_error")
		
		err := client.PostMessage("#general", "Test message")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to post message to #general")
		assert.Contains(t, err.Error(), "some_other_error")
	})
}

func TestExtractChannelIDs(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			"Single channel mention",
			"Check out <#C1234567890>",
			[]string{"C1234567890"},
		},
		{
			"Multiple channel mentions",
			"See <#C1111111111> and <#C2222222222>",
			[]string{"C1111111111", "C2222222222"},
		},
		{
			"Channel mention with pipe",
			"Visit <#C1234567890|general>",
			[]string{"C1234567890"},
		},
		{
			"Mixed format mentions",
			"Channels: <#C1111111111> and <#C2222222222|random>",
			[]string{"C1111111111", "C2222222222"},
		},
		{
			"No channel mentions",
			"This is just regular text",
			[]string{},
		},
		{
			"Announcement format",
			"New channel alert!\n\n• <#C1234567890> - created June 22, 2025 by <@U1234567>\n• <#C0987654321> - created June 22, 2025 by <@U2222222>",
			[]string{"C1234567890", "C0987654321"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractChannelIDs(tc.text)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetPreviouslyAnnouncedChannels(t *testing.T) {
	t.Run("Success with announced channels", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")
		client, _ := NewClientWithAPI(mockAPI)
		
		// Add some message history with channel announcements
		// Use bot's user ID (U0000000) for messages that should be processed
		mockAPI.AddMessageToHistory("CGENERAL", "New channel alert!\n\n• <#C1234567890> - created June 22, 2025 by <@U1234567>", "U0000000", "1234567890.123")
		// This message from another user should be ignored
		mockAPI.AddMessageToHistory("CGENERAL", "Check out <#C0987654321|random>", "U1111111", "1234567891.123")
		
		channels, err := client.GetPreviouslyAnnouncedChannels("#general")
		
		assert.NoError(t, err)
		assert.True(t, channels["C1234567890"])  // From bot message
		assert.False(t, channels["C0987654321"]) // From other user, should be ignored
		assert.False(t, channels["C9999999999"]) // Not announced
	})

	t.Run("No history", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")
		client, _ := NewClientWithAPI(mockAPI)
		
		channels, err := client.GetPreviouslyAnnouncedChannels("#general")
		
		assert.NoError(t, err)
		assert.Len(t, channels, 0)
	})

	t.Run("Ignore messages from other users", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")
		client, _ := NewClientWithAPI(mockAPI)
		
		// Add messages from bot and other users
		mockAPI.AddMessageToHistory("CGENERAL", "Bot announcement: <#C1234567890>", "U0000000", "1234567890.123") // Bot message
		mockAPI.AddMessageToHistory("CGENERAL", "User mention: <#C0987654321>", "U1111111", "1234567891.123")     // Other user message
		
		channels, err := client.GetPreviouslyAnnouncedChannels("#general")
		
		assert.NoError(t, err)
		assert.True(t, channels["C1234567890"])  // From bot, should be included
		assert.False(t, channels["C0987654321"]) // From other user, should be ignored
	})

	t.Run("API error", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")
		client, _ := NewClientWithAPI(mockAPI)
		
		mockAPI.SetConversationHistoryError(true)
		
		_, err := client.GetPreviouslyAnnouncedChannels("#general")
		assert.Error(t, err)
	})
}

func TestFilterAlreadyAnnouncedChannels(t *testing.T) {
	t.Run("Filter out previously announced", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")
		client, _ := NewClientWithAPI(mockAPI)
		
		// Set up history with one announced channel (from bot)
		mockAPI.AddMessageToHistory("CGENERAL", "New channel alert!\n\n• <#C1234567890> - created June 22, 2025", "U0000000", "1234567890.123")
		
		channels := []Channel{
			{ID: "C1234567890", Name: "already-announced"},
			{ID: "C0987654321", Name: "new-channel"},
		}
		
		filtered, err := client.FilterAlreadyAnnouncedChannels(channels, "#general")
		
		assert.NoError(t, err)
		assert.Len(t, filtered, 1)
		assert.Equal(t, "new-channel", filtered[0].Name)
	})

	t.Run("No announcement channel", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)
		
		channels := []Channel{
			{ID: "C1234567890", Name: "channel1"},
			{ID: "C0987654321", Name: "channel2"},
		}
		
		filtered, err := client.FilterAlreadyAnnouncedChannels(channels, "")
		
		assert.NoError(t, err)
		assert.Len(t, filtered, 2) // All channels returned
	})

	t.Run("All channels already announced", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")
		client, _ := NewClientWithAPI(mockAPI)
		
		// Set up history with both channels announced (from bot)
		mockAPI.AddMessageToHistory("CGENERAL", "• <#C1234567890> and <#C0987654321>", "U0000000", "1234567890.123")
		
		channels := []Channel{
			{ID: "C1234567890", Name: "channel1"},
			{ID: "C0987654321", Name: "channel2"},
		}
		
		filtered, err := client.FilterAlreadyAnnouncedChannels(channels, "#general")
		
		assert.NoError(t, err)
		assert.Len(t, filtered, 0) // No channels returned
	})
}

func TestGetChannelInfo(t *testing.T) {
	t.Run("GetChannelInfo returns expected error", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)
		
		// This function is a stub used for permission testing
		info, err := client.GetChannelInfo("C1234567890")
		
		assert.Error(t, err)
		assert.Nil(t, info)
		assert.Contains(t, err.Error(), "channel_not_found")
	})
}