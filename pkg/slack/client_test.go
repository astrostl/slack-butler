package slack

import (
	"fmt"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestAutoJoinPublicChannelsErrorPaths(t *testing.T) {
	t.Run("Rate limited during auto-join", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Set up mock to return rate limit error
		mockAPI.SetJoinError("rate_limited")

		channels := []slack.Channel{
			{
				GroupConversation: slack.GroupConversation{
					Conversation: slack.Conversation{
						ID:        "C1",
						IsPrivate: false,
					},
					Name: "test-channel",
				},
			},
		}

		joinedCount, err := client.autoJoinPublicChannels(channels)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limited during auto-join")
		assert.Equal(t, 0, joinedCount)
	})

	t.Run("Missing channels:join scope", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Set up mock to return missing scope error
		mockAPI.SetJoinError("missing_scope")

		channels := []slack.Channel{
			{
				GroupConversation: slack.GroupConversation{
					Conversation: slack.Conversation{
						ID:        "C1",
						IsPrivate: false,
					},
					Name: "test-channel",
				},
			},
		}

		joinedCount, err := client.autoJoinPublicChannels(channels)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required OAuth scope")
		assert.Contains(t, err.Error(), "channels:join")
		assert.Equal(t, 0, joinedCount)
	})

	t.Run("Invalid auth token", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Set up mock to return invalid auth error
		mockAPI.SetJoinError("invalid_auth")

		channels := []slack.Channel{
			{
				GroupConversation: slack.GroupConversation{
					Conversation: slack.Conversation{
						ID:        "C1",
						IsPrivate: false,
					},
					Name: "test-channel",
				},
			},
		}

		joinedCount, err := client.autoJoinPublicChannels(channels)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid authentication token")
		assert.Equal(t, 0, joinedCount)
	})

	t.Run("Successfully handles already_in_channel", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Set up mock to return already in channel
		mockAPI.SetJoinError("already_in_channel")

		channels := []slack.Channel{
			{
				GroupConversation: slack.GroupConversation{
					Conversation: slack.Conversation{
						ID:        "C1",
						IsPrivate: false,
					},
					Name: "test-channel",
				},
			},
		}

		joinedCount, err := client.autoJoinPublicChannels(channels)
		assert.NoError(t, err)
		assert.Equal(t, 1, joinedCount) // Should count as successful join
	})

	t.Run("Skips archived channels", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Set up mock to return archived error
		mockAPI.SetJoinError("is_archived")

		channels := []slack.Channel{
			{
				GroupConversation: slack.GroupConversation{
					Conversation: slack.Conversation{
						ID:        "C1",
						IsPrivate: false,
					},
					Name: "archived-channel",
				},
			},
		}

		joinedCount, err := client.autoJoinPublicChannels(channels)
		assert.NoError(t, err)
		assert.Equal(t, 0, joinedCount) // Should be skipped, not counted
	})

	t.Run("Skips invite-only channels", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Set up mock to return invite only error
		mockAPI.SetJoinError("invite_only")

		channels := []slack.Channel{
			{
				GroupConversation: slack.GroupConversation{
					Conversation: slack.Conversation{
						ID:        "C1",
						IsPrivate: false,
					},
					Name: "invite-only-channel",
				},
			},
		}

		joinedCount, err := client.autoJoinPublicChannels(channels)
		assert.NoError(t, err)
		assert.Equal(t, 0, joinedCount) // Should be skipped, not counted
	})

	t.Run("Skips private channels", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// No errors needed - should skip by design
		channels := []slack.Channel{
			{
				GroupConversation: slack.GroupConversation{
					Conversation: slack.Conversation{
						ID:        "C1",
						IsPrivate: true,
					},
					Name: "private-channel",
				},
			},
		}

		joinedCount, err := client.autoJoinPublicChannels(channels)
		assert.NoError(t, err)
		assert.Equal(t, 0, joinedCount) // Should be skipped
	})
}

func TestPostMessageToChannelIDErrorPaths(t *testing.T) {
	t.Run("Rate limited during message posting", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Set up mock to return rate limit error
		mockAPI.SetPostMessageError("rate_limited")

		err := client.postMessageToChannelID("C123", "test message")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limited")
	})

	t.Run("Missing chat:write scope", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Set up mock to return missing scope error
		mockAPI.SetPostMessageError("missing_scope")

		err := client.postMessageToChannelID("C123", "test message")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required permission to post messages")
		assert.Contains(t, err.Error(), "chat:write")
	})

	t.Run("Channel not found", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Set up mock to return channel not found error
		mockAPI.SetPostMessageError("channel_not_found")

		err := client.postMessageToChannelID("C123", "test message")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "channel with ID 'C123' not found")
	})

	t.Run("Bot not in channel", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Set up mock to return not in channel error
		mockAPI.SetPostMessageError("not_in_channel")

		err := client.postMessageToChannelID("C123", "test message")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bot is not a member of channel")
		assert.Contains(t, err.Error(), "Please add the bot to the channel")
	})

	t.Run("Invalid auth token", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Set up mock to return invalid auth error
		mockAPI.SetPostMessageError("invalid_auth")

		err := client.postMessageToChannelID("C123", "test message")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to post message to channel")
	})

	t.Run("Successful message posting", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// No errors - should succeed
		err := client.postMessageToChannelID("C123", "test message")
		assert.NoError(t, err)
	})
}

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

func TestGetNewChannelsErrorHandling(t *testing.T) {
	t.Run("Rate limit error handling", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Configure mock to return rate limit error
		mockAPI.SetGetConversationsErrorWithMessage(true, "rate_limited")

		since := time.Now().Add(-1 * time.Hour)
		channels, err := client.GetNewChannels(since)

		assert.Error(t, err)
		assert.Nil(t, channels)
		assert.Contains(t, err.Error(), "rate limited")

		// Rate limit error properly returned
	})

	t.Run("Missing scope error handling", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Configure mock to return missing scope error
		mockAPI.SetGetConversationsErrorWithMessage(true, "missing_scope")

		since := time.Now().Add(-1 * time.Hour)
		channels, err := client.GetNewChannels(since)

		assert.Error(t, err)
		assert.Nil(t, channels)
		assert.Contains(t, err.Error(), "missing required permissions")
		assert.Contains(t, err.Error(), "channels:read")
		assert.Contains(t, err.Error(), "groups:read")
	})

	t.Run("Invalid auth error handling", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Configure mock to return invalid auth error
		mockAPI.SetGetConversationsErrorWithMessage(true, "invalid_auth")

		since := time.Now().Add(-1 * time.Hour)
		channels, err := client.GetNewChannels(since)

		assert.Error(t, err)
		assert.Nil(t, channels)
		assert.Contains(t, err.Error(), "invalid token")
		assert.Contains(t, err.Error(), "SLACK_TOKEN")
	})

	t.Run("Generic error handling", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Configure mock to return generic error
		mockAPI.SetGetConversationsErrorWithMessage(true, "network_error")

		since := time.Now().Add(-1 * time.Hour)
		channels, err := client.GetNewChannels(since)

		assert.Error(t, err)
		assert.Nil(t, channels)
		assert.Contains(t, err.Error(), "failed to get conversations")
	})
}

