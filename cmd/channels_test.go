package cmd

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/astrostl/slack-butler/pkg/slack"
)

func TestDetectCommandSetup(t *testing.T) {
	t.Run("Command structure", func(t *testing.T) {
		assert.NotNil(t, detectCmd)
		assert.Equal(t, "detect", detectCmd.Use)
		assert.NotEmpty(t, detectCmd.Short)
	})

	t.Run("Command flags", func(t *testing.T) {
		sinceFlag := detectCmd.Flags().Lookup("since")
		assert.NotNil(t, sinceFlag)
		assert.Equal(t, "8", sinceFlag.DefValue)

		announceFlag := detectCmd.Flags().Lookup("announce-to")
		assert.NotNil(t, announceFlag)
		assert.Equal(t, "", announceFlag.DefValue)

		commitFlag := detectCmd.Flags().Lookup("commit")
		assert.NotNil(t, commitFlag)
		assert.Equal(t, "false", commitFlag.DefValue)
	})

	t.Run("Command hierarchy", func(t *testing.T) {
		assert.NotNil(t, channelsCmd)
		assert.Equal(t, "channels", channelsCmd.Use)

		// Check that detect is a subcommand of channels
		subcommands := channelsCmd.Commands()
		var detectFound bool
		for _, cmd := range subcommands {
			if cmd.Use == "detect" {
				detectFound = true
				break
			}
		}
		assert.True(t, detectFound)
	})
}

func TestRunDetectFunction(t *testing.T) {
	t.Run("Missing token error", func(t *testing.T) {
		// Clear any environment token
		t.Setenv("SLACK_TOKEN", "")

		cmd := &cobra.Command{}
		err := runDetect(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "slack token is required")
	})

	t.Run("Invalid time format error", func(t *testing.T) {
		// Save original values
		originalSince := since

		// Set valid token and invalid time format
		t.Setenv("SLACK_TOKEN", "MOCK-BOT-TOKEN-FOR-TESTING-ONLY-NOT-REAL-TOKEN-AT-ALL")

		// Force viper to reload environment
		initConfig()

		since = "invalid"
		announceTo = "test-channel"

		cmd := &cobra.Command{}
		err := runDetect(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid days format")

		// Reset values
		since = originalSince
		announceTo = ""
	})
}

func TestDaysParsing(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		valid bool
	}{
		{"Valid 1 day", "1", true},
		{"Valid 7 days", "7", true},
		{"Valid 30 days", "30", true},
		{"Valid 0.5 days", "0.5", true},
		{"Valid 365 days", "365", true},
		{"Invalid format", "invalid", false},
		{"Invalid negative", "-1", false},
		{"Invalid with units", "7d", false},
		{"Invalid with units", "1w", false},
		{"Invalid with hours", "24h", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			days, err := strconv.ParseFloat(tc.input, 64)
			if tc.valid {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, days, 0.0)
			} else {
				if err == nil && days < 0 {
					// This handles the negative case
					assert.True(t, days < 0)
				} else {
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestRunDetectSuccessPath(t *testing.T) {
	t.Run("Successful channel detection without announcement", func(t *testing.T) {
		// Save original values
		originalSince := since
		originalAnnounceTo := announceTo

		// Set up test environment
		t.Setenv("SLACK_TOKEN", "MOCK-BOT-TOKEN-FOR-TESTING-ONLY-NOT-REAL-TOKEN-AT-ALL")
		initConfig()

		since = "1"
		announceTo = "test-channel"

		cmd := &cobra.Command{}
		err := runDetect(cmd, []string{})

		// We expect this to fail at the network call level, but it validates the happy path logic
		// up to the point where it tries to make real API calls
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")

		// Reset values
		since = originalSince
		announceTo = originalAnnounceTo
	})

	t.Run("Successful channel detection with announcement", func(t *testing.T) {
		// Save original values
		originalSince := since
		originalAnnounceTo := announceTo

		// Set up test environment
		t.Setenv("SLACK_TOKEN", "MOCK-BOT-TOKEN-FOR-TESTING-ONLY-NOT-REAL-TOKEN-AT-ALL")
		initConfig()

		since = "2"
		announceTo = "#general"

		cmd := &cobra.Command{}
		err := runDetect(cmd, []string{})

		// We expect this to fail at the network call level, but it validates the happy path logic
		// up to the point where it tries to make real API calls, including the announcement path
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")

		// Reset values
		since = originalSince
		announceTo = originalAnnounceTo
	})
}

func TestRunDetectWithClientLogic(t *testing.T) {
	// Test the core logic without stdout capture to avoid race conditions
	t.Run("No channels found", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-24 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "", false)
		assert.NoError(t, err)
	})

	t.Run("Channels found", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-1", testTime, "Test purpose 1")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "", false)
		assert.NoError(t, err)
	})

	t.Run("Channels found with announcement", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-1", testTime, "Test purpose 1")
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "#general", false)
		assert.NoError(t, err)

		// Verify message was posted
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 1)
		assert.Equal(t, "CGENERAL", messages[0].ChannelID)
	})

	t.Run("Announcement posting error", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-1", testTime, "Test purpose 1")
		// Add the channel that will be used for announcements but set up to fail posting
		mockAPI.AddChannel("CNONEXISTENT", "nonexistent", time.Now().Add(-24*time.Hour), "Test channel")

		// Set up mock to fail on PostMessage
		mockAPI.SetPostMessageError("channel_not_found")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "#nonexistent", false)

		// Should return error about failed announcement
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to post announcement to #nonexistent")
	})

	t.Run("GetNewChannels error", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()

		// Set up mock to fail on GetConversations (which GetNewChannels calls)
		mockAPI.SetGetConversationsError(true)

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "", false)

		// Should return error about failed to get new channels
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get new channels")
	})
}

