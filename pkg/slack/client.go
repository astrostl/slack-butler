package slack

import (
	"context"
	"fmt"
	"regexp"
	"github.com/astrostl/slack-buddy-ai/pkg/logger"
	"strings"
	"sync"
	"time"

	"github.com/slack-go/slack"
)

type RateLimiter struct {
	mu           sync.Mutex
	lastRequest  time.Time
	minInterval  time.Duration
	backoffCount int
	maxBackoff   time.Duration
}

type Client struct {
	api         SlackAPI
	rateLimiter *RateLimiter
}

type Channel struct {
	ID      string
	Name    string
	Created time.Time
	Purpose string
	Creator string
}

type AuthInfo struct {
	User   string
	UserID string
	Team   string
	TeamID string
}

func NewClient(token string) (*Client, error) {
	// Validate token format before using it
	if err := ValidateSlackToken(token); err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	api := NewRealSlackAPI(token)
	
	auth, err := api.AuthTest()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %v", SanitizeForLogging(err.Error()))
	}

	logger.WithFields(logger.LogFields{
		"user": auth.User,
		"team": auth.Team,
	}).Debug("Successfully connected to Slack")
	// Connection info logged but not printed to reduce output noise
	return &Client{
		api: api,
		rateLimiter: &RateLimiter{
			minInterval: time.Second,     // 1 request per second baseline
			maxBackoff:  time.Minute * 5, // Max 5 minute backoff
		},
	}, nil
}

// NewClientWithAPI creates a client with a custom API implementation (for testing)
func NewClientWithAPI(api SlackAPI) (*Client, error) {
	if api == nil {
		return nil, fmt.Errorf("API cannot be nil")
	}
	
	auth, err := api.AuthTest()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}

	logger.WithFields(logger.LogFields{
		"user": auth.User,
		"team": auth.Team,
	}).Debug("Successfully connected to Slack")
	// Connection info logged but not printed to reduce output noise
	return &Client{
		api: api,
		rateLimiter: &RateLimiter{
			minInterval: time.Second,     // 1 request per second baseline
			maxBackoff:  time.Minute * 5, // Max 5 minute backoff
		},
	}, nil
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	timeSinceLastRequest := now.Sub(rl.lastRequest)
	
	// Calculate wait time with exponential backoff
	waitTime := rl.minInterval
	if rl.backoffCount > 0 {
		backoffMultiplier := time.Duration(1 << rl.backoffCount) // 2^backoffCount
		waitTime = rl.minInterval * backoffMultiplier
		if waitTime > rl.maxBackoff {
			waitTime = rl.maxBackoff
		}
	}

	if timeSinceLastRequest < waitTime {
		sleepTime := waitTime - timeSinceLastRequest
		logger.WithFields(logger.LogFields{
			"sleep_duration": sleepTime.String(),
			"backoff_count": rl.backoffCount,
		}).Debug("Rate limiting: sleeping before API request")

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleepTime):
		}
	}

	rl.lastRequest = time.Now()
	return nil
}

func (rl *RateLimiter) OnSuccess() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	// Reset backoff on successful request
	if rl.backoffCount > 0 {
		logger.WithField("previous_backoff", rl.backoffCount).Debug("Rate limiting: resetting backoff after success")
		rl.backoffCount = 0
	}
}

func (rl *RateLimiter) OnRateLimitError() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	rl.backoffCount++
	if rl.backoffCount > 6 { // Cap at 2^6 = 64 seconds
		rl.backoffCount = 6
	}
	
	logger.WithField("backoff_count", rl.backoffCount).Warn("Rate limiting: increasing backoff due to rate limit error")
}

