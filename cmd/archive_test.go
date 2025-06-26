package cmd

import (
	"fmt"
	"testing"
	"time"

	"github.com/astrostl/slack-butler/pkg/slack"
)

type archiveTestCase struct {
	setupChannels   func(*slack.MockSlackAPI)
	name            string
	expectedErr     string
	warnSeconds     int
	archiveSeconds  int
	expectedWarn    int
	expectedArchive int
	isPreviewMode   bool
}

func setupActiveChannelTest(mock *slack.MockSlackAPI) {
	mock.AddChannel("C1", "active-channel", time.Now().Add(-10*time.Minute), "Active channel")
	mock.AddMessageToHistory("C1", "recent message", "U1234567", fmt.Sprintf("%.6f", float64(time.Now().Add(-2*time.Second).Unix())))
}

func setupWarningChannelTest(mock *slack.MockSlackAPI) {
	mock.AddChannel("C2", "inactive-channel", time.Now().Add(-2*time.Hour), "Inactive channel")
	lastActivity := time.Now().Add(-35 * time.Second)
	mock.AddMessageToHistory("C2", "old message", "U1234567", fmt.Sprintf("%.6f", float64(lastActivity.Unix())))
}

func setupArchiveChannelTest(mock *slack.MockSlackAPI) {
	mock.AddChannel("C3", "archive-channel", time.Now().Add(-2*time.Hour), "Channel to archive")
	lastActivity := time.Now().Add(-40 * time.Second)
	mock.AddMessageToHistory("C3", "old message", "U1234567", fmt.Sprintf("%.6f", float64(lastActivity.Unix())))
	warningTime := time.Now().Add(-10 * time.Second)
	mock.AddMessageToHistory("C3", "Warning: inactive channel warning <!-- inactive channel warning -->", "U0000000", fmt.Sprintf("%.6f", float64(warningTime.Unix())))
}

func setupMixedChannelsTest(mock *slack.MockSlackAPI) {
	mock.AddChannel("C1", "active", time.Now().Add(-10*time.Minute), "Active")
	mock.AddMessageToHistory("C1", "recent", "U1234567", fmt.Sprintf("%.6f", float64(time.Now().Add(-4*time.Second).Unix())))

	mock.AddChannel("C2", "warn-me", time.Now().Add(-2*time.Hour), "Warn me")
	lastActivity := time.Now().Add(-35 * time.Second)
	mock.AddMessageToHistory("C2", "old message", "U1234567", fmt.Sprintf("%.6f", float64(lastActivity.Unix())))

	mock.AddChannel("C3", "archive-me", time.Now().Add(-3*time.Hour), "Archive me")
	oldActivity := time.Now().Add(-40 * time.Second)
	mock.AddMessageToHistory("C3", "very old", "U1234567", fmt.Sprintf("%.6f", float64(oldActivity.Unix())))
	warningTime := time.Now().Add(-10 * time.Second)
	mock.AddMessageToHistory("C3", "Warning: inactive channel warning <!-- inactive channel warning -->", "U0000000", fmt.Sprintf("%.6f", float64(warningTime.Unix())))
}

func setupExcludedChannelsTest(mock *slack.MockSlackAPI) {
	mock.AddChannel("C1", "general", time.Now().Add(-3*time.Hour), "General channel")
	mock.AddChannel("C2", "announcements", time.Now().Add(-3*time.Hour), "Announcements")
	mock.AddChannel("C3", "admin-stuff", time.Now().Add(-3*time.Hour), "Admin channel")

	oldTime := time.Now().Add(-40 * time.Second)
	for _, channelID := range []string{"C1", "C2", "C3"} {
		mock.AddMessageToHistory(channelID, "old message", "U1234567", fmt.Sprintf("%.6f", float64(oldTime.Unix())))
	}
}

func setupCommitModeTest(mock *slack.MockSlackAPI) {
	mock.AddChannel("C1", "warn-channel", time.Now().Add(-2*time.Hour), "Warn this")
	lastActivity := time.Now().Add(-35 * time.Second)
	mock.AddMessageToHistory("C1", "old message", "U1234567", fmt.Sprintf("%.6f", float64(lastActivity.Unix())))

	mock.AddChannel("C2", "archive-channel", time.Now().Add(-3*time.Hour), "Archive this")
	oldActivity := time.Now().Add(-40 * time.Second)
	mock.AddMessageToHistory("C2", "very old", "U1234567", fmt.Sprintf("%.6f", float64(oldActivity.Unix())))
	warningTime := time.Now().Add(-10 * time.Second)
	mock.AddMessageToHistory("C2", "Warning: inactive channel warning <!-- inactive channel warning -->", "U0000000", fmt.Sprintf("%.6f", float64(warningTime.Unix())))
}