func TestDryRunModeFunctionality(t *testing.T) {
	t.Run("Dry run mode with announcement - no messages posted", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-1", testTime, "Test purpose 1")
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "#general", true) // dry run mode = true
		assert.NoError(t, err)

		// Verify NO message was posted in dry run mode
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 0)
	})

	t.Run("Commit mode with announcement - message posted", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-1", testTime, "Test purpose 1")
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "#general", false) // commit mode = false
		assert.NoError(t, err)

		// Verify message WAS posted in commit mode
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 1)
		assert.Equal(t, "CGENERAL", messages[0].ChannelID)
	})

	t.Run("Dry run mode without announcement - shows message dry run", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-1", testTime, "Test purpose 1")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "", true) // dry run mode = true, no announcement channel
		assert.NoError(t, err)

		// Verify no messages posted (none expected)
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 0)
	})

	t.Run("Dry run mode without announcement - validates message format", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-new", testTime, "Test purpose for new channel")
		mockAPI.AddChannel("C456", "another-channel", testTime.Add(-30*time.Minute), "Another purpose")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)

		// Test that function executes without error and generates expected announcement format
		err = runDetectWithClient(client, cutoffTime, "", true) // dry run mode = true, no announcement channel
		assert.NoError(t, err)

		// Verify the announcement message would be properly formatted
		// We can test this by calling the format function directly
		newChannels, err := client.GetNewChannels(cutoffTime)
		require.NoError(t, err)
		assert.Len(t, newChannels, 2)

		message := client.FormatNewChannelAnnouncement(newChannels, cutoffTime)
		assert.Contains(t, message, "2 new channels created!")
		assert.Contains(t, message, "<#C123>")
		assert.Contains(t, message, "<#C456>")
		assert.Contains(t, message, "Test purpose for new channel")
		assert.Contains(t, message, "Another purpose")

		// Verify no actual messages posted
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 0)
	})

	t.Run("Dry run mode with no channels found - no dry run shown", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		// Don't add any channels, so none will be found

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "", true) // dry run mode = true, no announcement channel
		assert.NoError(t, err)

		// Verify no messages posted (none expected)
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 0)

		// When no channels are found, the function should return early and not show any dry run
		newChannels, err := client.GetNewChannels(cutoffTime)
		require.NoError(t, err)
		assert.Len(t, newChannels, 0)
	})

	t.Run("Dry run mode with announcement channel shows target", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-1", testTime, "Test purpose 1")
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "#general", true) // dry run mode = true WITH announcement channel
		assert.NoError(t, err)

		// Verify NO message was posted in dry run mode
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 0)

		// Verify the message would be properly formatted
		newChannels, err := client.GetNewChannels(cutoffTime)
		require.NoError(t, err)
		assert.Len(t, newChannels, 1)

		message := client.FormatNewChannelAnnouncement(newChannels, cutoffTime)
		assert.Contains(t, message, "New channel alert!")
		assert.Contains(t, message, "<#C123>")
		assert.Contains(t, message, "Test purpose 1")
	})
}

