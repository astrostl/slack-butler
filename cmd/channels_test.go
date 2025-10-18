package cmd

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
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

	t.Run("Announce-to required when commit is true", func(t *testing.T) {
		// Save original values
		originalCommit := commit
		originalAnnounceTo := announceTo

		// Set valid token but no announce-to with commit=true
		t.Setenv("SLACK_TOKEN", "MOCK-BOT-TOKEN-FOR-TESTING-ONLY-NOT-REAL-TOKEN-AT-ALL")

		// Force viper to reload environment
		initConfig()

		commit = true
		announceTo = ""

		cmd := &cobra.Command{}
		err := runDetect(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--announce-to is required when using --commit")

		// Reset values
		commit = originalCommit
		announceTo = originalAnnounceTo
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
		assert.Contains(t, message, "2 new channels created in the last")
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
		assert.Contains(t, message, "New channel created in the last")
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
		mockAPI.AddMessageToHistory("CGENERAL", "New channel created in the last 1 day! #test-channel", "U0000000", fmt.Sprintf("%.6f", float64(time.Now().Add(-1*time.Hour).Unix())))

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
		mockAPI.AddMessageToHistory("CGENERAL", "New channel created in the last 1 day! #other-channel", "U0000000", fmt.Sprintf("%.6f", float64(time.Now().Add(-1*time.Hour).Unix())))

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
		mockAPI.AddMessageToHistory("CGENERAL", "New channel created in the last 1 day! #test-channel", "U1234567", fmt.Sprintf("%.6f", float64(time.Now().Add(-1*time.Hour).Unix())))

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
		mockAPI.AddMessageToHistory("CGENERAL", "New channel created in the last 1 day! #test-channel-1", "U0000000", fmt.Sprintf("%.6f", float64(time.Now().Add(-30*time.Minute).Unix())))

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
		mockAPI.AddMessageToHistory("CGENERAL", "New channel created in the last 1 day! #other-channel", "U0000000", fmt.Sprintf("%.6f", float64(time.Now().Add(-30*time.Minute).Unix())))

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

func TestFilterChannelsByNames(t *testing.T) {
	channels := []slack.Channel{
		{Name: "general"},
		{Name: "random"},
		{Name: "development"},
		{Name: "marketing"},
	}

	t.Run("Filter matching channels", func(t *testing.T) {
		names := []string{"general", "development"}
		result := filterChannelsByNames(channels, names)
		assert.Len(t, result, 2)
		assert.Equal(t, "general", result[0].Name)
		assert.Equal(t, "development", result[1].Name)
	})

	t.Run("Filter with no matches", func(t *testing.T) {
		names := []string{"nonexistent"}
		result := filterChannelsByNames(channels, names)
		assert.Empty(t, result)
	})

	t.Run("Filter with empty names", func(t *testing.T) {
		names := []string{}
		result := filterChannelsByNames(channels, names)
		assert.Empty(t, result)
	})
}

func TestDisplayAnnouncingChannels(t *testing.T) {
	// Capture stdout for testing
	oldStdout := os.Stdout

	tests := []struct {
		name         string
		expectedOut  string
		channelNames []string
		skippedCount int
	}{
		{
			name:         "Single channel, no skipped",
			channelNames: []string{"test-channel"},
			skippedCount: 0,
			expectedOut:  "Announcing channels: #test-channel (skipped 0 already announced)",
		},
		{
			name:         "Multiple channels, some skipped",
			channelNames: []string{"channel1", "channel2", "channel3"},
			skippedCount: 2,
			expectedOut:  "Announcing channels: #channel1, #channel2, #channel3 (skipped 2 already announced)",
		},
		{
			name:         "No channels, some skipped",
			channelNames: []string{},
			skippedCount: 5,
			expectedOut:  "Announcing channels:  (skipped 5 already announced)",
		},
		{
			name:         "Single channel with special characters",
			channelNames: []string{"test-channel_2023"},
			skippedCount: 1,
			expectedOut:  "Announcing channels: #test-channel_2023 (skipped 1 already announced)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new pipe for this test
			r, w, err := os.Pipe()
			require.NoError(t, err)
			os.Stdout = w

			// Call the function
			displayAnnouncingChannels(tt.channelNames, tt.skippedCount)

			// Close writer and restore stdout
			err = w.Close()
			require.NoError(t, err)
			os.Stdout = oldStdout

			// Read the captured output
			output, err := io.ReadAll(r)
			require.NoError(t, err)
			outputStr := strings.TrimSpace(string(output))

			assert.Equal(t, tt.expectedOut, outputStr)
		})
	}

	// Restore stdout at the end
	os.Stdout = oldStdout
}

func TestParseExclusionLists(t *testing.T) {
	t.Run("Parse channels with # prefix", func(t *testing.T) {
		channels, prefixes := parseExclusionLists("#general,#random", "")
		assert.Equal(t, []string{"general", "random"}, channels)
		assert.Empty(t, prefixes)
	})

	t.Run("Parse channels without # prefix", func(t *testing.T) {
		channels, prefixes := parseExclusionLists("general,random", "")
		assert.Equal(t, []string{"general", "random"}, channels)
		assert.Empty(t, prefixes)
	})

	t.Run("Parse prefixes with # prefix", func(t *testing.T) {
		channels, prefixes := parseExclusionLists("", "#test-,#dev-")
		assert.Empty(t, channels)
		assert.Equal(t, []string{"test-", "dev-"}, prefixes)
	})

	t.Run("Parse prefixes without # prefix", func(t *testing.T) {
		channels, prefixes := parseExclusionLists("", "test-,dev-")
		assert.Empty(t, channels)
		assert.Equal(t, []string{"test-", "dev-"}, prefixes)
	})

	t.Run("Parse mixed channels and prefixes", func(t *testing.T) {
		channels, prefixes := parseExclusionLists("#general, random , announcements", "test-, #dev-, bot-")
		assert.Equal(t, []string{"general", "random", "announcements"}, channels)
		assert.Equal(t, []string{"test-", "dev-", "bot-"}, prefixes)
	})

	t.Run("Parse empty strings", func(t *testing.T) {
		channels, prefixes := parseExclusionLists("", "")
		assert.Empty(t, channels)
		assert.Empty(t, prefixes)
	})

	t.Run("Parse with empty elements", func(t *testing.T) {
		channels, prefixes := parseExclusionLists("general,,random", "test-,,dev-")
		assert.Equal(t, []string{"general", "random"}, channels)
		assert.Equal(t, []string{"test-", "dev-"}, prefixes)
	})

	t.Run("Parse with only # characters", func(t *testing.T) {
		channels, prefixes := parseExclusionLists("#,##", "#,##")
		// Note: TrimPrefix only removes one #, so ## becomes #, which is not empty
		assert.Equal(t, []string{"#"}, channels)
		assert.Equal(t, []string{"#"}, prefixes)
	})
}

func TestDisplayExclusionInfo(t *testing.T) {
	oldStdout := os.Stdout

	t.Run("Display channels and prefixes", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		displayExclusionInfo([]string{"general", "random"}, []string{"test-", "dev-"}, []string{})

		err = w.Close()
		require.NoError(t, err)
		os.Stdout = oldStdout

		output, err := io.ReadAll(r)
		require.NoError(t, err)
		outputStr := string(output)

		assert.Contains(t, outputStr, "📋 Channel exclusions configured:")
		assert.Contains(t, outputStr, "Manually excluded channels: general, random")
		assert.Contains(t, outputStr, "Excluded prefixes: test-, dev-")
	})

	t.Run("Display only channels", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		displayExclusionInfo([]string{"general"}, []string{}, []string{})

		err = w.Close()
		require.NoError(t, err)
		os.Stdout = oldStdout

		output, err := io.ReadAll(r)
		require.NoError(t, err)
		outputStr := string(output)

		assert.Contains(t, outputStr, "📋 Channel exclusions configured:")
		assert.Contains(t, outputStr, "Manually excluded channels: general")
		assert.NotContains(t, outputStr, "Excluded prefixes:")
	})

	t.Run("Display only prefixes", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		displayExclusionInfo([]string{}, []string{"test-"}, []string{})

		err = w.Close()
		require.NoError(t, err)
		os.Stdout = oldStdout

		output, err := io.ReadAll(r)
		require.NoError(t, err)
		outputStr := string(output)

		assert.Contains(t, outputStr, "📋 Channel exclusions configured:")
		assert.NotContains(t, outputStr, "Excluded channels:")
		assert.Contains(t, outputStr, "Excluded prefixes: test-")
	})

	t.Run("Display nothing when no exclusions", func(t *testing.T) {
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		displayExclusionInfo([]string{}, []string{}, []string{})

		err = w.Close()
		require.NoError(t, err)
		os.Stdout = oldStdout

		output, err := io.ReadAll(r)
		require.NoError(t, err)
		outputStr := string(output)

		assert.Empty(t, strings.TrimSpace(outputStr))
	})

	os.Stdout = oldStdout
}

