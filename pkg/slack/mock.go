package slack

import (
	"errors"
	"fmt"
	"time"

	"github.com/slack-go/slack"
)

// Mock error constants.
const (
	missingScope    = "missing_scope"
	channelNotFound = "channel_not_found"
)

// MockSlackAPI implements SlackAPI for testing.
type MockSlackAPI struct {

	// Error fields (16 bytes each on 64-bit) - grouped to minimize padding
	AuthTestError               error
	GetConversationsError       error
	GetConversationHistoryError error
	PostMessageError            error
	ArchiveConversationError    error
	JoinConversationError       error
	GetUsersError               error
	GetTeamInfoError            error

	// Map fields (8 bytes each on 64-bit) - grouped together
	ConversationHistory       map[string][]slack.Message
	ConversationHistoryErrors map[string]error
	ArchiveConversationErrors map[string]error
	JoinConversationErrors    map[string]error

	// Pointer fields (8 bytes each on 64-bit) - at end to minimize padding
	AuthTestResponse *slack.AuthTestResponse
	TeamInfo         *slack.TeamInfo
	// Slice fields (24 bytes each on 64-bit)
	Channels         []slack.Channel
	PostedMessages   []MockMessage
	ArchivedChannels []string
	JoinedChannels   []string
	Users            []slack.User
}

type MockMessage struct {
	ChannelID string
	Text      string
}

// NewMockSlackAPI creates a new mock Slack API.
func NewMockSlackAPI() *MockSlackAPI {
	return &MockSlackAPI{
		AuthTestResponse: &slack.AuthTestResponse{
			User:   "test-bot",
			UserID: "U0000000", // Bot's user ID for filtering
			Team:   "Test Team",
			TeamID: "T0000000",
		},
		TeamInfo: &slack.TeamInfo{
			ID:     "T0000000",
			Name:   "Test Team",
			Domain: "testteam",
		},
		Channels:                  []slack.Channel{},
		ConversationHistory:       make(map[string][]slack.Message),
		ConversationHistoryErrors: make(map[string]error),
		PostedMessages:            []MockMessage{},
		ArchivedChannels:          []string{},
		ArchiveConversationErrors: make(map[string]error),
		JoinedChannels:            []string{},
		JoinConversationErrors:    make(map[string]error),
		Users:                     []slack.User{},
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

func (m *MockSlackAPI) GetConversationHistory(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error) {
	// Check for channel-specific errors first
	if err, exists := m.ConversationHistoryErrors[params.ChannelID]; exists && err != nil {
		return nil, err
	}

	// Check for global error
	if m.GetConversationHistoryError != nil {
		return nil, m.GetConversationHistoryError
	}

	messages, exists := m.ConversationHistory[params.ChannelID]
	if !exists {
		messages = []slack.Message{}
	}

	// Slack API returns messages in reverse chronological order (newest first)
	// So we need to reverse our stored messages
	reversedMessages := make([]slack.Message, len(messages))
	for i, msg := range messages {
		reversedMessages[len(messages)-1-i] = msg
	}

	return &slack.GetConversationHistoryResponse{
		Messages: reversedMessages,
	}, nil
}

func (m *MockSlackAPI) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	if m.PostMessageError != nil {
		return "", "", m.PostMessageError
	}

	// For testing purposes, we'll just record that a message was posted
	// Extracting the actual text from slack options is complex, so we use a placeholder
	message := MockMessage{
		ChannelID: channelID,
		Text:      "mock-message-posted", // Simplified for testing
	}
	m.PostedMessages = append(m.PostedMessages, message)

	return "mock-channel-id", "mock-timestamp", nil
}

func (m *MockSlackAPI) ArchiveConversation(channelID string) error {
	// Check for channel-specific errors first
	if err, exists := m.ArchiveConversationErrors[channelID]; exists && err != nil {
		return err
	}

	// Check for global error
	if m.ArchiveConversationError != nil {
		return m.ArchiveConversationError
	}

	m.ArchivedChannels = append(m.ArchivedChannels, channelID)
	return nil
}

func (m *MockSlackAPI) JoinConversation(channelID string) (*slack.Channel, string, []string, error) {
	// Check for channel-specific errors first
	if err, exists := m.JoinConversationErrors[channelID]; exists && err != nil {
		return nil, "", nil, err
	}

	// Check for global error
	if m.JoinConversationError != nil {
		return nil, "", nil, m.JoinConversationError
	}

	m.JoinedChannels = append(m.JoinedChannels, channelID)

	// Return a mock channel for the joined conversation
	mockChannel := &slack.Channel{
		GroupConversation: slack.GroupConversation{
			Conversation: slack.Conversation{
				ID: channelID,
			},
		},
	}

	return mockChannel, "", []string{}, nil
}

func (m *MockSlackAPI) GetUsers() ([]slack.User, error) {
	if m.GetUsersError != nil {
		return nil, m.GetUsersError
	}
	return m.Users, nil
}

func (m *MockSlackAPI) GetTeamInfo() (*slack.TeamInfo, error) {
	if m.GetTeamInfoError != nil {
		return nil, m.GetTeamInfoError
	}
	return m.TeamInfo, nil
}

func (m *MockSlackAPI) GetConversationInfo(input *slack.GetConversationInfoInput) (*slack.Channel, error) {
	// Find channel by ID in the mock channels list
	for _, ch := range m.Channels {
		if ch.ID == input.ChannelID {
			return &ch, nil
		}
	}
	return nil, fmt.Errorf("channel not found: %s", input.ChannelID)
}

func (m *MockSlackAPI) GetConversationsForUser(params *slack.GetConversationsForUserParameters) ([]slack.Channel, string, error) {
	// Return all channels for simplicity in tests
	// In real implementation, this would filter by user membership
	return m.Channels, "", nil
}

// Helper methods for testing

func (m *MockSlackAPI) AddChannel(id, name string, created time.Time, purpose string) {
	m.AddChannelWithCreator(id, name, created, purpose, "U1234567")
}

func (m *MockSlackAPI) AddChannelWithCreator(id, name string, created time.Time, purpose, creator string) {
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
			Creator: creator,
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
	case missingScope:
		m.PostMessageError = errors.New(missingScope)
	case channelNotFound:
		m.PostMessageError = errors.New(channelNotFound)
	case "not_in_channel":
		m.PostMessageError = fmt.Errorf("not_in_channel")
	case "rate_limited":
		m.PostMessageError = fmt.Errorf("rate_limited")
	case "invalid_auth":
		m.PostMessageError = fmt.Errorf("invalid_auth")
	case "generic_error":
		m.PostMessageError = fmt.Errorf("generic error")
	case "":
		m.PostMessageError = nil
	default:
		m.PostMessageError = fmt.Errorf("%s", errorType)
	}
}