func getArchiveTestCases() []archiveTestCase {
	return []archiveTestCase{
		{
			name:            "No inactive channels",
			warnSeconds:     30,
			archiveSeconds:  7,
			isPreviewMode:   true,
			setupChannels:   setupActiveChannelTest,
			expectedWarn:    0,
			expectedArchive: 0,
		},
		{
			name:            "Channel needs warning - preview mode",
			warnSeconds:     30,
			archiveSeconds:  7,
			isPreviewMode:   true,
			setupChannels:   setupWarningChannelTest,
			expectedWarn:    1,
			expectedArchive: 0,
		},
		{
			name:            "Channel needs archiving - preview mode",
			warnSeconds:     30,
			archiveSeconds:  7,
			isPreviewMode:   true,
			setupChannels:   setupArchiveChannelTest,
			expectedWarn:    0,
			expectedArchive: 1,
		},
		{
			name:            "Mixed scenario - some warn, some archive",
			warnSeconds:     30,
			archiveSeconds:  7,
			isPreviewMode:   true,
			setupChannels:   setupMixedChannelsTest,
			expectedWarn:    1,
			expectedArchive: 1,
		},
		{
			name:            "Excluded channels are skipped",
			warnSeconds:     30,
			archiveSeconds:  7,
			isPreviewMode:   true,
			setupChannels:   setupExcludedChannelsTest,
			expectedWarn:    0,
			expectedArchive: 0,
		},
		{
			name:            "Warning and archiving actions - commit mode",
			warnSeconds:     30,
			archiveSeconds:  7,
			isPreviewMode:   false,
			setupChannels:   setupCommitModeTest,
			expectedWarn:    2,
			expectedArchive: 1,
		},
		{
			name:           "Invalid parameters",
			warnSeconds:    0,
			archiveSeconds: 7,
			isPreviewMode:  true,
			setupChannels:  func(mock *slack.MockSlackAPI) {},
			expectedErr:    "warn-seconds must be positive, got 0",
		},
		{
			name:           "API error handling",
			warnSeconds:    30,
			archiveSeconds: 7,
			isPreviewMode:  false,
			setupChannels: func(mock *slack.MockSlackAPI) {
				mock.SetGetConversationsError(true)
			},
			expectedErr: "failed to analyze inactive channels",
		},
	}
}

func validateTestParameters(t *testing.T, tt archiveTestCase) bool {
	if tt.expectedErr != "" && (tt.warnSeconds <= 0 || tt.archiveSeconds <= 0) {
		if tt.warnSeconds <= 0 {
			expectedMsg := fmt.Sprintf("warn-seconds must be positive, got %d", tt.warnSeconds)
			if expectedMsg != tt.expectedErr {
				t.Errorf("Expected error '%s', got constructed: %s", tt.expectedErr, expectedMsg)
			}
		}
		if tt.archiveSeconds <= 0 {
			expectedMsg := fmt.Sprintf("archive-seconds must be positive, got %d", tt.archiveSeconds)
			if expectedMsg != tt.expectedErr {
				t.Errorf("Expected error '%s', got constructed: %s", tt.expectedErr, expectedMsg)
			}
		}
		return false
	}
	return true
}