func TestGetUserMapWithErrorHandling(t *testing.T) {
	t.Run("Success with debug mode", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.AddUser("U1234567", "testuser", "Test User")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		userMap, err := getUserMapWithErrorHandling(client, true)
		assert.NoError(t, err)
		assert.Len(t, userMap, 1)
		assert.Equal(t, "Test User", userMap["U1234567"])
	})

	t.Run("Success without debug mode", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.AddUser("U1234567", "testuser", "Test User")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		userMap, err := getUserMapWithErrorHandling(client, false)
		assert.NoError(t, err)
		assert.Len(t, userMap, 1)
		assert.Equal(t, "Test User", userMap["U1234567"])
	})

	t.Run("Rate limit error", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetGetUsersError("rate_limited")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		userMap, err := getUserMapWithErrorHandling(client, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limited by Slack API")
		assert.Nil(t, userMap)
	})

	t.Run("Missing scope error", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetGetUsersError("missing_scope: users:read")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		userMap, err := getUserMapWithErrorHandling(client, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required OAuth scope 'users:read'")
		assert.Nil(t, userMap)
	})

	t.Run("Generic error", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetGetUsersError("generic API error")

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		userMap, err := getUserMapWithErrorHandling(client, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get users")
		assert.Nil(t, userMap)
	})
}

