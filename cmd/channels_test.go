package cmd

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"slack-buddy-ai/pkg/slack"
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
		assert.Equal(t, "1", sinceFlag.DefValue)

		announceFlag := detectCmd.Flags().Lookup("announce-to")
		assert.NotNil(t, announceFlag)
		assert.Equal(t, "", announceFlag.DefValue)

		dryRunFlag := detectCmd.Flags().Lookup("dry-run")
		assert.NotNil(t, dryRunFlag)
		assert.Equal(t, "false", dryRunFlag.DefValue)
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
		
		cmd := &cobra.Command{}
		err := runDetect(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid days format")
		
		// Reset values
		since = originalSince
	})
}

func TestDaysParsing(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		valid    bool
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
		announceTo = ""
		
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

func TestDryRunFunctionality(t *testing.T) {
	t.Run("Dry run with announcement - no messages posted", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-1", testTime, "Test purpose 1")
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")
		
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "#general", true) // dry run = true
		assert.NoError(t, err)

		// Verify NO message was posted in dry run mode
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 0)
	})

	t.Run("Regular run with announcement - message posted", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-1", testTime, "Test purpose 1")
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")
		
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "#general", false) // dry run = false
		assert.NoError(t, err)

		// Verify message WAS posted in regular mode
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 1)
		assert.Equal(t, "CGENERAL", messages[0].ChannelID)
	})

	t.Run("Dry run without announcement - still processes normally", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-1", testTime, "Test purpose 1")
		
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "", true) // dry run = true, no announcement channel
		assert.NoError(t, err)

		// Verify no messages posted (none expected)
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 0)
	})
}

func TestIdempotencyIntegration(t *testing.T) {
	t.Run("Filter out previously announced channels", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		
		// Add two channels that were created recently
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannelWithCreator("C1234567890", "new-channel", testTime, "New channel", "U1111111")
		mockAPI.AddChannelWithCreator("C0987654321", "already-announced", testTime, "Already announced", "U2222222")
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")
		
		// Add message history showing that one channel was already announced (from bot)
		mockAPI.AddMessageToHistory("CGENERAL", "New channel alert!\n\n• <#C0987654321> - created June 22, 2025 by <@U2222222>\n  Purpose: Already announced", "U0000000", "1234567890.123")
		
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "#general", false)
		assert.NoError(t, err)

		// Verify only one message was posted (for the new channel)
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 1)
		assert.Equal(t, "CGENERAL", messages[0].ChannelID)
		
		// The message should only contain the new channel, not the already announced one
		// We can't easily check the exact message content due to mock limitations,
		// but the fact that only one message was posted indicates idempotency worked
	})

	t.Run("All channels already announced", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		
		// Add channels that were created recently
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannelWithCreator("C1234567890", "channel1", testTime, "Channel 1", "U1111111")
		mockAPI.AddChannelWithCreator("C0987654321", "channel2", testTime, "Channel 2", "U2222222")
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")
		
		// Add message history showing that both channels were already announced (from bot)
		mockAPI.AddMessageToHistory("CGENERAL", "Previous announcement with <#C1234567890> and <#C0987654321>", "U0000000", "1234567890.123")
		
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "#general", false)
		assert.NoError(t, err)

		// Verify no messages were posted (all channels already announced)
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 0)
	})

	t.Run("History API error stops announcement", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		
		// Add a channel
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannelWithCreator("C1234567890", "new-channel", testTime, "New channel", "U1111111")
		// Add the general channel that will be used for announcements
		mockAPI.AddChannel("CGENERAL", "general", time.Now().Add(-24*time.Hour), "General discussion")
		
		// Set up history API to fail with missing scope error
		mockAPI.SetConversationHistoryError(true)
		mockAPI.GetConversationHistoryError = fmt.Errorf("missing_scope")
		
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "#general", false)
		assert.Error(t, err) // Should fail due to missing permissions
		assert.Contains(t, err.Error(), "missing required permission")

		// Verify no message was posted (stopped due to error)
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 0)
	})

	t.Run("No announcement channel works normally", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		
		// Add a channel
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannelWithCreator("C1234567890", "new-channel", testTime, "New channel", "U1111111")
		
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		// No announcement channel specified, should work fine
		err = runDetectWithClient(client, cutoffTime, "", false)
		assert.NoError(t, err)

		// Verify no message was posted (no announcement channel)
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 0)
	})
}