func TestPostMessageErrorHandling(t *testing.T) {
	t.Run("Channel not found error via resolve", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Don't add the channel to mock - this will cause the resolve to fail
		err := client.PostMessage("#nonexistent", "Test message")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find channel #nonexistent")
		assert.Contains(t, err.Error(), "channel '#nonexistent' not found")
	})

	t.Run("PostMessage API error with valid channel", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Add the channel so resolve works
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")

		// Configure mock to return rate limit error at PostMessage level
		mockAPI.SetPostMessageError("rate_limited")

		err := client.PostMessage("#general", "Test message")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limited")
	})

	t.Run("PostMessage missing scope error", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Add the channel so resolve works
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")

		// Configure mock to return missing scope error
		mockAPI.SetPostMessageError("missing_scope")

		err := client.PostMessage("#general", "Test message")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required permission")
		assert.Contains(t, err.Error(), "chat:write")
	})
}

func TestChannelExclusionLogic(t *testing.T) {
	t.Run("shouldSkipChannelWithExclusions with exact matches", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		excludeChannels := []string{"general", "announcements"}
		excludePrefixes := []string{}

		testCases := []struct {
			channelName string
			shouldSkip  bool
		}{
			{"general", true},         // Exact match
			{"announcements", true},   // Exact match
			{"random", false},         // No match
			{"general-backup", false}, // Similar but not exact
		}

		for _, tc := range testCases {
			result := client.shouldSkipChannelWithExclusions(tc.channelName, excludeChannels, excludePrefixes)
			assert.Equal(t, tc.shouldSkip, result,
				"Channel %s skip result should be %v", tc.channelName, tc.shouldSkip)
		}
	})

	t.Run("shouldSkipChannelWithExclusions with prefix matches", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		excludeChannels := []string{}
		excludePrefixes := []string{"test-", "dev-"}

		testCases := []struct {
			channelName string
			shouldSkip  bool
		}{
			{"test-channel", true},    // Prefix match
			{"dev-environment", true}, // Prefix match
			{"random", false},         // No match
			{"testing", false},        // Contains but not prefix
		}

		for _, tc := range testCases {
			result := client.shouldSkipChannelWithExclusions(tc.channelName, excludeChannels, excludePrefixes)
			assert.Equal(t, tc.shouldSkip, result,
				"Channel %s skip result should be %v", tc.channelName, tc.shouldSkip)
		}
	})

	t.Run("shouldSkipChannelWithExclusions with empty exclusions", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		result := client.shouldSkipChannelWithExclusions("any-channel", []string{}, []string{})
		assert.False(t, result, "No exclusions should not skip any channels")
	})

	t.Run("shouldSkipChannelWithExclusions with nil exclusions", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		result := client.shouldSkipChannelWithExclusions("any-channel", nil, nil)
		assert.False(t, result, "Nil exclusions should not skip any channels")
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

func TestTestAuth(t *testing.T) {
	tests := []struct {
		setupMock   func(*MockSlackAPI)
		checkResult func(*testing.T, *AuthInfo)
		name        string
		expectError bool
	}{
		{
			name: "Successful auth test",
			setupMock: func(mock *MockSlackAPI) {
				// Mock returns default auth response - values from mock.go defaults
			},
			expectError: false,
			checkResult: func(t *testing.T, auth *AuthInfo) {
				assert.Equal(t, "test-bot", auth.User)
				assert.Equal(t, "U0000000", auth.UserID)
				assert.Equal(t, "Test Team", auth.Team)
				assert.Equal(t, "", auth.TeamID) // Default mock doesn't set TeamID
			},
		},
		{
			name: "Auth failure",
			setupMock: func(mock *MockSlackAPI) {
				mock.SetAuthError(true)
			},
			expectError: true,
			checkResult: func(t *testing.T, auth *AuthInfo) {
				assert.Nil(t, auth)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := NewMockSlackAPI()
			tt.setupMock(mockAPI)

			// For auth failure test, we expect client creation to fail
			// since it calls auth test during construction
			client, err := NewClientWithAPI(mockAPI)

			if tt.expectError {
				assert.Error(t, err)
				tt.checkResult(t, nil)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, client)

			auth, err := client.TestAuth()
			assert.NoError(t, err)
			tt.checkResult(t, auth)
		})
	}
}

func TestFormatInactiveChannelWarning(t *testing.T) {
	mockAPI := NewMockSlackAPI()
	client, _ := NewClientWithAPI(mockAPI)

	channel := Channel{
		ID:      "C1234567890",
		Name:    "test-channel",
		Created: time.Now().Add(-24 * time.Hour),
		Purpose: "Test channel purpose",
		Creator: "U1111111",
	}

	tests := []struct {
		name           string
		warnSeconds    int
		archiveSeconds int
		expectedParts  []string
	}{
		{
			name:           "Warning in seconds",
			warnSeconds:    45,
			archiveSeconds: 15,
			expectedParts:  []string{"45 seconds", "15 seconds", "🚨", "Inactive Channel Warning"},
		},
		{
			name:           "Warning in minutes",
			warnSeconds:    300, // 5 minutes
			archiveSeconds: 120, // 2 minutes
			expectedParts:  []string{"5 minutes", "2 minutes", "🚨", "Inactive Channel Warning"},
		},
		{
			name:           "Warning in hours",
			warnSeconds:    7200, // 2 hours
			archiveSeconds: 3600, // 1 hour
			expectedParts:  []string{"2 hours", "60 minutes", "🚨", "Inactive Channel Warning"},
		},
		{
			name:           "Single minute",
			warnSeconds:    60, // 1 minute
			archiveSeconds: 30, // 30 seconds
			expectedParts:  []string{"1 minute", "30 seconds", "🚨", "Inactive Channel Warning"},
		},
		{
			name:           "Single hour",
			warnSeconds:    3600, // 1 hour
			archiveSeconds: 1800, // 30 minutes
			expectedParts:  []string{"1 hour", "30 minutes", "🚨", "Inactive Channel Warning"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warning := client.FormatInactiveChannelWarning(channel, tt.warnSeconds, tt.archiveSeconds)

			for _, part := range tt.expectedParts {
				assert.Contains(t, warning, part, "Warning should contain: %s", part)
			}

			// Should include the warning comment for detection
			assert.Contains(t, warning, "<!-- inactive channel warning -->")

			// Should include guidance
			assert.Contains(t, warning, "To keep this channel active")
			assert.Contains(t, warning, "Post a message")
		})
	}
}

func TestFormatChannelArchivalMessage(t *testing.T) {
	mockAPI := NewMockSlackAPI()
	client, _ := NewClientWithAPI(mockAPI)

	channel := Channel{
		ID:      "C1234567890",
		Name:    "test-channel",
		Created: time.Now().Add(-24 * time.Hour),
		Purpose: "Test channel purpose",
		Creator: "U1111111",
	}

	t.Run("Standard archival message", func(t *testing.T) {
		message := client.FormatChannelArchivalMessage(channel, 300, 60) // 5 minutes warn, 1 minute archive

		expectedParts := []string{
			"📋 **Channel Archival Notice**",
			"This channel is being archived",
			"inactive for more than",
			"slack-buddy bot",
		}

		for _, part := range expectedParts {
			assert.Contains(t, message, part, "Archival message should contain: %s", part)
		}
	})
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int
		expected string
	}{
		{"Less than minute", 45, "45 seconds"},
		{"Exactly one minute", 60, "1 minute"},
		{"Multiple minutes", 300, "5 minutes"},
		{"Exactly one hour", 3600, "1 hour"},
		{"Multiple hours", 7200, "2 hours"},
		{"Days", 86400, "1 day"},
		{"Multiple days", 172800, "2 days"},
	}

	// We can't test formatDuration directly since it's not exported,
	// but we can test it through FormatInactiveChannelWarning
	mockAPI := NewMockSlackAPI()
	client, _ := NewClientWithAPI(mockAPI)

	channel := Channel{ID: "C123", Name: "test"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warning := client.FormatInactiveChannelWarning(channel, tt.seconds, 60)
			assert.Contains(t, warning, tt.expected)
		})
	}
}