func TestGetInactiveChannelsWithErrorHandling(t *testing.T) {
	t.Run("Success with channels", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()

		// Add an inactive channel
		mockAPI.AddChannel("C1", "inactive-channel", time.Now().Add(-2*time.Hour), "Inactive")
		lastActivity := time.Now().Add(-35 * time.Second)
		mockAPI.AddMessageToHistory("C1", "old message", "U1234567", fmt.Sprintf("%.6f", float64(lastActivity.Unix())))

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		userMap := map[string]string{"U1234567": "testuser"}
		toWarn, toArchive, _, err := getInactiveChannelsWithErrorHandling(client, 30, 7, userMap, []string{}, []string{}, false)

		assert.NoError(t, err)
		assert.Len(t, toWarn, 1)
		assert.Len(t, toArchive, 0)
		assert.Equal(t, "inactive-channel", toWarn[0].Name)
	})

	t.Run("API error handling", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetGetConversationsError(true)

		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		userMap := map[string]string{}
		toWarn, toArchive, _, err := getInactiveChannelsWithErrorHandling(client, 30, 7, userMap, []string{}, []string{}, false)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to analyze inactive channels")
		assert.Nil(t, toWarn)
		assert.Nil(t, toArchive)
	})
}

func TestRunDetectErrorPaths(t *testing.T) {
	t.Run("Days parsing error path", func(t *testing.T) {
		originalSince := since

		t.Setenv("SLACK_TOKEN", "xoxb-valid-token-format")
		initConfig()

		since = "invalid-format"

		cmd := &cobra.Command{}
		err := runDetect(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid days format")

		since = originalSince
	})

	t.Run("Negative days error", func(t *testing.T) {
		originalSince := since

		t.Setenv("SLACK_TOKEN", "xoxb-valid-token-format")
		initConfig()

		since = "-1"

		cmd := &cobra.Command{}
		err := runDetect(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "days must be positive")

		since = originalSince
	})
}

func TestHighlightCommandSetup(t *testing.T) {
	t.Run("Command structure", func(t *testing.T) {
		assert.NotNil(t, highlightCmd)
		assert.Equal(t, "highlight", highlightCmd.Use)
		assert.NotEmpty(t, highlightCmd.Short)
		assert.Contains(t, highlightCmd.Long, "Randomly select and highlight channels")
	})

	t.Run("Command flags", func(t *testing.T) {
		countFlag := highlightCmd.Flags().Lookup("count")
		assert.NotNil(t, countFlag)
		assert.Equal(t, "3", countFlag.DefValue)

		announceFlag := highlightCmd.Flags().Lookup("announce-to")
		assert.NotNil(t, announceFlag)
		assert.Equal(t, "", announceFlag.DefValue)

		commitFlag := highlightCmd.Flags().Lookup("commit")
		assert.NotNil(t, commitFlag)
		assert.Equal(t, "false", commitFlag.DefValue)
	})

	t.Run("Command hierarchy", func(t *testing.T) {
		// Check that highlight is a subcommand of channels
		subcommands := channelsCmd.Commands()
		var highlightFound bool
		for _, cmd := range subcommands {
			if cmd.Use == "highlight" {
				highlightFound = true
				break
			}
		}
		assert.True(t, highlightFound)
	})
}

func TestRunHighlightFunction(t *testing.T) {
	t.Run("Missing token error", func(t *testing.T) {
		// Clear any environment token
		t.Setenv("SLACK_TOKEN", "")

		cmd := &cobra.Command{}
		err := runHighlight(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "slack token is required")
	})

	t.Run("Announce-to required when commit is true", func(t *testing.T) {
		// Save original values
		originalCommit := commit
		originalAnnounceTo := announceTo

		// Set valid token but no announce-to with commit=true
		t.Setenv("SLACK_TOKEN", "MOCK-BOT-TOKEN-FOR-TESTING-ONLY-NOT-REAL-TOKEN-AT-ALL")

		// Force viper to reload environment
		initConfig()

		commit = true
		announceTo = ""

		cmd := &cobra.Command{}
		err := runHighlight(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--announce-to is required when using --commit")

		// Reset values
		commit = originalCommit
		announceTo = originalAnnounceTo
	})

	t.Run("Invalid count error", func(t *testing.T) {
		// Save original values
		originalCount := count

		// Set valid token but invalid count
		t.Setenv("SLACK_TOKEN", "MOCK-BOT-TOKEN-FOR-TESTING-ONLY-NOT-REAL-TOKEN-AT-ALL")

		// Force viper to reload environment
		initConfig()

		count = 0

		cmd := &cobra.Command{}
		err := runHighlight(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "count must be positive")

		// Reset values
		count = originalCount
	})

	t.Run("Negative count error", func(t *testing.T) {
		originalCount := count

		t.Setenv("SLACK_TOKEN", "xoxb-valid-token-format")
		initConfig()

		count = -5

		cmd := &cobra.Command{}
		err := runHighlight(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "count must be positive")

		count = originalCount
	})
}

func TestRunHighlightWithClient(t *testing.T) {
	t.Run("Successful highlight without announcement", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		// Add some test channels for highlighting
		mockAPI.AddChannel("C1234567", "test-channel-1", time.Now().Add(-24*time.Hour), "Test purpose 1")
		mockAPI.AddChannel("C2345678", "test-channel-2", time.Now().Add(-48*time.Hour), "Test purpose 2")

		// Capture stdout
		r, w, _ := os.Pipe() //nolint:errcheck
		oldStdout := os.Stdout
		os.Stdout = w

		err = runHighlightWithClient(client, 2, "", true)

		// Restore stdout
		_ = w.Close() //nolint:errcheck
		os.Stdout = oldStdout

		// Read captured output
		output, _ := io.ReadAll(r) //nolint:errcheck
		outputStr := string(output)

		assert.NoError(t, err)
		assert.Contains(t, outputStr, "Random channels to highlight (2):")
		assert.Contains(t, outputStr, "#test-channel-")
	})

	t.Run("No channels found", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		// Capture stdout
		r, w, _ := os.Pipe() //nolint:errcheck
		oldStdout := os.Stdout
		os.Stdout = w

		err = runHighlightWithClient(client, 5, "", true)

		// Restore stdout
		_ = w.Close() //nolint:errcheck
		os.Stdout = oldStdout

		// Read captured output
		output, _ := io.ReadAll(r) //nolint:errcheck
		outputStr := string(output)

		assert.NoError(t, err)
		assert.Contains(t, outputStr, "No channels found to highlight")
	})

	t.Run("Successful highlight with announcement - dry run", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		// Add test channel
		mockAPI.AddChannel("C1234567", "test-channel-1", time.Now().Add(-24*time.Hour), "Test purpose 1")

		// Add announcement channel
		mockAPI.AddChannel("C9999999", "general", time.Now().Add(-48*time.Hour), "General channel")

		// Capture stdout
		r, w, _ := os.Pipe() //nolint:errcheck
		oldStdout := os.Stdout
		os.Stdout = w

		err = runHighlightWithClient(client, 1, "#general", true)

		// Restore stdout
		_ = w.Close() //nolint:errcheck
		os.Stdout = oldStdout

		// Read captured output
		output, _ := io.ReadAll(r) //nolint:errcheck
		outputStr := string(output)

		assert.NoError(t, err)
		assert.Contains(t, outputStr, "--- DRY RUN ---")
		assert.Contains(t, outputStr, "Would announce to channel: #general")
		assert.Contains(t, outputStr, "--- END DRY RUN ---")

		// Verify no message was actually posted
		messages := mockAPI.GetPostedMessages()
		assert.Empty(t, messages)
	})

	t.Run("Successful highlight with announcement - commit mode", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		// Add test channel
		mockAPI.AddChannel("C1234567", "test-channel-1", time.Now().Add(-24*time.Hour), "Test purpose 1")

		// Add announcement channel
		mockAPI.AddChannel("C9999999", "general", time.Now().Add(-48*time.Hour), "General channel")

		// Capture stdout
		r, w, _ := os.Pipe() //nolint:errcheck
		oldStdout := os.Stdout
		os.Stdout = w

		err = runHighlightWithClient(client, 1, "#general", false)

		// Restore stdout
		_ = w.Close() //nolint:errcheck
		os.Stdout = oldStdout

		// Read captured output
		output, _ := io.ReadAll(r) //nolint:errcheck
		outputStr := string(output)

		assert.NoError(t, err)
		assert.Contains(t, outputStr, "Channel highlight posted to #general")

		// Verify message was posted
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 1)
		assert.Equal(t, "C9999999", messages[0].ChannelID)
		assert.Contains(t, messages[0].Text, "mock-message-posted")
	})

	t.Run("GetRandomChannels error", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetGetConversationsErrorWithMessage(true, "API error")
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		err = runHighlightWithClient(client, 1, "", true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get random channels")
	})
}