func (c *Client) GetNewChannels(since time.Time) ([]Channel, error) {
	logger.WithField("since", since.Format("2006-01-02 15:04:05")).Debug("Fetching channels from Slack API")
	
	// Rate limit before API call
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter cancelled: %v", err)
	}
	
	channels, _, err := c.api.GetConversations(&slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
		Limit: 1000,
	})
	if err != nil {
		errStr := err.Error()
		logger.WithFields(logger.LogFields{
			"error": errStr,
			"operation": "get_conversations",
		}).Error("Slack API error")
		
		// Handle rate limiting
		if strings.Contains(errStr, "rate_limited") {
			c.rateLimiter.OnRateLimitError()
			return nil, fmt.Errorf("rate limited by Slack API. Will retry with exponential backoff on next request")
		}
		
		if strings.Contains(errStr, "missing_scope") {
			logger.Error("Missing OAuth scopes for channel access")
			return nil, fmt.Errorf("missing required permissions. Your bot needs these OAuth scopes:\n  - channels:read (to list public channels)\n  - groups:read (to list private channels)\n\nPlease add these scopes in your Slack app settings at https://api.slack.com/apps")
		}
		if strings.Contains(errStr, "invalid_auth") {
			logger.Error("Invalid Slack authentication token")
			return nil, fmt.Errorf("invalid token. Please check your SLACK_TOKEN")
		}
		return nil, fmt.Errorf("failed to get conversations: %v", err)
	}

	// Mark successful API call
	c.rateLimiter.OnSuccess()

	logger.WithField("total_channels", len(channels)).Debug("Retrieved channels from Slack API")
	
	var newChannels []Channel
	for _, ch := range channels {
		created := time.Unix(int64(ch.Created), 0)
		if created.After(since) {
			logger.WithFields(logger.LogFields{
				"channel_id": ch.ID,
				"channel_name": ch.Name,
				"created": created.Format("2006-01-02 15:04:05"),
			}).Debug("Found new channel")
			
			newChannels = append(newChannels, Channel{
				ID:      ch.ID,
				Name:    ch.Name,
				Created: created,
				Purpose: ch.Purpose.Value,
				Creator: ch.Creator,
			})
		}
	}

	logger.WithField("new_channels_count", len(newChannels)).Debug("Channel detection completed")
	return newChannels, nil
}

func (c *Client) FormatNewChannelAnnouncement(channels []Channel, since time.Time) string {
	logger.WithFields(logger.LogFields{
		"channel_count": len(channels),
		"since": since.Format("2006-01-02 15:04:05"),
	}).Debug("Formatting announcement message")
	
	var builder strings.Builder
	
	if len(channels) == 1 {
		builder.WriteString("New channel alert!")
	} else {
		builder.WriteString(fmt.Sprintf("%d new channels created!", len(channels)))
	}
	
	builder.WriteString("\n\n")
	
	for i, ch := range channels {
		builder.WriteString(fmt.Sprintf("• <#%s>", ch.ID))
		
		// Build the "created [DATE] by [USER]" line
		createdLine := fmt.Sprintf(" - created %s", ch.Created.Format("January 2, 2006"))
		if ch.Creator != "" {
			createdLine += fmt.Sprintf(" by <@%s>", ch.Creator)
		}
		builder.WriteString(createdLine)
		
		if ch.Purpose != "" {
			builder.WriteString(fmt.Sprintf("\n  Purpose: %s", ch.Purpose))
		}
		
		// Add spacing between channels (but not after the last one)
		if i < len(channels)-1 {
			builder.WriteString("\n\n")
		} else {
			builder.WriteString("\n")
		}
	}
	
	return builder.String()
}

