package cmd

import (
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
		assert.Equal(t, "24h", sinceFlag.DefValue)

		announceFlag := detectCmd.Flags().Lookup("announce-to")
		assert.NotNil(t, announceFlag)
		assert.Equal(t, "", announceFlag.DefValue)
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
		t.Setenv("SLACK_TOKEN", "xoxb-TEST-TOKEN-MOCK-FOR-TESTING-ONLY-NOT-REAL")
		
		// Force viper to reload environment
		initConfig()
		
		since = "invalid"
		
		cmd := &cobra.Command{}
		err := runDetect(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid time format")
		
		// Reset values
		since = originalSince
	})
}

func TestTimeParsing(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		valid    bool
	}{
		{"Valid 24h", "24h", true},
		{"Valid 168h", "168h", true},
		{"Valid 30m", "30m", true},
		{"Valid 1h30m", "1h30m", true},
		{"Invalid format", "invalid", false},
		{"Invalid 7d", "7d", false},
		{"Invalid 1w", "1w", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := time.ParseDuration(tc.input)
			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
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
		t.Setenv("SLACK_TOKEN", "xoxb-TEST-TOKEN-MOCK-FOR-TESTING-ONLY-NOT-REAL")
		initConfig()
		
		since = "1h"
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
		t.Setenv("SLACK_TOKEN", "xoxb-TEST-TOKEN-MOCK-FOR-TESTING-ONLY-NOT-REAL")
		initConfig()
		
		since = "2h"
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
		err = runDetectWithClient(client, cutoffTime, "")
		assert.NoError(t, err)
	})

	t.Run("Channels found", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-1", testTime, "Test purpose 1")
		
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "")
		assert.NoError(t, err)
	})

	t.Run("Channels found with announcement", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-1", testTime, "Test purpose 1")
		
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "#general")
		assert.NoError(t, err)

		// Verify message was posted
		messages := mockAPI.GetPostedMessages()
		assert.Len(t, messages, 1)
		assert.Equal(t, "general", messages[0].ChannelID)
	})

	t.Run("Announcement posting error", func(t *testing.T) {
		mockAPI := slack.NewMockSlackAPI()
		testTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C123", "test-channel-1", testTime, "Test purpose 1")
		
		// Set up mock to fail on PostMessage
		mockAPI.SetPostMessageError("channel_not_found")
		
		client, err := slack.NewClientWithAPI(mockAPI)
		require.NoError(t, err)

		cutoffTime := time.Now().Add(-2 * time.Hour)
		err = runDetectWithClient(client, cutoffTime, "#nonexistent")
		
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
		err = runDetectWithClient(client, cutoffTime, "")
		
		// Should return error about failed to get new channels
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get new channels")
	})
}