func TestHandleHighlightAnnouncement(t *testing.T) {
	t.Run("Successful announcement - dry run", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		channels := []slack.Channel{
			{
				ID:   "C1234567",
				Name: "test-channel-1",
			},
		}

		// Capture stdout
		r, w, _ := os.Pipe() //nolint:errcheck
		oldStdout := os.Stdout
		os.Stdout = w

		err = handleHighlightAnnouncement(client, channels, "#general", true)

		// Restore stdout
		_ = w.Close() //nolint:errcheck
		os.Stdout = oldStdout

		// Read captured output
		output, _ := io.ReadAll(r) //nolint:errcheck
		outputStr := string(output)

		assert.NoError(t, err)
		assert.Contains(t, outputStr, "--- DRY RUN ---")
		assert.Contains(t, outputStr, "Would announce to channel: #general")
		assert.Contains(t, outputStr, "To actually post this highlight")

		// Verify no message was posted
		messages := mockAPI.GetPostedMessages()
		assert.Empty(t, messages)
	})

	t.Run("Successful announcement - commit mode", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		// Add announcement channel
		mockAPI.AddChannel("C9999999", "general", time.Now().Add(-48*time.Hour), "General channel")

		channels := []slack.Channel{
			{
				ID:   "C1234567",
				Name: "test-channel-1",
			},
		}

		// Capture stdout
		r, w, _ := os.Pipe() //nolint:errcheck
		oldStdout := os.Stdout
		os.Stdout = w

		err = handleHighlightAnnouncement(client, channels, "#general", false)

		// Restore stdout
		_ = w.Close() //nolint:errcheck
		os.Stdout = oldStdout

		// Read captured output
		output, _ := io.ReadAll(r) //nolint:errcheck
		outputStr := string(output)

		assert.NoError(t, err)
		assert.Contains(t, outputStr, "Channel highlight posted to #general")

		// Verify message was posted
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 1)
		assert.Equal(t, "C9999999", messages[0].ChannelID)
	})

	t.Run("PostMessage error", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		mockAPI.SetPostMessageError("channel_not_found")
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		channels := []slack.Channel{
			{
				ID:   "C1234567",
				Name: "test-channel-1",
			},
		}

		err = handleHighlightAnnouncement(client, channels, "#nonexistent", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to post highlight")
	})
}