func (m *MockSlackAPI) SetJoinError(errorType string) {
	switch errorType {
	case "rate_limited":
		m.JoinConversationError = fmt.Errorf("rate_limited")
	case missingScope:
		m.JoinConversationError = errors.New(missingScope)
	case "invalid_auth":
		m.JoinConversationError = fmt.Errorf("invalid_auth")
	case "already_in_channel":
		m.JoinConversationError = fmt.Errorf("already_in_channel")
	case "is_archived":
		m.JoinConversationError = fmt.Errorf("is_archived")
	case "invite_only":
		m.JoinConversationError = fmt.Errorf("invite_only")
	case "":
		m.JoinConversationError = nil
	default:
		m.JoinConversationError = fmt.Errorf("%s", errorType)
	}
}

// Additional helper methods for specific error types.
func (m *MockSlackAPI) SetMissingScopeError(hasError bool) {
	if hasError {
		m.GetConversationsError = errors.New(missingScope)
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

func (m *MockSlackAPI) AddMessageToHistory(channelID, text, user string, timestamp string) {
	if m.ConversationHistory == nil {
		m.ConversationHistory = make(map[string][]slack.Message)
	}

	message := slack.Message{
		Msg: slack.Msg{
			Type:      "message",
			Text:      text,
			User:      user,
			Timestamp: timestamp,
		},
	}

	m.ConversationHistory[channelID] = append(m.ConversationHistory[channelID], message)
}

func (m *MockSlackAPI) SetConversationHistoryError(hasError bool) {
	if hasError {
		m.GetConversationHistoryError = fmt.Errorf("failed to get conversation history")
	} else {
		m.GetConversationHistoryError = nil
	}
}

// Simulate specific error types.
func (m *MockSlackAPI) SimulateMissingScopeError() {
	m.SetMissingScopeError(true)
}

func (m *MockSlackAPI) SimulateInvalidAuthError() {
	m.SetAuthError(true)
}

func (m *MockSlackAPI) SimulateChannelNotFoundError() {
	m.SetPostMessageError(channelNotFound)
}

func (m *MockSlackAPI) SimulateNotInChannelError() {
	m.SetPostMessageError("not_in_channel")
}

func (m *MockSlackAPI) SetArchiveConversationError(errorType string) {
	switch errorType {
	case missingScope:
		m.ArchiveConversationError = fmt.Errorf("%s", missingScope)
	case channelNotFound:
		m.ArchiveConversationError = fmt.Errorf("%s", channelNotFound)
	case "already_archived":
		m.ArchiveConversationError = fmt.Errorf("already_archived")
	case "":
		m.ArchiveConversationError = nil
	default:
		m.ArchiveConversationError = fmt.Errorf("%s", errorType)
	}
}

func (m *MockSlackAPI) GetArchivedChannels() []string {
	return m.ArchivedChannels
}

func (m *MockSlackAPI) ClearArchivedChannels() {
	m.ArchivedChannels = []string{}
}

func (m *MockSlackAPI) SetJoinConversationError(errorType string) {
	switch errorType {
	case missingScope:
		m.JoinConversationError = fmt.Errorf("%s", missingScope)
	case channelNotFound:
		m.JoinConversationError = fmt.Errorf("%s", channelNotFound)
	case "already_in_channel":
		m.JoinConversationError = fmt.Errorf("already_in_channel")
	case "":
		m.JoinConversationError = nil
	default:
		m.JoinConversationError = fmt.Errorf("%s", errorType)
	}
}

func (m *MockSlackAPI) GetJoinedChannels() []string {
	return m.JoinedChannels
}

func (m *MockSlackAPI) ClearJoinedChannels() {
	m.JoinedChannels = []string{}
}

func (m *MockSlackAPI) AddUser(id, name, realName string) {
	user := slack.User{
		ID:       id,
		Name:     name,
		RealName: realName,
		Profile: slack.UserProfile{
			DisplayName: realName,
		},
	}
	m.Users = append(m.Users, user)
}

func (m *MockSlackAPI) SetGetUsersError(errorType string) {
	switch errorType {
	case missingScope:
		m.GetUsersError = fmt.Errorf("%s", missingScope)
	case "":
		m.GetUsersError = nil
	default:
		m.GetUsersError = fmt.Errorf("%s", errorType)
	}
}

func (m *MockSlackAPI) SetGetTeamInfoError(errorType string) {
	switch errorType {
	case missingScope:
		m.GetTeamInfoError = fmt.Errorf("%s", missingScope)
	case "":
		m.GetTeamInfoError = nil
	default:
		m.GetTeamInfoError = fmt.Errorf("%s", errorType)
	}
}

// Additional helper methods for archive testing

// MockMessage represents a message in conversation history for testing.
type MockHistoryMessage struct {
	Timestamp string
	User      string
	Text      string
	SubType   string
}

// SetChannelHistory sets up mock conversation history for a channel.
func (m *MockSlackAPI) SetChannelHistory(channelID string, messages []MockHistoryMessage) {
	if m.ConversationHistory == nil {
		m.ConversationHistory = make(map[string][]slack.Message)
	}

	slackMessages := make([]slack.Message, len(messages))
	for i, msg := range messages {
		slackMessages[i] = slack.Message{
			Msg: slack.Msg{
				Type:      "message",
				Text:      msg.Text,
				User:      msg.User,
				Timestamp: msg.Timestamp,
				SubType:   msg.SubType,
			},
		}
	}

	m.ConversationHistory[channelID] = slackMessages
}

// SetBotUserID sets the bot user ID for testing.
func (m *MockSlackAPI) SetBotUserID(userID string) {
	if m.AuthTestResponse == nil {
		m.AuthTestResponse = &slack.AuthTestResponse{}
	}
	m.AuthTestResponse.UserID = userID
}

// SetGetConversationHistoryError sets an error for a specific channel's history.
func (m *MockSlackAPI) SetGetConversationHistoryError(channelID string, hasError bool) {
	if m.ConversationHistoryErrors == nil {
		m.ConversationHistoryErrors = make(map[string]error)
	}

	if hasError {
		m.ConversationHistoryErrors[channelID] = fmt.Errorf("failed to get conversation history")
	} else {
		delete(m.ConversationHistoryErrors, channelID)
	}
}

// SetJoinConversationErrorForChannel sets an error for joining a specific channel.
func (m *MockSlackAPI) SetJoinConversationErrorForChannel(channelID string, hasError bool) {
	if m.JoinConversationErrors == nil {
		m.JoinConversationErrors = make(map[string]error)
	}

	if hasError {
		m.JoinConversationErrors[channelID] = fmt.Errorf("failed to join conversation")
	} else {
		delete(m.JoinConversationErrors, channelID)
	}
}

// SetArchiveConversationErrorWithMessage sets a specific error message for archiving a channel.
func (m *MockSlackAPI) SetArchiveConversationErrorWithMessage(channelID string, hasError bool, errorType string) {
	if m.ArchiveConversationErrors == nil {
		m.ArchiveConversationErrors = make(map[string]error)
	}

	if hasError {
		switch errorType {
		case missingScope:
			m.ArchiveConversationErrors[channelID] = fmt.Errorf("%s", missingScope)
		case channelNotFound:
			m.ArchiveConversationErrors[channelID] = fmt.Errorf("%s", channelNotFound)
		case "already_archived":
			m.ArchiveConversationErrors[channelID] = fmt.Errorf("already_archived")
		default:
			m.ArchiveConversationErrors[channelID] = fmt.Errorf("%s", errorType)
		}
	} else {
		delete(m.ArchiveConversationErrors, channelID)
	}
}