func TestCheckOAuthScopes(t *testing.T) {
	t.Run("Success - all scopes available", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		scopes, err := client.CheckOAuthScopes()
		assert.NoError(t, err)
		assert.NotNil(t, scopes)

		// Check that all expected scopes are tested
		expectedScopes := []string{"channels:read", "channels:join", "chat:write", "channels:manage", "users:read", "groups:read"}
		for _, scope := range expectedScopes {
			_, exists := scopes[scope]
			assert.True(t, exists, "Scope %s should be tested", scope)
		}
	})

	t.Run("Auth test failure", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.SetAuthError(true)
		client, clientErr := NewClientWithAPI(mockAPI)

		// Client creation should fail due to auth error
		assert.Error(t, clientErr)
		assert.Nil(t, client)
		assert.Contains(t, clientErr.Error(), "authentication failed")
	})
}

func TestScopeTestFunctions(t *testing.T) {
	mockAPI := NewMockSlackAPI()
	client, _ := NewClientWithAPI(mockAPI)

	t.Run("testChannelsReadScope - success", func(t *testing.T) {
		result := client.testChannelsReadScope()
		assert.True(t, result)
	})

	t.Run("testChannelsReadScope - missing scope", func(t *testing.T) {
		mockAPI.SetGetConversationsErrorWithMessage(true, "missing_scope")
		result := client.testChannelsReadScope()
		assert.False(t, result)

		// Reset for other tests
		mockAPI.SetGetConversationsErrorWithMessage(false, "")
	})

	t.Run("testChannelsJoinScope - always returns true", func(t *testing.T) {
		result := client.testChannelsJoinScope()
		assert.True(t, result)
	})

	t.Run("testChatWriteScope - always returns true", func(t *testing.T) {
		result := client.testChatWriteScope()
		assert.True(t, result)
	})

	t.Run("testChannelsManageScope - always returns true", func(t *testing.T) {
		result := client.testChannelsManageScope()
		assert.True(t, result)
	})

	t.Run("testUsersReadScope - always returns true", func(t *testing.T) {
		result := client.testUsersReadScope()
		assert.True(t, result)
	})

	t.Run("testGroupsReadScope - always returns true", func(t *testing.T) {
		result := client.testGroupsReadScope()
		assert.True(t, result)
	})
}