func TestHandleHighlightDryRunWithoutChannel(t *testing.T) {
	t.Run("Dry run without channel", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		channels := []slack.Channel{
			{
				ID:   "C1234567",
				Name: "test-channel-1",
			},
		}

		// Capture stdout
		r, w, _ := os.Pipe() //nolint:errcheck
		oldStdout := os.Stdout
		os.Stdout = w

		err = handleHighlightDryRunWithoutChannel(client, channels, true)

		// Restore stdout
		_ = w.Close() //nolint:errcheck
		os.Stdout = oldStdout

		// Read captured output
		output, _ := io.ReadAll(r) //nolint:errcheck
		outputStr := string(output)

		assert.NoError(t, err)
		assert.Contains(t, outputStr, "--- DRY RUN ---")
		assert.Contains(t, outputStr, "Channel highlight message dry run")
		assert.Contains(t, outputStr, "use --announce-to to specify target")
		assert.Contains(t, outputStr, "--- END DRY RUN ---")
	})

	t.Run("Non-dry run mode does nothing", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		channels := []slack.Channel{
			{
				ID:   "C1234567",
				Name: "test-channel-1",
			},
		}

		// Capture stdout
		r, w, _ := os.Pipe() //nolint:errcheck
		oldStdout := os.Stdout
		os.Stdout = w

		err = handleHighlightDryRunWithoutChannel(client, channels, false)

		// Restore stdout
		_ = w.Close() //nolint:errcheck
		os.Stdout = oldStdout

		// Read captured output
		output, _ := io.ReadAll(r) //nolint:errcheck
		outputStr := string(output)

		assert.NoError(t, err)
		assert.Empty(t, outputStr)
	})
}

