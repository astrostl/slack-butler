package slack

import (
	"github.com/slack-go/slack"
)

// SlackAPI defines the interface for Slack API operations
type SlackAPI interface {
	AuthTest() (*slack.AuthTestResponse, error)
	GetConversations(params *slack.GetConversationsParameters) ([]slack.Channel, string, error)
	GetConversationHistory(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error)
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
}

// RealSlackAPI wraps the actual Slack API client
type RealSlackAPI struct {
	client *slack.Client
}

// NewRealSlackAPI creates a new real Slack API wrapper
func NewRealSlackAPI(token string) *RealSlackAPI {
	return &RealSlackAPI{
		client: slack.New(token),
	}
}

func (r *RealSlackAPI) AuthTest() (*slack.AuthTestResponse, error) {
	return r.client.AuthTest()
}

func (r *RealSlackAPI) GetConversations(params *slack.GetConversationsParameters) ([]slack.Channel, string, error) {
	return r.client.GetConversations(params)
}

func (r *RealSlackAPI) GetConversationHistory(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error) {
	return r.client.GetConversationHistory(params)
}

func (r *RealSlackAPI) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	return r.client.PostMessage(channelID, options...)
}