func TestGetInactiveChannels(t *testing.T) {
	t.Run("No inactive channels", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Add an active channel (created 1 hour ago, no messages but recent creation)
		now := time.Now()
		recentTime := now.Add(-1 * time.Hour)
		mockAPI.AddChannel("C1234567890", "active-channel", recentTime, "Active channel")

		// Check for inactive channels (warn after 2 hours, archive after 1 hour)
		toWarn, toArchive, err := client.GetInactiveChannels(7200, 3600) // 2 hours warn, 1 hour archive

		assert.NoError(t, err)
		assert.Len(t, toWarn, 0)
		assert.Len(t, toArchive, 0)
	})

	t.Run("Channel needing warning", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Add a channel created 3 hours ago (old enough to be inactive)
		now := time.Now()
		oldTime := now.Add(-3 * time.Hour)
		mockAPI.AddChannel("C1234567890", "inactive-channel", oldTime, "Inactive channel")

		// Mock conversation history to show old activity
		mockAPI.SetChannelHistory("C1234567890", []MockHistoryMessage{
			{
				Timestamp: formatTimestamp(oldTime.Add(10 * time.Minute)),
				User:      "U1234567",
				Text:      "Last message",
			},
		})

		// Check for inactive channels (warn after 2 hours, archive after 1 hour)
		toWarn, toArchive, err := client.GetInactiveChannels(7200, 3600) // 2 hours warn, 1 hour archive

		assert.NoError(t, err)
		assert.Len(t, toWarn, 1)
		assert.Len(t, toArchive, 0)
		assert.Equal(t, "inactive-channel", toWarn[0].Name)
	})

	t.Run("Channel needing archival after warning", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Add a channel created 4 hours ago
		now := time.Now()
		oldTime := now.Add(-4 * time.Hour)
		warningTime := now.Add(-2 * time.Hour)
		mockAPI.AddChannel("C1234567890", "to-archive-channel", oldTime, "Channel to archive")

		// Mock conversation history showing:
		// 1. Old user activity
		// 2. Bot warning message (using correct format that matches FormatInactiveChannelWarning)
		// 3. No activity after warning
		// Note: Mock stores messages oldest-first, then reverses them to newest-first for API
		mockAPI.SetChannelHistory("C1234567890", []MockHistoryMessage{
			// Oldest first
			{
				Timestamp: formatTimestamp(oldTime.Add(30 * time.Minute)),
				User:      "U1234567",
				Text:      "Last real user message",
			},
			{
				Timestamp: formatTimestamp(warningTime),
				User:      "UBOT123456",
				Text:      "🚨 **Inactive Channel Warning** 🚨\n\n<!-- inactive channel warning -->",
			},
		})

		// Set the bot user ID to match our warning message
		mockAPI.SetBotUserID("UBOT123456")

		// Check for inactive channels (warn after 3 hours, archive after 1 hour from warning)
		// Since warning was 2 hours ago, and archive threshold is 1 hour, this should be archived
		toWarn, toArchive, err := client.GetInactiveChannels(10800, 3600) // 3 hours warn, 1 hour archive

		assert.NoError(t, err)
		assert.Len(t, toWarn, 0)
		assert.Len(t, toArchive, 1)
		assert.Equal(t, "to-archive-channel", toArchive[0].Name)
	})

	t.Run("Excluded channels are skipped", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Add channels that should be excluded by default
		now := time.Now()
		oldTime := now.Add(-5 * time.Hour)

		// These should be excluded by shouldSkipChannel
		mockAPI.AddChannel("C1111111111", "general", oldTime, "General channel")
		mockAPI.AddChannel("C2222222222", "random", oldTime, "Random channel")
		mockAPI.AddChannel("C3333333333", "announcements", oldTime, "Announcements")
		mockAPI.AddChannel("C4444444444", "admin-tools", oldTime, "Admin tools")

		// This should not be excluded
		mockAPI.AddChannel("C5555555555", "old-project", oldTime, "Old project channel")

		// Mock all as having old activity
		for _, channelID := range []string{"C1111111111", "C2222222222", "C3333333333", "C4444444444", "C5555555555"} {
			mockAPI.SetChannelHistory(channelID, []MockHistoryMessage{
				{
					Timestamp: formatTimestamp(oldTime.Add(10 * time.Minute)),
					User:      "U1234567",
					Text:      "Old message",
				},
			})
		}

		// Check for inactive channels (warn after 3 hours, archive after 1 hour)
		toWarn, _, err := client.GetInactiveChannels(10800, 3600)

		assert.NoError(t, err)
		assert.Len(t, toWarn, 1) // Only old-project should be warned
		assert.Equal(t, "old-project", toWarn[0].Name)
	})

	t.Run("API error handling", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Configure mock to return error on GetConversations
		mockAPI.SetGetConversationsError(true)

		toWarn, toArchive, err := client.GetInactiveChannels(7200, 3600)

		assert.Error(t, err)
		assert.Nil(t, toWarn)
		assert.Nil(t, toArchive)
		assert.Contains(t, err.Error(), "failed to get conversations")
	})
}

func TestGetChannelActivity(t *testing.T) {
	t.Run("Channel with recent user activity", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		now := time.Now()
		recentTime := now.Add(-30 * time.Minute)

		// Mock conversation history with recent user message
		mockAPI.SetChannelHistory("C1234567890", []MockHistoryMessage{
			{
				Timestamp: formatTimestamp(recentTime),
				User:      "U1234567",
				Text:      "Recent user message",
			},
		})

		lastActivity, hasWarning, warningTime, err := client.GetChannelActivity("C1234567890")

		assert.NoError(t, err)
		assert.False(t, hasWarning)
		assert.True(t, warningTime.IsZero())
		// Activity time should be close to our recent time (within a few seconds)
		assert.WithinDuration(t, recentTime, lastActivity, 5*time.Second)
	})

	t.Run("Channel with bot warning message", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		now := time.Now()
		oldUserTime := now.Add(-4 * time.Hour)
		warningTime := now.Add(-1 * time.Hour)

		warningText := "🚨 **Inactive Channel Warning** 🚨\n\n<!-- inactive channel warning -->"

		// Mock conversation history showing warning from bot (using correct format)
		// Note: Mock stores messages oldest-first, then reverses them to newest-first for API
		mockAPI.SetChannelHistory("C1234567890", []MockHistoryMessage{
			// Oldest first
			{
				Timestamp: formatTimestamp(oldUserTime),
				User:      "U1234567",
				Text:      "Old user message",
			},
			{
				Timestamp: formatTimestamp(warningTime),
				User:      "UBOT123456",
				Text:      warningText,
			},
		})

		mockAPI.SetBotUserID("UBOT123456")

		lastActivity, hasWarning, actualWarningTime, err := client.GetChannelActivity("C1234567890")

		assert.NoError(t, err)
		assert.True(t, hasWarning)
		assert.WithinDuration(t, warningTime, actualWarningTime, 5*time.Second)
		assert.WithinDuration(t, oldUserTime, lastActivity, 5*time.Second)
	})

	t.Run("Channel with no messages", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Mock empty conversation history
		mockAPI.SetChannelHistory("C1234567890", []MockHistoryMessage{})

		lastActivity, hasWarning, warningTime, err := client.GetChannelActivity("C1234567890")

		assert.NoError(t, err)
		assert.False(t, hasWarning)
		assert.True(t, warningTime.IsZero())
		// Should return Unix epoch for no activity
		assert.Equal(t, time.Unix(0, 0), lastActivity)
	})

	t.Run("Channel with only system messages", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		now := time.Now()
		systemTime := now.Add(-1 * time.Hour)

		// Mock conversation history with only system messages
		mockAPI.SetChannelHistory("C1234567890", []MockHistoryMessage{
			{
				Timestamp: formatTimestamp(systemTime),
				User:      "",
				Text:      "User joined the channel",
				SubType:   "channel_join",
			},
		})

		lastActivity, hasWarning, warningTime, err := client.GetChannelActivity("C1234567890")

		assert.NoError(t, err)
		assert.False(t, hasWarning)
		assert.True(t, warningTime.IsZero())
		// Should use system message time as fallback when no real messages exist
		assert.WithinDuration(t, systemTime, lastActivity, 5*time.Second)
	})

	t.Run("API error handling", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Configure mock to return error on GetConversationHistory
		mockAPI.SetGetConversationHistoryError("C1234567890", true)

		_, _, _, err := client.GetChannelActivity("C1234567890")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get channel history")
	})
}