func TestMergeChannelLists(t *testing.T) {
	t.Run("Merge two lists without duplicates", func(t *testing.T) {
		list1 := []string{"channel1", "channel2"}
		list2 := []string{"channel3", "channel4"}
		result := mergeChannelLists(list1, list2)
		assert.Len(t, result, 4)
		assert.Contains(t, result, "channel1")
		assert.Contains(t, result, "channel2")
		assert.Contains(t, result, "channel3")
		assert.Contains(t, result, "channel4")
	})

	t.Run("Merge lists with duplicates", func(t *testing.T) {
		list1 := []string{"channel1", "channel2"}
		list2 := []string{"channel2", "channel3"}
		result := mergeChannelLists(list1, list2)
		assert.Len(t, result, 3)
		assert.Contains(t, result, "channel1")
		assert.Contains(t, result, "channel2")
		assert.Contains(t, result, "channel3")
	})

	t.Run("Merge empty lists", func(t *testing.T) {
		result := mergeChannelLists([]string{}, []string{})
		assert.Len(t, result, 0)
	})

	t.Run("Merge with one empty list", func(t *testing.T) {
		list1 := []string{"channel1"}
		result := mergeChannelLists(list1, []string{})
		assert.Len(t, result, 1)
		assert.Contains(t, result, "channel1")
	})
}

func TestSeparateManualExclusions(t *testing.T) {
	t.Run("Separate manual from defaults", func(t *testing.T) {
		excludeList := []string{"manual1", "default1", "manual2"}
		defaultList := []string{"default1", "default2"}
		result := separateManualExclusions(excludeList, defaultList)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "manual1")
		assert.Contains(t, result, "manual2")
		assert.NotContains(t, result, "default1")
	})

	t.Run("All are manual exclusions", func(t *testing.T) {
		excludeList := []string{"manual1", "manual2"}
		defaultList := []string{"default1"}
		result := separateManualExclusions(excludeList, defaultList)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "manual1")
		assert.Contains(t, result, "manual2")
	})

	t.Run("All are default channels", func(t *testing.T) {
		excludeList := []string{"default1", "default2"}
		defaultList := []string{"default1", "default2"}
		result := separateManualExclusions(excludeList, defaultList)
		assert.Len(t, result, 0)
	})
}

func TestAddHashPrefix(t *testing.T) {
	t.Run("Add prefix to channels", func(t *testing.T) {
		channels := []string{"general", "random", "tech"}
		result := addHashPrefix(channels)
		assert.Len(t, result, 3)
		assert.Equal(t, "#general", result[0])
		assert.Equal(t, "#random", result[1])
		assert.Equal(t, "#tech", result[2])
	})

	t.Run("Add prefix to empty list", func(t *testing.T) {
		channels := []string{}
		result := addHashPrefix(channels)
		assert.Len(t, result, 0)
	})

	t.Run("Add prefix to single channel", func(t *testing.T) {
		channels := []string{"general"}
		result := addHashPrefix(channels)
		assert.Len(t, result, 1)
		assert.Equal(t, "#general", result[0])
	})
}