func (c *Client) PostMessage(channel, message string) error {
	logger.WithFields(logger.LogFields{
		"channel": channel,
		"message_length": len(message),
	}).Debug("Attempting to post message to channel")
	
	// Validate channel name format
	if err := ValidateChannelName(channel); err != nil {
		logger.WithFields(logger.LogFields{
			"channel": channel,
			"error": err.Error(),
		}).Error("Invalid channel name format")
		return fmt.Errorf("invalid channel name '%s': %v", channel, err)
	}

	// Resolve channel name to channel ID
	channelID, err := c.resolveChannelNameToID(channel)
	if err != nil {
		return fmt.Errorf("failed to find channel %s: %v", channel, err)
	}
	
	// Rate limit before API call
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter cancelled: %v", err)
	}
	
	_, _, err = c.api.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil {
		errStr := err.Error()
		logger.WithFields(logger.LogFields{
			"channel": channel,
			"error": errStr,
			"operation": "post_message",
		}).Error("Failed to post message to Slack")
		
		// Handle rate limiting
		if strings.Contains(errStr, "rate_limited") {
			c.rateLimiter.OnRateLimitError()
			return fmt.Errorf("rate limited by Slack API. Will retry with exponential backoff on next request")
		}
		
		if strings.Contains(errStr, "missing_scope") {
			logger.Error("Missing chat:write OAuth scope")
			return fmt.Errorf("missing required permission to post messages. Your bot needs the 'chat:write' OAuth scope.\nPlease add this scope in your Slack app settings at https://api.slack.com/apps")
		}
		if strings.Contains(errStr, "channel_not_found") {
			logger.WithField("channel", channel).Error("Channel not found")
			return fmt.Errorf("channel '%s' not found. Make sure the bot is added to the channel", channel)
		}
		if strings.Contains(errStr, "not_in_channel") {
			logger.WithField("channel", channel).Error("Bot not in channel")
			return fmt.Errorf("bot is not a member of channel '%s'. Please add the bot to the channel", channel)
		}
		return fmt.Errorf("failed to post message to %s: %v", channel, err)
	}
	
	// Mark successful API call
	c.rateLimiter.OnSuccess()
	
	logger.WithField("channel", channel).Info("Message posted successfully")
	return nil
}

func (c *Client) TestAuth() (*AuthInfo, error) {
	auth, err := c.api.AuthTest()
	if err != nil {
		return nil, err
	}
	
	return &AuthInfo{
		User:   auth.User,
		UserID: auth.UserID,
		Team:   auth.Team,
		TeamID: auth.TeamID,
	}, nil
}

func (c *Client) GetChannelInfo(channelID string) (*Channel, error) {
	// This is used for permission testing in health checks
	// We'll just return a mock error for permission testing
	return nil, fmt.Errorf("channel_not_found")
}

// GetPreviouslyAnnouncedChannels retrieves channel IDs that have been previously announced
// in the specified announcement channel by parsing the message history from our bot only
func (c *Client) GetPreviouslyAnnouncedChannels(announcementChannel string) (map[string]bool, error) {
	logger.WithField("announcement_channel", announcementChannel).Debug("Fetching previously announced channels")
	
	// First get our bot's user ID for filtering
	authInfo, err := c.TestAuth()
	if err != nil {
		return nil, fmt.Errorf("failed to get bot user ID: %v", err)
	}
	
	// Resolve channel name to channel ID
	channelID, err := c.resolveChannelNameToID(announcementChannel)
	if err != nil {
		return nil, fmt.Errorf("failed to find channel %s: %v", announcementChannel, err)
	}
	
	// Rate limit before API call
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter cancelled: %v", err)
	}
	
	// Get conversation history - limit to 1000 messages to look back reasonably far
	params := &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     1000,
	}
	
	response, err := c.api.GetConversationHistory(params)
	if err != nil {
		errStr := err.Error()
		
		// Debug: Log the exact error from Slack API
		logger.WithFields(logger.LogFields{
			"channel": announcementChannel,
			"raw_error": errStr,
		}).Debug("Raw Slack API error for conversation history")
		
		// Handle rate limiting
		if strings.Contains(errStr, "rate_limited") {
			c.rateLimiter.OnRateLimitError()
			return nil, fmt.Errorf("rate limited by Slack API. Will retry with exponential backoff on next request")
		}
		
		if strings.Contains(errStr, "missing_scope") {
			return nil, fmt.Errorf("missing required permission: your bot needs the 'channels:history' OAuth scope to prevent duplicate announcements.\n\nPlease add this scope in your Slack app settings at https://api.slack.com/apps")
		}
		if strings.Contains(errStr, "channel_not_found") {
			return nil, fmt.Errorf("cannot access channel '%s' to read message history. Make sure the bot is a member of the channel and has the 'channels:history' permission", announcementChannel)
		}
		return nil, fmt.Errorf("failed to get history from %s: %v", announcementChannel, err)
	}
	
	// Mark successful API call
	c.rateLimiter.OnSuccess()
	
	// Parse messages to extract channel IDs that have been announced by our bot only
	announcedChannels := make(map[string]bool)
	botMessagesChecked := 0
	
	for _, message := range response.Messages {
		// Only check messages posted by our bot
		if message.User != authInfo.UserID {
			continue
		}
		
		botMessagesChecked++
		
		// Look for channel mentions in the format <#CHANNEL_ID>
		// This matches our announcement format
		channelIDs := extractChannelIDs(message.Text)
		for _, channelID := range channelIDs {
			announcedChannels[channelID] = true
		}
	}
	
	logger.WithFields(logger.LogFields{
		"announcement_channel": announcementChannel,
		"total_messages_checked": len(response.Messages),
		"bot_messages_checked": botMessagesChecked,
		"bot_user_id": authInfo.UserID,
		"previously_announced_count": len(announcedChannels),
	}).Debug("Completed parsing announcement history")
	
	return announcedChannels, nil
}