func TestSeemsActiveFromMetadata(t *testing.T) {
	client := &Client{}
	cutoff := time.Now().Add(-1 * time.Hour)

	t.Run("Returns true when latest message is after cutoff", func(t *testing.T) {
		recentTime := time.Now().Add(-30 * time.Minute)
		ch := slack.Channel{
			GroupConversation: slack.GroupConversation{
				Conversation: slack.Conversation{
					Latest: &slack.Message{
						Msg: slack.Msg{
							Timestamp: fmt.Sprintf("%.6f", float64(recentTime.Unix())),
						},
					},
				},
			},
		}
		assert.True(t, client.seemsActiveFromMetadata(ch, cutoff))
	})

	t.Run("Returns false when latest message is before cutoff", func(t *testing.T) {
		oldTime := time.Now().Add(-2 * time.Hour)
		ch := slack.Channel{
			GroupConversation: slack.GroupConversation{
				Conversation: slack.Conversation{
					Latest: &slack.Message{
						Msg: slack.Msg{
							Timestamp: fmt.Sprintf("%.6f", float64(oldTime.Unix())),
						},
					},
				},
			},
		}
		assert.False(t, client.seemsActiveFromMetadata(ch, cutoff))
	})

	t.Run("Returns false when Latest is nil", func(t *testing.T) {
		ch := slack.Channel{
			GroupConversation: slack.GroupConversation{
				Conversation: slack.Conversation{
					Latest: nil,
				},
			},
		}
		assert.False(t, client.seemsActiveFromMetadata(ch, cutoff))
	})

	t.Run("Returns false when timestamp parsing fails", func(t *testing.T) {
		ch := slack.Channel{
			GroupConversation: slack.GroupConversation{
				Conversation: slack.Conversation{
					Latest: &slack.Message{
						Msg: slack.Msg{
							Timestamp: "invalid-timestamp",
						},
					},
				},
			},
		}
		assert.False(t, client.seemsActiveFromMetadata(ch, cutoff))
	})

	t.Run("Returns false when timestamp is empty", func(t *testing.T) {
		ch := slack.Channel{
			GroupConversation: slack.GroupConversation{
				Conversation: slack.Conversation{
					Latest: &slack.Message{
						Msg: slack.Msg{
							Timestamp: "",
						},
					},
				},
			},
		}
		assert.False(t, client.seemsActiveFromMetadata(ch, cutoff))
	})

	t.Run("Handles edge case with exactly cutoff time", func(t *testing.T) {
		ch := slack.Channel{
			GroupConversation: slack.GroupConversation{
				Conversation: slack.Conversation{
					Latest: &slack.Message{
						Msg: slack.Msg{
							Timestamp: fmt.Sprintf("%.6f", float64(cutoff.Unix())),
						},
					},
				},
			},
		}
		assert.False(t, client.seemsActiveFromMetadata(ch, cutoff))
	})

	t.Run("Handles fractional timestamps", func(t *testing.T) {
		recentTime := time.Now().Add(-30 * time.Minute)
		fractionalTimestamp := fmt.Sprintf("%.6f", float64(recentTime.Unix())+0.123456)
		ch := slack.Channel{
			GroupConversation: slack.GroupConversation{
				Conversation: slack.Conversation{
					Latest: &slack.Message{
						Msg: slack.Msg{
							Timestamp: fractionalTimestamp,
						},
					},
				},
			},
		}
		assert.True(t, client.seemsActiveFromMetadata(ch, cutoff))
	})
}

func TestWarnInactiveChannel(t *testing.T) {
	t.Run("Successfully warn inactive channel", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		channel := Channel{
			ID:   "C1234567890",
			Name: "inactive-channel",
		}

		err := client.WarnInactiveChannel(channel, 7200, 3600) // 2 hours warn, 1 hour archive

		assert.NoError(t, err)

		// Verify that the bot joined the channel
		assert.Contains(t, mockAPI.JoinedChannels, "C1234567890")

		// Verify that a message was posted
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 1)
		assert.Equal(t, "C1234567890", messages[0].ChannelID)
	})

	t.Run("Failed to join channel", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Configure mock to fail joining
		mockAPI.SetJoinConversationErrorForChannel("C1234567890", true)

		channel := Channel{
			ID:   "C1234567890",
			Name: "private-channel",
		}

		err := client.WarnInactiveChannel(channel, 7200, 3600)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to join channel")

		// Should not have posted message if join failed
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 0)
	})
}

func TestArchiveChannel(t *testing.T) {
	t.Run("Successfully archive channel", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		channel := Channel{
			ID:   "C1234567890",
			Name: "to-archive",
		}

		err := client.ArchiveChannelWithThresholds(channel, 7200, 3600)

		assert.NoError(t, err)

		// Verify that the bot joined the channel
		assert.Contains(t, mockAPI.JoinedChannels, "C1234567890")

		// Verify that archival message was posted
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 1)
		assert.Equal(t, "C1234567890", messages[0].ChannelID)

		// Verify that the channel was archived
		assert.Contains(t, mockAPI.ArchivedChannels, "C1234567890")
	})

	t.Run("Archive succeeds even if join fails", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Configure mock to fail joining but allow archiving
		mockAPI.SetJoinConversationErrorForChannel("C1234567890", true)

		channel := Channel{
			ID:   "C1234567890",
			Name: "private-to-archive",
		}

		err := client.ArchiveChannelWithThresholds(channel, 7200, 3600)

		assert.NoError(t, err)

		// Should still have archived the channel even if join/message failed
		assert.Contains(t, mockAPI.ArchivedChannels, "C1234567890")
	})

	t.Run("Missing scope error", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Configure mock to return missing scope error for archive
		mockAPI.SetArchiveConversationErrorWithMessage("C1234567890", true, "missing_scope")

		channel := Channel{
			ID:   "C1234567890",
			Name: "test-channel",
		}

		err := client.ArchiveChannelWithThresholds(channel, 7200, 3600)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required permission to archive channels")
		assert.Contains(t, err.Error(), "channels:manage")
	})

	t.Run("Already archived channel", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Configure mock to return already archived error
		mockAPI.SetArchiveConversationErrorWithMessage("C1234567890", true, "already_archived")

		channel := Channel{
			ID:   "C1234567890",
			Name: "test-channel",
		}

		err := client.ArchiveChannelWithThresholds(channel, 7200, 3600)

		assert.NoError(t, err) // Should not error for already archived
	})
}

// Helper function to format timestamp for mock messages
func formatTimestamp(t time.Time) string {
	return fmt.Sprintf("%d.%06d", t.Unix(), t.Nanosecond()/1000)
}

