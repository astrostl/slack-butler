package slack

import (
	"fmt"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

type Client struct {
	api *slack.Client
}

type Channel struct {
	ID      string
	Name    string
	Created time.Time
	Purpose string
}

func NewClient(token string) (*Client, error) {
	api := slack.New(token)
	
	auth, err := api.AuthTest()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}

	fmt.Printf("Connected as: %s (team: %s)\n", auth.User, auth.Team)
	return &Client{api: api}, nil
}

func (c *Client) GetNewChannels(since time.Time) ([]Channel, error) {
	channels, _, err := c.api.GetConversations(&slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
		Limit: 1000,
	})
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "missing_scope") {
			return nil, fmt.Errorf("missing required permissions. Your bot needs these OAuth scopes:\n  - channels:read (to list public channels)\n  - groups:read (to list private channels)\n\nPlease add these scopes in your Slack app settings at https://api.slack.com/apps")
		}
		if strings.Contains(errStr, "invalid_auth") {
			return nil, fmt.Errorf("invalid token. Please check your SLACK_TOKEN")
		}
		return nil, fmt.Errorf("failed to get conversations: %v", err)
	}

	var newChannels []Channel
	for _, ch := range channels {
		created := time.Unix(int64(ch.Created), 0)
		if created.After(since) {
			newChannels = append(newChannels, Channel{
				ID:      ch.ID,
				Name:    ch.Name,
				Created: created,
				Purpose: ch.Purpose.Value,
			})
		}
	}

	return newChannels, nil
}

func (c *Client) FormatNewChannelAnnouncement(channels []Channel, since time.Time) string {
	var builder strings.Builder
	
	if len(channels) == 1 {
		builder.WriteString("🆕 New channel alert!")
	} else {
		builder.WriteString(fmt.Sprintf("🆕 %d new channels created!", len(channels)))
	}
	
	builder.WriteString(fmt.Sprintf("\n\nChannels created since %s:\n", since.Format("2006-01-02 15:04")))
	
	for _, ch := range channels {
		builder.WriteString(fmt.Sprintf("• <#%s> - created %s", ch.ID, ch.Created.Format("2006-01-02 15:04")))
		if ch.Purpose != "" {
			builder.WriteString(fmt.Sprintf("\n  Purpose: %s", ch.Purpose))
		}
		builder.WriteString("\n")
	}
	
	return builder.String()
}

func (c *Client) PostMessage(channel, message string) error {
	channelID := strings.TrimPrefix(channel, "#")
	
	_, _, err := c.api.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "missing_scope") {
			return fmt.Errorf("missing required permission to post messages. Your bot needs the 'chat:write' OAuth scope.\nPlease add this scope in your Slack app settings at https://api.slack.com/apps")
		}
		if strings.Contains(errStr, "channel_not_found") {
			return fmt.Errorf("channel '%s' not found. Make sure the bot is added to the channel", channel)
		}
		if strings.Contains(errStr, "not_in_channel") {
			return fmt.Errorf("bot is not a member of channel '%s'. Please add the bot to the channel", channel)
		}
		return fmt.Errorf("failed to post message to %s: %v", channel, err)
	}
	
	return nil
}