package slack

import (
	"fmt"
	"time"

	"github.com/slack-go/slack"
)

// MockSlackAPI implements SlackAPI for testing
type MockSlackAPI struct {
	AuthTestResponse    *slack.AuthTestResponse
	AuthTestError       error
	Channels           []slack.Channel
	GetConversationsError error
	PostMessageError    error
	PostedMessages      []MockMessage
}

type MockMessage struct {
	ChannelID string
	Text      string
}

// NewMockSlackAPI creates a new mock Slack API
func NewMockSlackAPI() *MockSlackAPI {
	return &MockSlackAPI{
		AuthTestResponse: &slack.AuthTestResponse{
			User: "test-bot",
			Team: "Test Team",
		},
		Channels:       []slack.Channel{},
		PostedMessages: []MockMessage{},
	}
}

func (m *MockSlackAPI) AuthTest() (*slack.AuthTestResponse, error) {
	if m.AuthTestError != nil {
		return nil, m.AuthTestError
	}
	return m.AuthTestResponse, nil
}

func (m *MockSlackAPI) GetConversations(params *slack.GetConversationsParameters) ([]slack.Channel, string, error) {
	if m.GetConversationsError != nil {
		return nil, "", m.GetConversationsError
	}
	return m.Channels, "", nil
}

func (m *MockSlackAPI) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	if m.PostMessageError != nil {
		return "", "", m.PostMessageError
	}

	// For testing purposes, we'll just record that a message was posted
	message := MockMessage{
		ChannelID: channelID,
		Text:      "mock-message-text", // Simplified for testing
	}
	m.PostedMessages = append(m.PostedMessages, message)

	return "mock-channel-id", "mock-timestamp", nil
}

// Helper methods for testing

func (m *MockSlackAPI) AddChannel(id, name string, created time.Time, purpose string) {
	channel := slack.Channel{
		GroupConversation: slack.GroupConversation{
			Conversation: slack.Conversation{
				ID:      id,
				Created: slack.JSONTime(created.Unix()),
			},
			Name: name,
			Purpose: slack.Purpose{
				Value: purpose,
			},
		},
	}
	m.Channels = append(m.Channels, channel)
}

func (m *MockSlackAPI) SetAuthError(hasError bool) {
	if hasError {
		m.AuthTestError = fmt.Errorf("authentication failed")
	} else {
		m.AuthTestError = nil
	}
}

func (m *MockSlackAPI) SetGetConversationsError(hasError bool) {
	if hasError {
		m.GetConversationsError = fmt.Errorf("failed to get conversations")
	} else {
		m.GetConversationsError = nil
	}
}

func (m *MockSlackAPI) SetGetConversationsErrorWithMessage(hasError bool, message string) {
	if hasError {
		if message == "" {
			message = "failed to get conversations"
		}
		m.GetConversationsError = fmt.Errorf("%s", message)
	} else {
		m.GetConversationsError = nil
	}
}

func (m *MockSlackAPI) SetPostMessageError(errorType string) {
	switch errorType {
	case "missing_scope":
		m.PostMessageError = fmt.Errorf("missing_scope")
	case "channel_not_found":
		m.PostMessageError = fmt.Errorf("channel_not_found")
	case "not_in_channel":
		m.PostMessageError = fmt.Errorf("not_in_channel")
	case "generic_error":
		m.PostMessageError = fmt.Errorf("generic error")
	case "":
		m.PostMessageError = nil
	default:
		m.PostMessageError = fmt.Errorf("%s", errorType)
	}
}

// Additional helper methods for specific error types
func (m *MockSlackAPI) SetMissingScopeError(hasError bool) {
	if hasError {
		m.GetConversationsError = fmt.Errorf("missing_scope")
	} else {
		m.GetConversationsError = nil
	}
}

func (m *MockSlackAPI) SetInvalidAuthError(hasError bool) {
	if hasError {
		m.GetConversationsError = fmt.Errorf("invalid_auth")
	} else {
		m.GetConversationsError = nil
	}
}

func (m *MockSlackAPI) GetPostedMessages() []MockMessage {
	return m.PostedMessages
}

func (m *MockSlackAPI) ClearPostedMessages() {
	m.PostedMessages = []MockMessage{}
}

// Simulate specific error types
func (m *MockSlackAPI) SimulateMissingScopeError() {
	m.SetMissingScopeError(true)
}

func (m *MockSlackAPI) SimulateInvalidAuthError() {
	m.SetAuthError(true)
}

func (m *MockSlackAPI) SimulateChannelNotFoundError() {
	m.SetPostMessageError("channel_not_found")
}

func (m *MockSlackAPI) SimulateNotInChannelError() {
	m.SetPostMessageError("not_in_channel")
}