// FilterAlreadyAnnouncedChannels removes channels that have already been announced
func (c *Client) FilterAlreadyAnnouncedChannels(channels []Channel, announcementChannel string) ([]Channel, error) {
	if announcementChannel == "" {
		// No announcement channel specified, return all channels
		return channels, nil
	}
	
	// Get previously announced channels
	previouslyAnnounced, err := c.GetPreviouslyAnnouncedChannels(announcementChannel)
	if err != nil {
		return nil, fmt.Errorf("failed to get previously announced channels: %v", err)
	}
	
	// Filter out already announced channels
	var filteredChannels []Channel
	for _, channel := range channels {
		if !previouslyAnnounced[channel.ID] {
			filteredChannels = append(filteredChannels, channel)
			logger.WithField("channel", channel.Name).Debug("Channel not previously announced, including in announcement")
		} else {
			logger.WithField("channel", channel.Name).Debug("Channel previously announced, skipping")
		}
	}
	
	logger.WithFields(logger.LogFields{
		"original_count": len(channels),
		"filtered_count": len(filteredChannels),
		"skipped_count": len(channels) - len(filteredChannels),
	}).Info("Filtered already announced channels")
	
	return filteredChannels, nil
}

// extractChannelIDs extracts channel IDs from Slack message text in the format <#CHANNEL_ID>
func extractChannelIDs(text string) []string {
	// Regular expression to match Slack channel mentions: <#CHANNEL_ID> or <#CHANNEL_ID|channel-name>
	re := regexp.MustCompile(`<#([A-Z0-9]+)(?:\|[^>]+)?>`)
	matches := re.FindAllStringSubmatch(text, -1)
	
	// Initialize with empty slice to ensure we never return nil
	channelIDs := []string{}
	for _, match := range matches {
		if len(match) > 1 {
			channelIDs = append(channelIDs, match[1])
		}
	}
	
	return channelIDs
}

// resolveChannelNameToID converts a channel name (like "#general" or "general") to its Slack channel ID
func (c *Client) resolveChannelNameToID(channelName string) (string, error) {
	// Clean the channel name
	cleanName := strings.TrimPrefix(channelName, "#")
	
	// Rate limit before API call
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()
	
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter cancelled: %v", err)
	}
	
	// Get all channels to find the matching one
	params := &slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
		Limit: 1000,
	}
	
	channels, _, err := c.api.GetConversations(params)
	if err != nil {
		c.rateLimiter.OnRateLimitError()
		return "", fmt.Errorf("failed to get channels: %v", err)
	}
	
	// Mark successful API call
	c.rateLimiter.OnSuccess()
	
	// Find channel by name
	for _, channel := range channels {
		if channel.Name == cleanName {
			return channel.ID, nil
		}
	}
	
	return "", fmt.Errorf("channel '%s' not found", channelName)
}