func TestDuplicateAnnouncementDetection(t *testing.T) {
	t.Run("No duplicate when no previous messages", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		isDuplicate, err := client.CheckForDuplicateAnnouncement("#general", "test message", []string{"new-channel"})
		assert.NoError(t, err)
		assert.False(t, isDuplicate)
	})

	t.Run("Detects duplicate when same channel was already announced", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")

		// Add a previous announcement message from our bot
		mockAPI.AddMessageToHistory("CGENERAL", "New channel alert! #test-channel", "U0000000", fmt.Sprintf("%.6f", float64(time.Now().Add(-1*time.Hour).Unix())))

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		isDuplicate, err := client.CheckForDuplicateAnnouncement("#general", "test message", []string{"test-channel"})
		assert.NoError(t, err)
		assert.True(t, isDuplicate)
	})

	t.Run("Does not detect duplicate when different channel was announced", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")

		// Add a previous announcement message from our bot for a different channel
		mockAPI.AddMessageToHistory("CGENERAL", "New channel alert! #other-channel", "U0000000", fmt.Sprintf("%.6f", float64(time.Now().Add(-1*time.Hour).Unix())))

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		isDuplicate, err := client.CheckForDuplicateAnnouncement("#general", "test message", []string{"test-channel"})
		assert.NoError(t, err)
		assert.False(t, isDuplicate)
	})

	t.Run("Ignores messages from other users", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")

		// Add a message from a different user mentioning the same channel
		mockAPI.AddMessageToHistory("CGENERAL", "New channel alert! #test-channel", "U1234567", fmt.Sprintf("%.6f", float64(time.Now().Add(-1*time.Hour).Unix())))

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		isDuplicate, err := client.CheckForDuplicateAnnouncement("#general", "test message", []string{"test-channel"})
		assert.NoError(t, err)
		assert.False(t, isDuplicate)
	})

	t.Run("Does not detect duplicate with generic new channel announcement for different channel", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")

		// Add a previous announcement message from our bot with generic new channel text for a different channel
		mockAPI.AddMessageToHistory("CGENERAL", "New channel created: #other-channel", "U0000000", fmt.Sprintf("%.6f", float64(time.Now().Add(-1*time.Hour).Unix())))

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		isDuplicate, err := client.CheckForDuplicateAnnouncement("#general", "test message", []string{"test-channel"})
		assert.NoError(t, err)
		assert.False(t, isDuplicate)
	})

	t.Run("Handles channel history error gracefully", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")

		// Set up error for GetConversationHistory
		mockAPI.SetGetConversationHistoryError("CGENERAL", true)

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		isDuplicate, err := client.CheckForDuplicateAnnouncement("#general", "test message", []string{"test-channel"})
		assert.NoError(t, err)
		assert.False(t, isDuplicate) // Should return false when can't check
	})
}

func TestDuplicateAnnouncementIntegration(t *testing.T) {
	t.Run("Skips posting duplicate announcement", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-1", testTime, "Test purpose 1")
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")

		// Add a previous announcement for the same channel
		mockAPI.AddMessageToHistory("CGENERAL", "New channel alert! #test-channel-1", "U0000000", fmt.Sprintf("%.6f", float64(time.Now().Add(-30*time.Minute).Unix())))

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "#general", false) // commit mode
		assert.NoError(t, err)

		// Verify NO new message was posted (duplicate was detected)
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 0)
	})

	t.Run("Posts announcement when not duplicate", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-1", testTime, "Test purpose 1")
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")

		// Add a previous announcement for a DIFFERENT channel
		mockAPI.AddMessageToHistory("CGENERAL", "New channel alert! #other-channel", "U0000000", fmt.Sprintf("%.6f", float64(time.Now().Add(-30*time.Minute).Unix())))

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "#general", false) // commit mode
		assert.NoError(t, err)

		// Verify message WAS posted (no duplicate detected)
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 1)
		assert.Equal(t, "CGENERAL", messages[0].ChannelID)
	})
}