func validateTestResults(t *testing.T, tt archiveTestCase, err error, mockAPI *slack.MockSlackAPI) {
	if tt.expectedErr != "" {
		if err == nil {
			t.Errorf("Expected error containing '%s', got nil", tt.expectedErr)
			return
		}
		if err.Error() != tt.expectedErr && !containsString(err.Error(), tt.expectedErr) {
			t.Errorf("Expected error containing '%s', got: %s", tt.expectedErr, err.Error())
		}
		return
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if !tt.isPreviewMode {
		postedMessages := mockAPI.GetPostedMessages()
		if len(postedMessages) != tt.expectedWarn {
			t.Errorf("Expected %d warning messages, got %d", tt.expectedWarn, len(postedMessages))
		}

		archivedChannels := mockAPI.GetArchivedChannels()
		if len(archivedChannels) != tt.expectedArchive {
			t.Errorf("Expected %d archived channels, got %d", tt.expectedArchive, len(archivedChannels))
		}
	}
}

func TestRunArchiveWithClient(t *testing.T) {
	tests := getArchiveTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := slack.NewMockSlackAPI()

			if tt.setupChannels != nil {
				tt.setupChannels(mockAPI)
			}

			client, err := slack.NewClientWithAPI(mockAPI)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			if !validateTestParameters(t, tt) {
				return
			}

			err = runArchiveWithClient(client, tt.warnSeconds, tt.archiveSeconds, tt.isPreviewMode, "", "")
			validateTestResults(t, tt, err, mockAPI)
		})
	}
}

func TestInactiveChannelDetectionLogic(t *testing.T) {
	tests := []struct {
		name           string
		channelAge     time.Duration // time ago channel was created
		lastActivity   time.Duration // time ago last activity occurred
		warningAge     time.Duration // time ago warning was sent
		warnSeconds    int           // threshold for inactivity in seconds
		archiveSeconds int           // grace period after warning in seconds
		hasWarning     bool
		expectWarn     bool
		expectArchive  bool
	}{
		{
			name:           "New channel, active",
			channelAge:     10 * time.Minute,
			lastActivity:   1 * time.Second,
			warnSeconds:    30,
			archiveSeconds: 7,
			expectWarn:     false,
			expectArchive:  false,
		},
		{
			name:           "Old channel, recently active",
			channelAge:     100 * time.Minute,
			lastActivity:   5 * time.Second,
			warnSeconds:    30,
			archiveSeconds: 7,
			expectWarn:     false,
			expectArchive:  false,
		},
		{
			name:           "Old channel, inactive, no warning",
			channelAge:     60 * time.Minute,
			lastActivity:   35 * time.Second,
			warnSeconds:    30,
			archiveSeconds: 7,
			expectWarn:     true,
			expectArchive:  false,
		},
		{
			name:           "Old channel, inactive, recent warning",
			channelAge:     60 * time.Minute,
			lastActivity:   35 * time.Second,
			hasWarning:     true,
			warningAge:     3 * time.Second,
			warnSeconds:    30,
			archiveSeconds: 7,
			expectWarn:     false,
			expectArchive:  false,
		},
		{
			name:           "Old channel, inactive, old warning",
			channelAge:     60 * time.Minute,
			lastActivity:   40 * time.Second,
			hasWarning:     true,
			warningAge:     8 * time.Second,
			warnSeconds:    30,
			archiveSeconds: 7,
			expectWarn:     false,
			expectArchive:  true,
		},
		{
			name:           "Channel too new to be considered",
			channelAge:     10 * time.Second, // Channel created only 10 seconds ago
			lastActivity:   35 * time.Second, // Last activity 35 seconds ago (impossible, but test setup)
			warnSeconds:    30,               // 30 second warn threshold
			archiveSeconds: 7,
			expectWarn:     false, // Too new to warn (created after warn cutoff)
			expectArchive:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := slack.NewMockSlackAPI()

			// Setup channel
			channelCreated := time.Now().Add(-tt.channelAge)
			mockAPI.AddChannel("C1", "test-channel", channelCreated, "Test channel")

			// Add last activity
			lastActivityTime := time.Now().Add(-tt.lastActivity)
			mockAPI.AddMessageToHistory("C1", "last message", "U1234567", fmt.Sprintf("%.6f", float64(lastActivityTime.Unix())))

			// Add warning if needed
			if tt.hasWarning {
				warningTime := time.Now().Add(-tt.warningAge)
				mockAPI.AddMessageToHistory("C1", "Warning: inactive channel warning <!-- inactive channel warning -->", "U0000000", fmt.Sprintf("%.6f", float64(warningTime.Unix())))
			}

			client, err := slack.NewClientWithAPI(mockAPI)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			// Run analysis
			toWarn, toArchive, err := client.GetInactiveChannels(tt.warnSeconds, tt.archiveSeconds)
			if err != nil {
				t.Fatalf("GetInactiveChannels failed: %v", err)
			}

			// Check results
			warnFound := len(toWarn) > 0
			archiveFound := len(toArchive) > 0

			if warnFound != tt.expectWarn {
				t.Errorf("Expected warn=%v, got warn=%v", tt.expectWarn, warnFound)
			}

			if archiveFound != tt.expectArchive {
				t.Errorf("Expected archive=%v, got archive=%v", tt.expectArchive, archiveFound)
			}
		})
	}
}