// Test formatDuration function - direct testing of the function
func TestFormatDurationDirect(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"Zero duration", 0, "0 seconds"},
		{"One second", time.Second, "1 second"},
		{"One minute exactly", time.Minute, "1 minute"},
		{"One hour exactly", time.Hour, "1 hour"},
		{"One day exactly", 24 * time.Hour, "1 day"},
		{"Complex duration (1h30m)", time.Hour + 30*time.Minute, "1 hour"},
		{"Very small duration", 500 * time.Millisecond, "0 seconds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test parseSlackRetryAfter function
func TestParseSlackRetryAfter(t *testing.T) {
	tests := []struct {
		name     string
		errorStr string
		expected time.Duration
	}{
		{"Valid retry after seconds", "Error: retry after 30s", 31 * time.Second}, // 30s + 1s buffer
		{"Valid retry after minutes", "Error: retry after 2m", 2*time.Minute + 1*time.Second},
		{"Valid retry after with text", "Rate limited, retry after 1m30s", 1*time.Minute + 30*time.Second + 1*time.Second},
		{"No retry after in string", "Generic error message", 0},
		{"Empty string", "", 0},
		{"Invalid duration format", "retry after invalid", 0},
		{"Multiple retry after (first wins)", "retry after 10s and retry after 20s", 11 * time.Second},
		{"Retry after with extra text", "API error: retry after 45s due to rate limiting", 46 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSlackRetryAfter(tt.errorStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test showProgressBar function
func TestShowProgressBar(t *testing.T) {
	// Capture output for testing - this is a UI function, so we test basic behavior
	// without mocking complex console output since it's primarily for user feedback
	t.Run("Zero duration returns immediately", func(t *testing.T) {
		// Should return immediately without any output or panic
		start := time.Now()
		showProgressBar(0)
		elapsed := time.Since(start)

		// Should complete almost instantly (within 10ms)
		assert.Less(t, elapsed, 10*time.Millisecond)
	})

	t.Run("Negative duration returns immediately", func(t *testing.T) {
		// Should return immediately without any output or panic
		start := time.Now()
		showProgressBar(-5 * time.Second)
		elapsed := time.Since(start)

		// Should complete almost instantly (within 10ms)
		assert.Less(t, elapsed, 10*time.Millisecond)
	})

	t.Run("Very short duration works without panic", func(t *testing.T) {
		// Test with 1 millisecond - should not panic and complete quickly
		// since it gets rounded to 0 seconds in the function
		start := time.Now()
		showProgressBar(1 * time.Millisecond)
		elapsed := time.Since(start)

		// Should complete almost instantly since duration < 1 second
		assert.Less(t, elapsed, 10*time.Millisecond)
	})

	// Note: We don't test actual progress bar output since it involves:
	// - Real-time sleep operations (would slow down tests significantly)
	// - Complex console output formatting that's hard to capture reliably
	// - User interface behavior that's better verified through manual testing
	// The function's core logic (duration handling, bounds checking) is covered above
}

// Test getUserMap function with mock API
func TestGetUserMapUtility(t *testing.T) {
	t.Run("Success with users", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Add some test users to mock
		mockAPI.AddUser("U123", "john.doe", "John Doe") // Has RealName - should use "John Doe"
		mockAPI.AddUser("U456", "jane.smith", "")       // No RealName, has Name - should use "jane.smith"
		mockAPI.AddUser("U789", "", "")                 // No RealName or Name, DisplayName will be "" - should use ID
		mockAPI.AddUser("U999", "", "")                 // Only ID available - should use "U999"

		userMap, err := client.GetUserMap()
		assert.NoError(t, err)
		assert.Len(t, userMap, 4)

		// Test name priority (RealName > Name > Profile.DisplayName > ID)
		assert.Equal(t, "John Doe", userMap["U123"])   // Uses RealName
		assert.Equal(t, "jane.smith", userMap["U456"]) // Uses Name (no RealName)
		assert.Equal(t, "U789", userMap["U789"])       // Falls back to ID (no RealName, Name, or DisplayName)
		assert.Equal(t, "U999", userMap["U999"])       // Falls back to ID
	})

	t.Run("Missing scope error", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.SetGetUsersError("missing_scope")
		client, _ := NewClientWithAPI(mockAPI)

		userMap, err := client.GetUserMap()
		assert.Error(t, err)
		assert.Nil(t, userMap)
		assert.Contains(t, err.Error(), "missing_scope")
	})

	t.Run("Empty user list", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)
		// Users slice is empty by default

		userMap, err := client.GetUserMap()
		assert.NoError(t, err)
		assert.Empty(t, userMap)
	})

	t.Run("Generic error", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.SetGetUsersError("api_error")
		client, _ := NewClientWithAPI(mockAPI)

		userMap, err := client.GetUserMap()
		assert.Error(t, err)
		assert.Nil(t, userMap)
		assert.Contains(t, err.Error(), "api_error")
	})
}

func TestGetInactiveChannelsWithDetails(t *testing.T) {
	t.Run("Basic functionality", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Add channels with different activity levels
		now := time.Now()
		inactiveTime := now.Add(-30 * time.Minute) // 30 minutes ago
		activeTime := now.Add(-2 * time.Minute)    // 2 minutes ago

		// Add an inactive channel (created 1 hour ago, last message 30 min ago)
		mockAPI.AddChannelWithCreator("C001", "inactive-channel", now.Add(-1*time.Hour), "U123", "Inactive test channel")
		mockAPI.SetChannelHistory("C001", []MockHistoryMessage{
			{Timestamp: fmt.Sprintf("%.6f", float64(inactiveTime.Unix())), User: "U123", Text: "Old message"},
		})

		// Add an active channel (created 1 hour ago, last message 2 min ago)
		mockAPI.AddChannelWithCreator("C002", "active-channel", now.Add(-1*time.Hour), "U123", "Active test channel")
		mockAPI.SetChannelHistory("C002", []MockHistoryMessage{
			{Timestamp: fmt.Sprintf("%.6f", float64(activeTime.Unix())), User: "U123", Text: "Recent message"},
		})

		// Test with 10 minute warn threshold, 5 minute archive threshold
		userMap := map[string]string{"U123": "testuser"}
		toWarn, toArchive, err := client.GetInactiveChannelsWithDetails(600, 300, userMap, false) // 10min warn, 5min archive

		assert.NoError(t, err)
		// inactive-channel should be in toWarn (30 min > 10 min threshold)
		assert.Len(t, toWarn, 1)
		assert.Equal(t, "inactive-channel", toWarn[0].Name)
		// toArchive should be empty since channel doesn't have existing warning
		assert.Len(t, toArchive, 0)
	})

	t.Run("API error handling", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.SetGetConversationsError(true)
		client, _ := NewClientWithAPI(mockAPI)

		userMap := map[string]string{}
		toWarn, toArchive, err := client.GetInactiveChannelsWithDetails(600, 300, userMap, false)

		assert.Error(t, err)
		assert.Nil(t, toWarn)
		assert.Nil(t, toArchive)
		assert.Contains(t, err.Error(), "failed to get conversations")
	})
}

