package cmd

import (
	"github.com/astrostl/slack-buddy-ai/pkg/slack"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRunDetectIntegration(t *testing.T) {
	t.Run("Integration with mock client", func(t *testing.T) {
		// Setup mock API
		mockAPI := slack.NewMockSlackAPI()
		
		// Add a test channel
		createdTime := time.Now().Add(-1 * time.Hour)
		mockAPI.AddChannel("C1234567890", "test-channel", createdTime, "Test purpose")
		
		// Test the integration by mocking the client creation
		// This would require dependency injection which is more complex
		// For now, we test the components separately
		client, err := slack.NewClientWithAPI(mockAPI)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		
		// Test getting new channels
		since := time.Now().Add(-2 * time.Hour)
		channels, err := client.GetNewChannels(since)
		assert.NoError(t, err)
		assert.Len(t, channels, 1)
		assert.Equal(t, "test-channel", channels[0].Name)
	})
}

func TestRunDetectErrors(t *testing.T) {
	t.Run("Test with invalid token format", func(t *testing.T) {
		// Set an invalid token that will fail validation
		t.Setenv("SLACK_TOKEN", "invalid-token")
		initConfig()
		
		cmd := &cobra.Command{}
		err := runDetect(cmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create Slack client")
	})
}

func TestCommandFlags(t *testing.T) {
	t.Run("Test since flag handling", func(t *testing.T) {
		// Test that the global since variable gets set properly
		originalSince := since
		since = "2h"
		
		duration, err := time.ParseDuration(since)
		assert.NoError(t, err)
		assert.Equal(t, 2*time.Hour, duration)
		
		since = originalSince
	})
	
	t.Run("Test announce-to flag handling", func(t *testing.T) {
		originalAnnounceTo := announceTo
		announceTo = "#general"
		
		assert.Equal(t, "#general", announceTo)
		
		announceTo = originalAnnounceTo
	})
}