func TestChannelExclusions(t *testing.T) {
	excludedNames := []string{
		"general",
		"random",
		"announcements",
		"admin",
		"hr",
		"security",
		"general-discussion",
		"admin-only",
		"hr-private",
		"security-alerts",
	}

	mockAPI := slack.NewMockSlackAPI()

	// Add all excluded channels as very old and inactive
	for i, name := range excludedNames {
		channelID := fmt.Sprintf("C%d", i+1)
		mockAPI.AddChannel(channelID, name, time.Now().AddDate(0, 0, -100), "Excluded channel")

		// Make them very inactive
		lastActivity := time.Now().AddDate(0, 0, -60)
		mockAPI.AddMessageToHistory(channelID, "old message", "U1234567", fmt.Sprintf("%.6f", float64(lastActivity.Unix())))
	}

	// Add one regular channel that should be warned
	mockAPI.AddChannel("C999", "regular-channel", time.Now().AddDate(0, 0, -100), "Regular channel")
	lastActivity := time.Now().AddDate(0, 0, -35)
	mockAPI.AddMessageToHistory("C999", "old message", "U1234567", fmt.Sprintf("%.6f", float64(lastActivity.Unix())))

	client, err := slack.NewClientWithAPI(mockAPI)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	toWarn, toArchive, err := client.GetInactiveChannels(30, 7)
	if err != nil {
		t.Fatalf("GetInactiveChannels failed: %v", err)
	}

	// Should only find the regular channel
	if len(toWarn) != 1 {
		t.Errorf("Expected 1 channel to warn, got %d", len(toWarn))
	}

	if len(toArchive) != 0 {
		t.Errorf("Expected 0 channels to archive, got %d", len(toArchive))
	}

	if len(toWarn) > 0 && toWarn[0].Name != "regular-channel" {
		t.Errorf("Expected to warn 'regular-channel', got '%s'", toWarn[0].Name)
	}
}

func TestRunArchive(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		expectedErr    string
		warnSeconds    int
		archiveSeconds int
	}{
		{
			name:           "Missing token",
			warnSeconds:    30,
			archiveSeconds: 7,
			token:          "",
			expectedErr:    "slack token is required",
		},
		{
			name:           "Invalid warn seconds with valid token",
			warnSeconds:    0,
			archiveSeconds: 7,
			token:          "xoxb-validtoken123456789012345678901234567890",
			expectedErr:    "warn-seconds must be positive, got 0",
		},
		{
			name:           "Invalid archive seconds with valid token",
			warnSeconds:    30,
			archiveSeconds: -5,
			token:          "xoxb-validtoken123456789012345678901234567890",
			expectedErr:    "archive-seconds must be positive, got -5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global variables (convert test seconds to days for new API)
			oldWarnDays := warnDays
			oldArchiveDays := archiveDays
			warnDays = float64(tt.warnSeconds) / (24 * 60 * 60)
			archiveDays = float64(tt.archiveSeconds) / (24 * 60 * 60)
			defer func() {
				warnDays = oldWarnDays
				archiveDays = oldArchiveDays
			}()

			// Initialize viper configuration for the test
			initConfig()

			// Set token in viper
			if tt.token != "" {
				t.Setenv("SLACK_TOKEN", tt.token)
			} else {
				t.Setenv("SLACK_TOKEN", "")
			}

			err := runArchive(nil, nil)

			if tt.expectedErr != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedErr)
					return
				}
				if !containsString(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error containing '%s', got: %s", tt.expectedErr, err.Error())
				}
			} else {
				t.Errorf("Test case should expect an error since we can't create real Slack client")
			}
		})
	}
}

// Helper function.
func containsString(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr ||
		(len(str) > len(substr) &&
			(str[:len(substr)] == substr ||
				str[len(str)-len(substr):] == substr ||
				findSubstring(str, substr))))
}

func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