func TestGetInactiveChannelsWithDetailsAndExclusions(t *testing.T) {
	t.Run("Channel exclusions", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		// Add channels
		now := time.Now()
		inactiveTime := now.Add(-30 * time.Minute)

		// Add channels that should be excluded
		mockAPI.AddChannelWithCreator("C001", "general", now.Add(-1*time.Hour), "U123", "General channel")
		mockAPI.AddChannelWithCreator("C002", "prod-alerts", now.Add(-1*time.Hour), "U123", "Production alerts")
		mockAPI.AddChannelWithCreator("C003", "test-channel", now.Add(-1*time.Hour), "U123", "Test channel")

		// Set all as inactive
		for _, channelID := range []string{"C001", "C002", "C003"} {
			mockAPI.SetChannelHistory(channelID, []MockHistoryMessage{
				{Timestamp: fmt.Sprintf("%.6f", float64(inactiveTime.Unix())), User: "U123", Text: "Old message"},
			})
		}

		userMap := map[string]string{"U123": "testuser"}
		excludeChannels := []string{"general"}
		excludePrefixes := []string{"prod-"}

		toWarn, toArchive, err := client.GetInactiveChannelsWithDetailsAndExclusions(600, 300, userMap, excludeChannels, excludePrefixes, false)

		assert.NoError(t, err)
		// Only test-channel should remain (general excluded by name, prod-alerts excluded by prefix)
		assert.Len(t, toWarn, 1)
		assert.Equal(t, "test-channel", toWarn[0].Name)
		// toArchive should be empty since channel doesn't have existing warning
		assert.Len(t, toArchive, 0)
	})
}

func TestArchiveChannelLegacyMethod(t *testing.T) {
	t.Run("Legacy ArchiveChannel method", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		channel := Channel{
			ID:   "C1234567890",
			Name: "legacy-archive-test",
		}

		// Test the legacy method (should call ArchiveChannelWithThresholds with default values)
		err := client.ArchiveChannel(channel)

		assert.NoError(t, err)

		// Verify that the channel was archived
		assert.Contains(t, mockAPI.ArchivedChannels, "C1234567890")

		// Verify that the bot joined the channel
		assert.Contains(t, mockAPI.JoinedChannels, "C1234567890")

		// Verify that archival message was posted
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 1)
		assert.Equal(t, "C1234567890", messages[0].ChannelID)
	})
}

func TestGetChannelsWithMetadata(t *testing.T) {
	t.Run("Success with channels", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		now := time.Now()
		mockAPI.AddChannel("C123", "test-channel", now.Add(-1*time.Hour), "Test channel")
		mockAPI.AddChannel("C456", "another-channel", now.Add(-2*time.Hour), "Another channel")

		client, _ := NewClientWithAPI(mockAPI)

		channels, err := client.GetChannelsWithMetadata()

		assert.NoError(t, err)
		assert.Len(t, channels, 2)

		// Verify channel details
		channelNames := make(map[string]bool)
		for _, ch := range channels {
			channelNames[ch.Name] = true
		}
		assert.True(t, channelNames["test-channel"])
		assert.True(t, channelNames["another-channel"])
	})

	t.Run("API error", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.SetGetConversationsError(true)
		client, _ := NewClientWithAPI(mockAPI)

		channels, err := client.GetChannelsWithMetadata()

		assert.Error(t, err)
		assert.Nil(t, channels)
		assert.Contains(t, err.Error(), "failed to get conversations")
	})

	t.Run("No channels", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		channels, err := client.GetChannelsWithMetadata()

		assert.NoError(t, err)
		assert.Len(t, channels, 0)
	})
}

func TestCheckForDuplicateAnnouncement(t *testing.T) {
	t.Run("No duplicate when no previous messages", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.AddChannel("C123", "general", time.Now().Add(-24*time.Hour), "General discussion")
		client, _ := NewClientWithAPI(mockAPI)

		isDuplicate, err := client.CheckForDuplicateAnnouncement("#general", "New channel alert! #test-channel", []string{"test-channel"})

		assert.NoError(t, err)
		assert.False(t, isDuplicate)
	})

	t.Run("Detects duplicate when same channel was already announced", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.AddChannel("C123", "general", time.Now().Add(-24*time.Hour), "General discussion")

		// Add a previous announcement from the bot (U0000000 is bot's user ID)
		mockAPI.AddMessageToHistory("C123", "New channel alert! #test-channel", "U0000000", fmt.Sprintf("%.6f", float64(time.Now().Add(-1*time.Hour).Unix())))

		client, _ := NewClientWithAPI(mockAPI)

		isDuplicate, err := client.CheckForDuplicateAnnouncement("#general", "New channel alert! #test-channel", []string{"test-channel"})

		assert.NoError(t, err)
		assert.True(t, isDuplicate)
	})

	t.Run("No duplicate when different channel was announced", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.AddChannel("C123", "general", time.Now().Add(-24*time.Hour), "General discussion")

		// Add a previous announcement for a different channel
		mockAPI.AddMessageToHistory("C123", "New channel alert! #other-channel", "U0000000", fmt.Sprintf("%.6f", float64(time.Now().Add(-1*time.Hour).Unix())))

		client, _ := NewClientWithAPI(mockAPI)

		isDuplicate, err := client.CheckForDuplicateAnnouncement("#general", "New channel alert! #test-channel", []string{"test-channel"})

		assert.NoError(t, err)
		assert.False(t, isDuplicate)
	})

	t.Run("Ignores messages from other users", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.AddChannel("C123", "general", time.Now().Add(-24*time.Hour), "General discussion")

		// Add a message from a different user (not the bot)
		mockAPI.AddMessageToHistory("C123", "New channel alert! #test-channel", "U1234567", fmt.Sprintf("%.6f", float64(time.Now().Add(-1*time.Hour).Unix())))

		client, _ := NewClientWithAPI(mockAPI)

		isDuplicate, err := client.CheckForDuplicateAnnouncement("#general", "New channel alert! #test-channel", []string{"test-channel"})

		assert.NoError(t, err)
		assert.False(t, isDuplicate)
	})

	t.Run("Returns false when channel not found", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		isDuplicate, err := client.CheckForDuplicateAnnouncement("#nonexistent", "New channel alert! #test-channel", []string{"test-channel"})

		assert.Error(t, err)
		assert.False(t, isDuplicate)
		assert.Contains(t, err.Error(), "failed to find channel")
	})

	t.Run("Returns false when history API fails", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.AddChannel("C123", "general", time.Now().Add(-24*time.Hour), "General discussion")
		mockAPI.SetGetConversationHistoryError("C123", true)

		client, _ := NewClientWithAPI(mockAPI)

		isDuplicate, err := client.CheckForDuplicateAnnouncement("#general", "New channel alert! #test-channel", []string{"test-channel"})

		assert.NoError(t, err)
		assert.False(t, isDuplicate)
	})
}

func TestCheckForDuplicateAnnouncementWithDetails(t *testing.T) {
	t.Run("Returns detailed information about duplicate channels", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.AddChannel("C123", "general", time.Now().Add(-24*time.Hour), "General discussion")

		// Add previous announcements for multiple channels
		mockAPI.AddMessageToHistory("C123", "New channel alert! #channel1 #channel2", "U0000000", fmt.Sprintf("%.6f", float64(time.Now().Add(-2*time.Hour).Unix())))
		mockAPI.AddMessageToHistory("C123", "New channel created: #channel3", "U0000000", fmt.Sprintf("%.6f", float64(time.Now().Add(-1*time.Hour).Unix())))

		client, _ := NewClientWithAPI(mockAPI)
		cutoffTime := time.Now().Add(-24 * time.Hour)

		isDuplicate, duplicateChannels, err := client.CheckForDuplicateAnnouncementWithDetails("#general", "New channel alert! #channel1 #channel4", []string{"channel1", "channel4"}, cutoffTime)

		assert.NoError(t, err)
		assert.True(t, isDuplicate)
		assert.Contains(t, duplicateChannels, "channel1")
		assert.NotContains(t, duplicateChannels, "channel4")
	})

	t.Run("Currently does not filter by cutoff time (known limitation)", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.AddChannel("C123", "general", time.Now().Add(-24*time.Hour), "General discussion")

		// Add an old announcement beyond cutoff time
		oldTimestamp := time.Now().Add(-25 * time.Hour)
		mockAPI.AddMessageToHistory("C123", "New channel alert! #test-channel", "U0000000", fmt.Sprintf("%.6f", float64(oldTimestamp.Unix())))

		client, _ := NewClientWithAPI(mockAPI)
		cutoffTime := time.Now().Add(-24 * time.Hour)

		isDuplicate, duplicateChannels, err := client.CheckForDuplicateAnnouncementWithDetails("#general", "New channel alert! #test-channel", []string{"test-channel"}, cutoffTime)

		assert.NoError(t, err)
		// NOTE: Current implementation does not filter by cutoff time - this is a known limitation
		// The test documents the current behavior rather than the intended behavior
		assert.True(t, isDuplicate)
		assert.Contains(t, duplicateChannels, "test-channel")
	})
}

func TestCheckForDuplicateAnnouncementWithDetailsAndChannels(t *testing.T) {
	t.Run("Uses provided channel list when available", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.AddChannel("C123", "general", time.Now().Add(-24*time.Hour), "General discussion")

		// Add previous announcement
		mockAPI.AddMessageToHistory("C123", "New channel alert! #test-channel", "U0000000", fmt.Sprintf("%.6f", float64(time.Now().Add(-1*time.Hour).Unix())))

		client, _ := NewClientWithAPI(mockAPI)
		cutoffTime := time.Now().Add(-24 * time.Hour)

		// Provide pre-fetched channel list
		allChannels := []slack.Channel{
			{
				GroupConversation: slack.GroupConversation{
					Conversation: slack.Conversation{ID: "C456"},
					Name:         "test-channel",
				},
			},
		}

		isDuplicate, duplicateChannels, err := client.CheckForDuplicateAnnouncementWithDetailsAndChannels("#general", "New channel alert! #test-channel", []string{"test-channel"}, cutoffTime, allChannels)

		assert.NoError(t, err)
		assert.True(t, isDuplicate)
		assert.Contains(t, duplicateChannels, "test-channel")
	})

	t.Run("Fetches channel list when not provided", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.AddChannel("C123", "general", time.Now().Add(-24*time.Hour), "General discussion")
		mockAPI.AddChannel("C456", "test-channel", time.Now().Add(-2*time.Hour), "Test channel")

		// Add previous announcement
		mockAPI.AddMessageToHistory("C123", "New channel alert! #test-channel", "U0000000", fmt.Sprintf("%.6f", float64(time.Now().Add(-1*time.Hour).Unix())))

		client, _ := NewClientWithAPI(mockAPI)
		cutoffTime := time.Now().Add(-24 * time.Hour)

		isDuplicate, duplicateChannels, err := client.CheckForDuplicateAnnouncementWithDetailsAndChannels("#general", "New channel alert! #test-channel", []string{"test-channel"}, cutoffTime, nil)

		assert.NoError(t, err)
		assert.True(t, isDuplicate)
		assert.Contains(t, duplicateChannels, "test-channel")
	})
}

func TestGetAllChannelNameToIDMap(t *testing.T) {
	t.Run("Creates correct name to ID mapping", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.AddChannel("C123", "general", time.Now().Add(-24*time.Hour), "General discussion")
		mockAPI.AddChannel("C456", "random", time.Now().Add(-12*time.Hour), "Random stuff")

		client, _ := NewClientWithAPI(mockAPI)

		channelMap, err := client.getAllChannelNameToIDMap()

		assert.NoError(t, err)
		assert.Equal(t, "C123", channelMap["general"])
		assert.Equal(t, "C456", channelMap["random"])
		assert.Len(t, channelMap, 2)
	})

	t.Run("Returns error when API fails", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		mockAPI.SetGetConversationsError(true)

		client, _ := NewClientWithAPI(mockAPI)

		channelMap, err := client.getAllChannelNameToIDMap()

		assert.Error(t, err)
		assert.Nil(t, channelMap)
	})

	t.Run("Handles empty channel list", func(t *testing.T) {
		mockAPI := NewMockSlackAPI()
		client, _ := NewClientWithAPI(mockAPI)

		channelMap, err := client.getAllChannelNameToIDMap()

		assert.NoError(t, err)
		assert.Len(t, channelMap, 0)
	})
}
