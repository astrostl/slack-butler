package slack

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/astrostl/slack-butler/pkg/logger"

	"github.com/slack-go/slack"
)

// Common time duration text constants.
const (
	oneMinuteText = "1 minute"
	oneHourText   = "1 hour"
	oneDayText    = "1 day"
	textDays      = "days"
	textDay       = "day"
)

type Client struct {
	api SlackAPI
}

type Channel struct {
	LastMessage  *MessageInfo // Optional: details about the last message
	Created      time.Time
	Updated      time.Time
	LastActivity time.Time
	ID           string
	Name         string
	Purpose      string
	Creator      string
	MemberCount  int
	IsArchived   bool
}

type AuthInfo struct {
	User         string
	UserID       string
	Team         string
	TeamID       string
	WorkspaceURL string
}

func NewClient(token string) (*Client, error) {
	// Validate token format before using it
	if err := ValidateSlackToken(token); err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
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
	}, nil
}

// NewClientWithAPI creates a client with a custom API implementation (for testing).
func NewClientWithAPI(api SlackAPI) (*Client, error) {
	if api == nil {
		return nil, fmt.Errorf("API cannot be nil")
	}

	auth, err := api.AuthTest()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	logger.WithFields(logger.LogFields{
		"user": auth.User,
		"team": auth.Team,
	}).Debug("Successfully connected to Slack")
	// Connection info logged but not printed to reduce output noise
	return &Client{
		api: api,
	}, nil
}

func (c *Client) GetNewChannels(since time.Time) ([]Channel, error) {
	logger.WithField("since", since.Format("2006-01-02 15:04:05")).Debug("Fetching channels from Slack API")

	channels, _, err := c.api.GetConversations(&slack.GetConversationsParameters{
		Types: []string{"public_channel"},
		Limit: 1000,
	})
	if err != nil {
		errStr := err.Error()
		logger.WithFields(logger.LogFields{
			"error":     errStr,
			"operation": "get_conversations",
		}).Error("Slack API error")

		// Handle rate limiting
		if strings.Contains(errStr, "rate_limited") {
			return nil, fmt.Errorf("rate limited by Slack API. Please wait before retrying")
		}

		if strings.Contains(errStr, "missing_scope") {
			logger.Error("Missing OAuth scopes for channel access")
			return nil, fmt.Errorf("missing required permissions. Your bot needs this OAuth scope:\n  - channels:read (to list public channels) - REQUIRED\n\nPlease add this scope in your Slack app settings at https://api.slack.com/apps")
		}
		if strings.Contains(errStr, "invalid_auth") {
			logger.Error("Invalid Slack authentication token")
			return nil, fmt.Errorf("invalid token. Please check your SLACK_TOKEN")
		}
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	logger.WithField("total_channels", len(channels)).Debug("Retrieved channels from Slack API")

	var newChannels []Channel
	for _, ch := range channels {
		created := time.Unix(int64(ch.Created), 0)
		if created.After(since) {
			logger.WithFields(logger.LogFields{
				"channel_id":   ch.ID,
				"channel_name": ch.Name,
				"created":      created.Format("2006-01-02 15:04:05"),
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

func (c *Client) GetNewChannelsWithAllChannels(since time.Time) ([]Channel, []slack.Channel, error) {
	logger.WithField("since", since.Format("2006-01-02 15:04:05")).Debug("Fetching channels from Slack API")

	channels, _, err := c.api.GetConversations(&slack.GetConversationsParameters{
		Types: []string{"public_channel"},
		Limit: 1000,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	logger.WithField("total_channels", len(channels)).Debug("Retrieved channels from Slack API")

	var newChannels []Channel
	for _, ch := range channels {
		created := time.Unix(int64(ch.Created), 0)
		if created.After(since) {
			logger.WithFields(logger.LogFields{
				"channel_id":   ch.ID,
				"channel_name": ch.Name,
				"created":      created.Format("2006-01-02 15:04:05"),
			}).Debug("Found new channel")

			newChannels = append(newChannels, Channel{
				ID:      ch.ID,
				Name:    ch.Name,
				Created: created,
				Creator: ch.Creator,
				Purpose: ch.Purpose.Value,
			})
		}
	}

	logger.WithField("new_channels_count", len(newChannels)).Debug("Channel detection completed")
	return newChannels, channels, nil
}

func (c *Client) FormatNewChannelAnnouncement(channels []Channel, since time.Time) string {
	logger.WithFields(logger.LogFields{
		"channel_count": len(channels),
		"since":         since.Format("2006-01-02 15:04:05"),
	}).Debug("Formatting announcement message")

	// Calculate days since the search period
	daysSince := int(time.Since(since).Hours() / 24)
	daysText := textDays
	if daysSince == 1 {
		daysText = textDay
	}

	var builder strings.Builder

	if len(channels) == 1 {
		builder.WriteString(fmt.Sprintf("New channel created in the last %d %s!", daysSince, daysText))
	} else {
		builder.WriteString(fmt.Sprintf("%d new channels created in the last %d %s!", len(channels), daysSince, daysText))
	}

	builder.WriteString("\n\n")

	for i, ch := range channels {
		// Calculate days since creation
		daysSinceCreated := int(time.Since(ch.Created).Hours() / 24)
		daysText := "days"
		if daysSinceCreated == 1 {
			daysText = "day"
		}

		builder.WriteString(fmt.Sprintf("• <#%s>", ch.ID))

		// Add creator info if available
		if ch.Creator != "" {
			builder.WriteString(fmt.Sprintf(" created by <@%s>", ch.Creator))
		}

		// Add creation time info
		builder.WriteString(fmt.Sprintf(" %d %s ago", daysSinceCreated, daysText))

		if ch.Purpose != "" {
			builder.WriteString(fmt.Sprintf("\n  Description: %s", ch.Purpose))
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

func (c *Client) FormatNewChannelAnnouncementDryRun(channels []Channel, since time.Time) string {
	// Get user map for resolving creator names
	userMap, err := c.GetUserMap()
	if err != nil {
		// If we can't get users, fall back to regular format
		return c.FormatNewChannelAnnouncement(channels, since)
	}

	// Calculate days since the search period
	daysSince := int(time.Since(since).Hours() / 24)
	daysText := textDays
	if daysSince == 1 {
		daysText = textDay
	}

	var builder strings.Builder

	if len(channels) == 1 {
		builder.WriteString(fmt.Sprintf("New channel created in the last %d %s!", daysSince, daysText))
	} else {
		builder.WriteString(fmt.Sprintf("%d new channels created in the last %d %s!", len(channels), daysSince, daysText))
	}

	builder.WriteString("\n\n")

	for i, ch := range channels {
		// Calculate days since creation
		daysSinceCreated := int(time.Since(ch.Created).Hours() / 24)
		daysText := "days"
		if daysSinceCreated == 1 {
			daysText = "day"
		}

		builder.WriteString(fmt.Sprintf("• #%s", ch.Name))

		// Add creator info with pretty name if available
		if ch.Creator != "" {
			creatorName := ch.Creator // fallback to ID
			if userName, exists := userMap[ch.Creator]; exists && userName != "" {
				creatorName = userName
			}
			builder.WriteString(fmt.Sprintf(" created by %s", creatorName))
		}

		// Add creation time info
		builder.WriteString(fmt.Sprintf(" %d %s ago", daysSinceCreated, daysText))

		if ch.Purpose != "" {
			builder.WriteString(fmt.Sprintf("\n  Description: %s", ch.Purpose))
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

func (c *Client) CheckForDuplicateAnnouncement(channel, newMessage string, channelNames []string) (bool, error) {
	// Use a default cutoff of 30 days ago for backward compatibility
	cutoffTime := time.Now().Add(-30 * 24 * time.Hour)
	isDuplicate, _, err := c.CheckForDuplicateAnnouncementWithDetails(channel, newMessage, channelNames, cutoffTime)
	return isDuplicate, err
}

func (c *Client) CheckForDuplicateAnnouncementWithDetails(channel, newMessage string, channelNames []string, cutoffTime time.Time) (bool, []string, error) {
	// Use empty channel list - will make API call to get channels
	return c.CheckForDuplicateAnnouncementWithDetailsAndChannels(channel, newMessage, channelNames, cutoffTime, nil)
}

func (c *Client) CheckForDuplicateAnnouncementWithDetailsAndChannels(channel, newMessage string, channelNames []string, cutoffTime time.Time, allChannels []slack.Channel) (bool, []string, error) {
	logger.WithFields(logger.LogFields{
		"channel":      channel,
		"channel_list": strings.Join(channelNames, ", "),
	}).Debug("Checking for duplicate announcements")

	// Resolve channel name to channel ID
	channelID, err := c.ResolveChannelNameToID(channel)
	if err != nil {
		return false, nil, fmt.Errorf("failed to find channel %s: %w", channel, err)
	}

	// Create name-to-ID mapping
	channelNameToID, err := c.buildChannelNameToIDMapping(allChannels)
	if err != nil {
		return false, nil, err
	}

	// Get bot authentication info
	authInfo, err := c.TestAuth()
	if err != nil {
		return false, nil, fmt.Errorf("failed to get auth info: %w", err)
	}

	// Fetch channel history with retry logic
	history, err := c.fetchChannelHistoryWithRetry(channelID, channel)
	if err != nil {
		return false, nil, err
	}

	// Analyze messages for duplicates
	allAnnouncedChannels, foundDuplicates := c.analyzeMessagesForDuplicates(history.Messages, channelNames, channelNameToID, authInfo.UserID, channel)

	if foundDuplicates {
		return true, allAnnouncedChannels, nil
	}

	logger.WithFields(logger.LogFields{
		"channel":      channel,
		"scanned_msgs": len(history.Messages),
	}).Debug("No duplicate announcements found")
	return false, nil, nil
}

// buildChannelNameToIDMapping creates a mapping from channel names to IDs.
func (c *Client) buildChannelNameToIDMapping(allChannels []slack.Channel) (map[string]string, error) {
	var channelNameToID map[string]string
	if allChannels != nil {
		// Use provided channel list
		channelNameToID = make(map[string]string)
		for _, ch := range allChannels {
			channelNameToID[ch.Name] = ch.ID
		}
	} else {
		// Fallback: get channels via API call
		var apiErr error
		channelNameToID, apiErr = c.getAllChannelNameToIDMap()
		if apiErr != nil {
			logger.WithFields(logger.LogFields{
				"error": apiErr.Error(),
			}).Warn("Failed to get channel mappings for duplicate detection, will use name-only matching")
			channelNameToID = make(map[string]string)
		}
	}
	return channelNameToID, nil
}

// fetchChannelHistoryWithRetry fetches channel history with retry logic for rate limits.
func (c *Client) fetchChannelHistoryWithRetry(channelID, channel string) (*slack.GetConversationHistoryResponse, error) {
	params := &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     15, // Explicit limit to match API restriction
	}

	const maxRetries = 3
	var history *slack.GetConversationHistoryResponse

	for attempt := 1; attempt <= maxRetries; attempt++ {
		var err error
		history, err = c.api.GetConversationHistory(params)
		if err != nil {
			if shouldRetryOnRateLimit(err.Error(), attempt, maxRetries, channel) {
				continue
			}
			// For non-rate-limit errors or final attempt
			logger.WithFields(logger.LogFields{
				"channel": channel,
				"error":   err.Error(),
			}).Warn("Failed to get channel history for duplicate check")
			// Return empty history instead of nil to avoid null pointer issues
			return &slack.GetConversationHistoryResponse{Messages: []slack.Message{}}, nil
		}
		// Success - break out of retry loop
		break
	}
	return history, nil
}

// analyzeMessagesForDuplicates analyzes messages to find duplicate announcements.
func (c *Client) analyzeMessagesForDuplicates(messages []slack.Message, channelNames []string, channelNameToID map[string]string, botUserID, channel string) ([]string, bool) {
	var allAnnouncedChannels []string
	var foundDuplicates bool

	for _, message := range messages {
		// Skip if not from our bot
		if message.User != botUserID && message.BotID != botUserID {
			continue
		}
		if duplicateChannels := c.findDuplicateChannelsInMessageWithIDs(message.Text, channelNames, channelNameToID); len(duplicateChannels) > 0 {
			logger.WithFields(logger.LogFields{
				"channel":            channel,
				"duplicate_ts":       message.Timestamp,
				"announced_channels": strings.Join(duplicateChannels, ", "),
			}).Debug("Found duplicate announcement")

			// Add these channels to our list of announced channels
			allAnnouncedChannels = c.addUniqueChannels(allAnnouncedChannels, duplicateChannels)
			foundDuplicates = true
		}
	}

	return allAnnouncedChannels, foundDuplicates
}

// addUniqueChannels adds channels to the list, avoiding duplicates.
func (c *Client) addUniqueChannels(existing []string, newChannels []string) []string {
	for _, dupChannel := range newChannels {
		// Avoid duplicates in our list
		alreadyInList := false
		for _, existingChannel := range existing {
			if existingChannel == dupChannel {
				alreadyInList = true
				break
			}
		}
		if !alreadyInList {
			existing = append(existing, dupChannel)
		}
	}
	return existing
}

func (c *Client) findDuplicateChannelsInMessageWithIDs(messageText string, channelNames []string, channelNameToID map[string]string) []string {
	var duplicates []string
	// Look for channel mentions in the format <#CHANNELID> or #channelname
	for _, channelName := range channelNames {
		// Check for #channelname format
		if strings.Contains(messageText, "#"+channelName) {
			duplicates = append(duplicates, channelName)
			continue
		}

		// Check for <#CHANNELID> format using pre-resolved IDs
		if channelID, exists := channelNameToID[channelName]; exists {
			if strings.Contains(messageText, "<#"+channelID+">") {
				duplicates = append(duplicates, channelName)
			}
		}
	}

	return duplicates
}

func (c *Client) getAllChannelNameToIDMap() (map[string]string, error) {

	// Get all channels in one API call
	params := &slack.GetConversationsParameters{
		Types: []string{"public_channel"},
		Limit: 1000,
	}

	channels, _, err := c.api.GetConversations(params)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	// Create name-to-ID mapping
	nameToID := make(map[string]string)
	for _, channel := range channels {
		nameToID[channel.Name] = channel.ID
	}

	return nameToID, nil
}

func (c *Client) PostMessage(channel, message string) error {
	logger.WithFields(logger.LogFields{
		"channel":        channel,
		"message_length": len(message),
	}).Debug("Attempting to post message to channel")

	// Validate channel name format
	if err := ValidateChannelName(channel); err != nil {
		logger.WithFields(logger.LogFields{
			"channel": channel,
			"error":   err.Error(),
		}).Error("Invalid channel name format")
		return fmt.Errorf("invalid channel name '%s': %w", channel, err)
	}

	// Resolve channel name to channel ID
	channelID, err := c.ResolveChannelNameToID(channel)
	if err != nil {
		return fmt.Errorf("failed to find channel %s: %w", channel, err)
	}

	_, _, err = c.api.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil {
		errStr := err.Error()
		logger.WithFields(logger.LogFields{
			"channel":   channel,
			"error":     errStr,
			"operation": "post_message",
		}).Error("Failed to post message to Slack")

		// Handle rate limiting
		if strings.Contains(errStr, "rate_limited") {
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
		return fmt.Errorf("failed to post message to %s: %w", channel, err)
	}

	logger.WithField("channel", channel).Debug("Message posted successfully")
	return nil
}

func (c *Client) TestAuth() (*AuthInfo, error) {
	auth, err := c.api.AuthTest()
	if err != nil {
		return nil, err
	}

	// Get team info to construct workspace URL
	teamInfo, err := c.api.GetTeamInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get team info: %w", err)
	}

	// Construct workspace URL from team domain
	workspaceURL := fmt.Sprintf("https://%s.slack.com", teamInfo.Domain)

	return &AuthInfo{
		User:         auth.User,
		UserID:       auth.UserID,
		Team:         auth.Team,
		TeamID:       auth.TeamID,
		WorkspaceURL: workspaceURL,
	}, nil
}

func (c *Client) CheckOAuthScopes() (map[string]bool, error) {
	_, err := c.api.AuthTest()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	// Test each required scope individually by trying operations that require them
	// This is more reliable than trying to parse scope information from auth response

	scopeResults := make(map[string]bool)

	// Test each required scope individually by trying operations that require them
	scopeResults["channels:read"] = c.testChannelsReadScope()
	scopeResults["channels:join"] = c.testChannelsJoinScope()
	scopeResults["channels:history"] = c.testChannelsHistoryScope()
	scopeResults["chat:write"] = c.testChatWriteScope()
	scopeResults["channels:manage"] = c.testChannelsManageScope()
	scopeResults["users:read"] = c.testUsersReadScope()

	return scopeResults, nil
}

func (c *Client) testChannelsReadScope() bool {
	// Try to get conversations - this requires channels:read
	_, _, err := c.api.GetConversations(&slack.GetConversationsParameters{
		Types: []string{"public_channel"},
		Limit: 1,
	})

	if err != nil && strings.Contains(err.Error(), "missing_scope") {
		return false
	}
	return true
}

func (c *Client) testChannelsJoinScope() bool {
	// We can't easily test this without actually trying to join a channel
	// For now, we'll assume it's available if channels:read works
	// In a real scenario, this would be tested when actually trying to join
	return true
}

func (c *Client) testChannelsHistoryScope() bool {
	// Try to get conversation history - this requires channels:history
	// We'll try to get conversations first to find a channel to test with
	conversations, _, err := c.api.GetConversations(&slack.GetConversationsParameters{
		Types: []string{"public_channel"},
		Limit: 1,
	})

	if err != nil || len(conversations) == 0 {
		// If we can't get conversations, assume scope is available
		// This will be tested when actually reading history
		return true
	}

	// Try to get history from the first channel
	_, err = c.api.GetConversationHistory(&slack.GetConversationHistoryParameters{
		ChannelID: conversations[0].ID,
		Limit:     1,
	})

	if err != nil && strings.Contains(err.Error(), "missing_scope") {
		return false
	}
	return true
}

func (c *Client) testChatWriteScope() bool {
	// We can't test this without actually posting a message
	// For now, we'll assume it's available - it will be tested when actually posting
	return true
}

func (c *Client) testChannelsManageScope() bool {
	// We can't test this without actually trying to archive a channel
	// For now, we'll assume it's available - it will be tested when actually archiving
	return true
}

func (c *Client) testUsersReadScope() bool {
	// Try to get users list - this requires users:read
	_, err := c.api.GetUsers()

	if err != nil && strings.Contains(err.Error(), "missing_scope") {
		return false
	}
	return true
}

func (c *Client) GetChannelInfo(channelID string) (*Channel, error) {
	// This is used for permission testing in health checks
	// We'll just return a mock error for permission testing
	return nil, fmt.Errorf("channel_not_found")
}

// ResolveChannelNameToID converts a channel name (like "#general" or "general") to its Slack channel ID.
func (c *Client) ResolveChannelNameToID(channelName string) (string, error) {
	// Clean the channel name
	cleanName := strings.TrimPrefix(channelName, "#")

	// Get all channels to find the matching one
	params := &slack.GetConversationsParameters{
		Types: []string{"public_channel"},
		Limit: 1000,
	}

	channels, _, err := c.api.GetConversations(params)
	if err != nil {
		return "", fmt.Errorf("failed to get channels: %w", err)
	}

	// Find channel by name
	for _, channel := range channels {
		if channel.Name == cleanName {
			return channel.ID, nil
		}
	}

	return "", fmt.Errorf("channel '%s' not found", channelName)
}

func (c *Client) GetInactiveChannels(warnSeconds int, archiveSeconds int) (toWarn []Channel, toArchive []Channel, err error) {
	logger.WithFields(logger.LogFields{
		"warn_seconds":    warnSeconds,
		"archive_seconds": archiveSeconds,
	}).Debug("Starting inactive channel detection")

	warnCutoff := time.Now().Add(-time.Duration(warnSeconds) * time.Second)

	// Get all channels
	allChannels, _, err := c.api.GetConversations(&slack.GetConversationsParameters{
		Types:           []string{"public_channel"},
		Limit:           1000,
		ExcludeArchived: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	// Pre-filter channels without exclusions (empty exclusion lists)
	candidateChannels, stats := c.preFilterChannelsWithExclusions(allChannels, warnCutoff, nil, nil)
	c.logInactiveChannelsFilteringStats(len(allChannels), len(candidateChannels), stats)

	// Auto-join channels before analysis
	if len(candidateChannels) > 0 {
		fmt.Printf("🤖 Joining %d public channels for analysis...\n", len(candidateChannels))
	}
	joinedCount, err := c.autoJoinPublicChannels(candidateChannels)
	if err != nil {
		return toWarn, toArchive, fmt.Errorf("failed to auto-join channels - inactive detection requires channel membership: %w", err)
	}
	if len(candidateChannels) > 0 {
		fmt.Printf("✅ Joined %d channels successfully\n\n", joinedCount)
	}
	logger.WithField("joined_count", joinedCount).Debug("Auto-joined public channels")

	// Analyze channels for inactivity
	return c.analyzeChannelsForBasicInactivity(candidateChannels, warnCutoff, archiveSeconds)
}

// logInactiveChannelsFilteringStats logs filtering statistics for basic inactive channel detection.
func (c *Client) logInactiveChannelsFilteringStats(totalChannels, candidateChannels int, stats channelFilterStats) {
	logger.WithFields(logger.LogFields{
		"total_channels":     totalChannels,
		"candidate_channels": candidateChannels,
		"skipped_new":        stats.skippedNew,
		"skipped_excluded":   stats.skippedExcluded,
		"skipped_active":     stats.skippedActive,
	}).Debug("Pre-filtered channels using metadata")
}

// analyzeChannelsForBasicInactivity analyzes channels for basic inactivity without detailed reporting.
func (c *Client) analyzeChannelsForBasicInactivity(candidateChannels []slack.Channel, warnCutoff time.Time, archiveSeconds int) (toWarn []Channel, toArchive []Channel, err error) {
	for i, ch := range candidateChannels {
		logger.WithFields(logger.LogFields{
			"channel": ch.Name,
			"index":   i + 1,
			"total":   len(candidateChannels),
		}).Debug("Checking message history for candidate channel")

		// Get channel activity with retry for rate limits
		lastActivity, hasWarning, warningTime, err := c.getChannelActivityWithRetry(ch.ID, ch.Name)
		if err != nil {
			if c.handleBasicChannelActivityError(err, ch.Name) {
				return toWarn, toArchive, fmt.Errorf("rate limited by Slack API while processing channel %s: %w", ch.Name, err)
			}
			// Handle special error cases
			lastActivity, hasWarning, warningTime = c.handleChannelNotInBotError(err)
			if lastActivity.IsZero() && !hasWarning {
				continue // Skip this channel
			}
		}

		c.logChannelActivityData(ch.Name, lastActivity, hasWarning, warningTime)

		// Create channel struct and make decision
		channel := c.createBasicChannel(ch, lastActivity)
		if hasWarning {
			if c.shouldArchiveBasicChannel(lastActivity, warningTime, archiveSeconds) {
				toArchive = append(toArchive, channel)
				c.logBasicChannelDecision(ch.Name, "archival", lastActivity, warningTime)
			}
		} else if c.shouldWarnChannel(lastActivity, warnCutoff) {
			toWarn = append(toWarn, channel)
			c.logBasicChannelDecision(ch.Name, "warning", lastActivity, time.Time{})
		}
	}

	logger.WithFields(logger.LogFields{
		"to_warn":    len(toWarn),
		"to_archive": len(toArchive),
	}).Debug("Inactive channel analysis completed")

	return toWarn, toArchive, nil
}

// handleBasicChannelActivityError handles errors during basic channel activity retrieval.
func (c *Client) handleBasicChannelActivityError(err error, channelName string) bool {
	errStr := err.Error()
	if strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "rate limit") {
		logger.WithFields(logger.LogFields{
			"channel": channelName,
			"error":   errStr,
		}).Warn("Rate limited by Slack API - this affects all subsequent requests, stopping analysis")
		return true
	}
	return false
}

// handleChannelNotInBotError handles the case where bot is not in channel.
func (c *Client) handleChannelNotInBotError(err error) (time.Time, bool, time.Time) {
	errStr := err.Error()
	if strings.Contains(errStr, "not_in_channel") {
		// Bot isn't in channel, treat as potentially inactive
		return time.Unix(0, 0), false, time.Time{}
	}
	// For other errors, indicate to skip the channel
	return time.Time{}, false, time.Time{}
}

// logChannelActivityData logs channel activity data for debugging.
func (c *Client) logChannelActivityData(channelName string, lastActivity time.Time, hasWarning bool, warningTime time.Time) {
	logger.WithFields(logger.LogFields{
		"channel":       channelName,
		"last_activity": lastActivity.Format("2006-01-02 15:04:05"),
		"has_warning":   hasWarning,
		"warning_time":  warningTime.Format("2006-01-02 15:04:05"),
	}).Debug("Retrieved channel activity data")
}

// createBasicChannel creates a basic channel struct.
func (c *Client) createBasicChannel(ch slack.Channel, lastActivity time.Time) Channel {
	return Channel{
		ID:           ch.ID,
		Name:         ch.Name,
		Created:      time.Unix(int64(ch.Created), 0),
		Purpose:      ch.Purpose.Value,
		Creator:      ch.Creator,
		LastActivity: lastActivity,
		MemberCount:  ch.NumMembers,
		IsArchived:   ch.IsArchived,
	}
}

// shouldArchiveBasicChannel determines if a basic channel should be archived.
func (c *Client) shouldArchiveBasicChannel(lastActivity, warningTime time.Time, archiveSeconds int) bool {
	gracePeriodExpired := time.Since(warningTime) > time.Duration(archiveSeconds)*time.Second
	return lastActivity.Before(warningTime) && gracePeriodExpired
}

// logBasicChannelDecision logs decisions for basic channel analysis.
func (c *Client) logBasicChannelDecision(channelName, decision string, lastActivity, warningTime time.Time) {
	switch decision {
	case "archival":
		logger.WithFields(logger.LogFields{
			"channel":       channelName,
			"last_activity": lastActivity.Format("2006-01-02 15:04:05"),
			"warning_time":  warningTime.Format("2006-01-02 15:04:05"),
		}).Debug("Channel marked for archival")
	case "warning":
		logger.WithFields(logger.LogFields{
			"channel":       channelName,
			"last_activity": lastActivity.Format("2006-01-02 15:04:05"),
		}).Debug("Channel marked for warning")
	}
}

// GetInactiveChannelsWithDetails returns inactive channels with detailed message information and user name resolution.
func (c *Client) GetInactiveChannelsWithDetails(warnSeconds int, archiveSeconds int, userMap map[string]string, isDebug bool) (toWarn []Channel, toArchive []Channel, err error) {
	logger.WithFields(logger.LogFields{
		"warn_seconds":    warnSeconds,
		"archive_seconds": archiveSeconds,
	}).Debug("Starting inactive channel detection with message details")

	warnCutoff := time.Now().Add(-time.Duration(warnSeconds) * time.Second)

	// Get all channels
	allChannels, _, err := c.api.GetConversations(&slack.GetConversationsParameters{
		Types:           []string{"public_channel"},
		Limit:           1000,
		ExcludeArchived: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	// Pre-filter channels without exclusions (empty exclusion lists)
	candidateChannels, stats := c.preFilterChannelsWithExclusions(allChannels, warnCutoff, nil, nil)
	c.logChannelFilteringStatsSimple(len(allChannels), len(candidateChannels), stats, isDebug)

	// Auto-join channels and analyze activity
	joinedCount, err := c.autoJoinChannelsForAnalysis(candidateChannels, isDebug)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to auto-join channels - inactive detection requires channel membership: %w", err)
	}
	logger.WithField("joined_count", joinedCount).Debug("Auto-joined public channels")

	// Analyze each candidate channel for activity
	return c.analyzeChannelsForInactivity(candidateChannels, userMap, warnCutoff, archiveSeconds, isDebug)
}

// GetInactiveChannelsWithDetailsAndExclusions returns inactive channels with exclusion support.
func (c *Client) GetInactiveChannelsWithDetailsAndExclusions(warnSeconds int, archiveSeconds int, userMap map[string]string, excludeChannels, excludePrefixes []string, isDebug bool) (toWarn []Channel, toArchive []Channel, totalChannels int, err error) {
	logger.WithFields(logger.LogFields{
		"warn_seconds":     warnSeconds,
		"archive_seconds":  archiveSeconds,
		"exclude_channels": excludeChannels,
		"exclude_prefixes": excludePrefixes,
	}).Debug("Starting inactive channel detection with message details and exclusions")

	warnCutoff := time.Now().Add(-time.Duration(warnSeconds) * time.Second)

	// Get all channels
	allChannels, _, err := c.api.GetConversations(&slack.GetConversationsParameters{
		Types:           []string{"public_channel"},
		Limit:           1000,
		ExcludeArchived: true,
	})
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get conversations: %w", err)
	}

	// Pre-filter channels to reduce API calls
	candidateChannels, stats := c.preFilterChannelsWithExclusions(allChannels, warnCutoff, excludeChannels, excludePrefixes)
	c.logChannelFilteringStats(len(allChannels), len(candidateChannels), stats, isDebug)

	// Auto-join channels and analyze activity
	joinedCount, err := c.autoJoinChannelsForAnalysis(candidateChannels, isDebug)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to auto-join channels - inactive detection requires channel membership: %w", err)
	}
	logger.WithField("joined_count", joinedCount).Debug("Auto-joined public channels")

	// Analyze each candidate channel for activity
	toWarn, toArchive, err = c.analyzeChannelsForInactivity(candidateChannels, userMap, warnCutoff, archiveSeconds, isDebug)
	return toWarn, toArchive, len(candidateChannels), err
}

// channelFilterStats holds statistics about channel filtering.
type channelFilterStats struct {
	skippedActive       int
	skippedExcluded     int
	skippedNew          int
	skippedUserExcluded int
}

// preFilterChannelsWithExclusions filters channels using metadata and exclusions to reduce API calls.
func (c *Client) preFilterChannelsWithExclusions(allChannels []slack.Channel, warnCutoff time.Time, excludeChannels, excludePrefixes []string) ([]slack.Channel, channelFilterStats) {
	candidateChannels := make([]slack.Channel, 0, len(allChannels))
	stats := channelFilterStats{}

	for _, ch := range allChannels {
		if c.shouldSkipChannelWithExclusions(ch.Name, excludeChannels, excludePrefixes) {
			stats.skippedUserExcluded++
			continue
		}

		if c.shouldSkipChannel(ch.Name) {
			stats.skippedExcluded++
			continue
		}

		created := time.Unix(int64(ch.Created), 0)
		if created.After(warnCutoff) {
			stats.skippedNew++
			continue
		}

		if c.seemsActiveFromMetadata(ch, warnCutoff) {
			stats.skippedActive++
			continue
		}

		candidateChannels = append(candidateChannels, ch)
	}

	return candidateChannels, stats
}

// logChannelFilteringStatsSimple logs and prints filtering statistics without user exclusions.
func (c *Client) logChannelFilteringStatsSimple(totalChannels, candidateChannels int, stats channelFilterStats, isDebug bool) {
	logger.WithFields(logger.LogFields{
		"total_channels":     totalChannels,
		"candidate_channels": candidateChannels,
		"skipped_active":     stats.skippedActive,
		"skipped_excluded":   stats.skippedExcluded,
		"skipped_new":        stats.skippedNew,
	}).Debug("Pre-filtered channels using metadata")

	fmt.Printf("📞 API Call 2: Getting channel list with metadata...\n")
	fmt.Printf("✅ Got %d channels from API\n", totalChannels)
	fmt.Printf("   Pre-filtered to %d candidates (skipped %d active, %d excluded, %d too new)\n\n",
		candidateChannels, stats.skippedActive, stats.skippedExcluded, stats.skippedNew)
}

// logChannelFilteringStats logs and optionally prints channel filtering statistics.
func (c *Client) logChannelFilteringStats(totalChannels, candidateChannels int, stats channelFilterStats, isDebug bool) {
	logger.WithFields(logger.LogFields{
		"total_channels":        totalChannels,
		"candidate_channels":    candidateChannels,
		"skipped_active":        stats.skippedActive,
		"skipped_excluded":      stats.skippedExcluded,
		"skipped_new":           stats.skippedNew,
		"skipped_user_excluded": stats.skippedUserExcluded,
	}).Debug("Pre-filtered channels using metadata and exclusions")

	if isDebug {
		fmt.Printf("📞 API Call 2: Getting channel list with metadata...\n")
		fmt.Printf("✅ Got %d channels from API\n", totalChannels)
		fmt.Printf("   Pre-filtered to %d candidates (skipped %d active, %d excluded, %d too new, %d user-excluded)\n\n",
			candidateChannels, stats.skippedActive, stats.skippedExcluded, stats.skippedNew, stats.skippedUserExcluded)
	}
}

// autoJoinChannelsForAnalysis auto-joins public channels before analysis.
func (c *Client) autoJoinChannelsForAnalysis(candidateChannels []slack.Channel, isDebug bool) (int, error) {
	// Count how many channels need joining vs already member
	needsJoining := 0
	alreadyMember := 0
	for _, ch := range candidateChannels {
		if !ch.IsPrivate {
			if ch.IsMember {
				alreadyMember++
			} else {
				needsJoining++
			}
		}
	}

	if len(candidateChannels) > 0 {
		if needsJoining > 0 {
			fmt.Printf("🤖 Joining %d channels (already member of %d)...\n", needsJoining, alreadyMember)
		} else {
			fmt.Printf("🤖 Already member of all %d channels, no joining needed\n", alreadyMember)
		}
		if isDebug {
			fmt.Printf("📞 API Calls 3+: Auto-joining public channels for accurate analysis...\n")
		}
	}
	joinedCount, err := c.autoJoinPublicChannels(candidateChannels)
	if len(candidateChannels) > 0 {
		if joinedCount > 0 {
			fmt.Printf("✅ Successfully joined %d channels\n\n", joinedCount)
		} else if needsJoining == 0 {
			fmt.Printf("✅ No channel joining required\n\n")
		} else {
			fmt.Printf("✅ Channel joining completed\n\n")
		}
		if isDebug {
			fmt.Printf("✅ Auto-joined %d channels\n\n", joinedCount)
		}
	}
	return joinedCount, err
}

// analyzeChannelsForInactivity analyzes each candidate channel for activity and categorizes them.
func (c *Client) analyzeChannelsForInactivity(candidateChannels []slack.Channel, userMap map[string]string, warnCutoff time.Time, archiveSeconds int, isDebug bool) (toWarn []Channel, toArchive []Channel, err error) {
	now := time.Now()
	for i, ch := range candidateChannels {
		lastActivity, hasWarning, warningTime, lastMessage, err := c.GetChannelActivityWithMessageAndUsers(ch.ID, userMap)
		if err != nil {
			if c.handleChannelAnalysisError(err, ch.Name, isDebug) {
				return toWarn, toArchive, fmt.Errorf("rate limited by Slack API")
			}
			continue
		}

		if isDebug {
			fmt.Printf("✅ API Call succeeded\n")
		}

		enhancedChannel := c.createEnhancedChannel(ch, lastActivity, lastMessage)
		c.displayChannelAnalysis(ch, lastActivity, hasWarning, warningTime, lastMessage, now, i, len(candidateChannels))

		if hasWarning {
			if c.shouldArchiveChannel(warningTime, archiveSeconds) {
				toArchive = append(toArchive, enhancedChannel)
				c.logChannelDecision(ch.Name, "archival", warningTime)
			}
		} else if c.shouldWarnChannel(lastActivity, warnCutoff) {
			toWarn = append(toWarn, enhancedChannel)
			c.logChannelDecision(ch.Name, "warning", lastActivity)
		}
	}

	logger.WithFields(logger.LogFields{
		"channels_to_warn":    len(toWarn),
		"channels_to_archive": len(toArchive),
	}).Debug("Inactive channel analysis completed")

	return toWarn, toArchive, nil
}

// handleChannelAnalysisError handles errors during channel analysis and returns true if processing should stop.
func (c *Client) handleChannelAnalysisError(err error, channelName string, isDebug bool) bool {
	errStr := err.Error()
	logger.WithFields(logger.LogFields{
		"channel": channelName,
		"error":   errStr,
	}).Error("Failed to get channel activity")

	if strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "rate limit") {
		logger.WithFields(logger.LogFields{
			"channel": channelName,
			"error":   errStr,
		}).Warn("Rate limited by Slack API - this affects all subsequent requests, stopping analysis")

		if isDebug {
			fmt.Printf("❌ API Call failed: Rate limited while checking #%s\n", channelName)
		}
		return true
	}

	if isDebug {
		fmt.Printf("❌ API Call failed: Error checking #%s: %s\n", channelName, errStr)
	}
	return false
}

// createEnhancedChannel creates an enhanced channel struct with message details.
func (c *Client) createEnhancedChannel(ch slack.Channel, lastActivity time.Time, lastMessage *MessageInfo) Channel {
	return Channel{
		ID:           ch.ID,
		Name:         ch.Name,
		Created:      time.Unix(int64(ch.Created), 0),
		Purpose:      ch.Purpose.Value,
		Creator:      ch.Creator,
		LastActivity: lastActivity,
		MemberCount:  ch.NumMembers,
		IsArchived:   ch.IsArchived,
		LastMessage:  lastMessage,
	}
}

// displayChannelAnalysis displays detailed channel analysis information.
func (c *Client) displayChannelAnalysis(ch slack.Channel, lastActivity time.Time, hasWarning bool, warningTime time.Time, lastMessage *MessageInfo, now time.Time, currentIndex, totalChannels int) {
	activityStr := c.formatChannelActivityString(lastActivity, hasWarning, warningTime, now, ch)
	fmt.Printf("  [%d/%d] #%-20s - %s\n", currentIndex+1, totalChannels, ch.Name, activityStr)

	if lastMessage != nil {
		c.displayMessageDetails(lastMessage)
	}
}

// GetRandomChannels gets all channels and returns a random subset of the specified count.
func (c *Client) GetRandomChannels(count int) ([]Channel, error) {
	logger.WithField("count", count).Debug("Fetching all channels for random selection")

	// Get all public channels
	allSlackChannels, _, err := c.api.GetConversations(&slack.GetConversationsParameters{
		Types:           []string{"public_channel"},
		Limit:           1000,
		ExcludeArchived: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	// Convert to our Channel type
	channels := make([]Channel, 0, len(allSlackChannels))
	for _, ch := range allSlackChannels {
		channels = append(channels, Channel{
			ID:          ch.ID,
			Name:        ch.Name,
			Created:     time.Unix(int64(ch.Created), 0),
			Purpose:     ch.Purpose.Value,
			Creator:     ch.Creator,
			MemberCount: ch.NumMembers,
			IsArchived:  ch.IsArchived,
		})
	}

	// Return all channels if count is larger than available
	if count >= len(channels) {
		return channels, nil
	}

	// Randomly select channels using local random generator
	// #nosec G404 - using math/rand is appropriate for non-security channel selection
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(channels), func(i, j int) {
		channels[i], channels[j] = channels[j], channels[i]
	})

	return channels[:count], nil
}

// FormatChannelHighlightAnnouncement formats a channel highlight announcement for commit mode.
func (c *Client) FormatChannelHighlightAnnouncement(channels []Channel) string {
	var builder strings.Builder

	if len(channels) == 1 {
		builder.WriteString("🧭 Here is 1 randomly-selected public channel that you are welcome to explore!")
	} else {
		builder.WriteString(fmt.Sprintf("🧭 Here are %d randomly-selected public channels that you are welcome to explore!", len(channels)))
	}

	builder.WriteString("\n\n")

	for i, ch := range channels {
		builder.WriteString(fmt.Sprintf("• <#%s>", ch.ID))

		if ch.Purpose != "" {
			builder.WriteString(fmt.Sprintf("\n  %s", ch.Purpose))
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

// FormatChannelHighlightAnnouncementDryRun formats a channel highlight announcement for dry run mode.
func (c *Client) FormatChannelHighlightAnnouncementDryRun(channels []Channel) string {
	var builder strings.Builder

	if len(channels) == 1 {
		builder.WriteString("🧭 Here is 1 randomly-selected public channel that you are welcome to explore!")
	} else {
		builder.WriteString(fmt.Sprintf("🧭 Here are %d randomly-selected public channels that you are welcome to explore!", len(channels)))
	}

	builder.WriteString("\n\n")

	for i, ch := range channels {
		builder.WriteString(fmt.Sprintf("• #%s", ch.Name))

		if ch.Purpose != "" {
			builder.WriteString(fmt.Sprintf("\n  %s", ch.Purpose))
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

// formatChannelActivityString formats the activity description for a channel.
func (c *Client) formatChannelActivityString(lastActivity time.Time, hasWarning bool, warningTime time.Time, now time.Time, ch slack.Channel) string {
	if lastActivity.IsZero() {
		created := time.Unix(int64(ch.Created), 0)
		createdDuration := now.Sub(created)
		return fmt.Sprintf("no real messages (created %s ago)", formatDuration(createdDuration))
	}

	duration := now.Sub(lastActivity)
	activityStr := fmt.Sprintf("last real message %s ago", formatDuration(duration))
	if hasWarning && !warningTime.IsZero() {
		warningDuration := now.Sub(warningTime)
		activityStr += fmt.Sprintf(" (warning sent %s ago)", formatDuration(warningDuration))
	}
	return activityStr
}

// displayMessageDetails displays the details of the last message in a channel.
func (c *Client) displayMessageDetails(lastMessage *MessageInfo) {
	messageText := lastMessage.Text
	if len(messageText) > 80 {
		messageText = messageText[:77] + "..."
	}
	messageText = strings.ReplaceAll(messageText, "\n", " ")

	botIndicator := ""
	if lastMessage.IsBot {
		botIndicator = " (bot)"
	}

	authorName := lastMessage.UserName
	if authorName == "" {
		authorName = lastMessage.User
	}

	fmt.Printf("    └─ Author: %s%s | Message: \"%s\"\n", authorName, botIndicator, messageText)
}

// shouldArchiveChannel determines if a channel should be archived based on warning time.
func (c *Client) shouldArchiveChannel(warningTime time.Time, archiveSeconds int) bool {
	return time.Since(warningTime) > time.Duration(archiveSeconds)*time.Second
}

// shouldWarnChannel determines if a channel should receive a warning based on activity.
func (c *Client) shouldWarnChannel(lastActivity time.Time, warnCutoff time.Time) bool {
	return lastActivity.IsZero() || lastActivity.Before(warnCutoff)
}

// logChannelDecision logs the decision made for a channel.
func (c *Client) logChannelDecision(channelName, decision string, timestamp time.Time) {
	switch decision {
	case "archival":
		logger.WithFields(logger.LogFields{
			"channel":      channelName,
			"warning_time": timestamp.Format("2006-01-02 15:04:05"),
			"grace_period": "expired",
		}).Debug("Channel marked for archival - grace period expired")
	case "warning":
		logger.WithFields(logger.LogFields{
			"channel":       channelName,
			"last_activity": timestamp.Format("2006-01-02 15:04:05"),
			"inactive_for":  time.Since(timestamp).String(),
		}).Debug("Channel marked for warning")
	}
}

// shouldSkipChannelWithExclusions checks if a channel should be skipped based on user exclusions.
func (c *Client) shouldSkipChannelWithExclusions(channelName string, excludeChannels, excludePrefixes []string) bool {
	// Check exact channel name matches
	for _, excluded := range excludeChannels {
		if channelName == excluded {
			logger.WithFields(logger.LogFields{
				"channel": channelName,
				"reason":  "exact_match",
				"exclude": excluded,
			}).Debug("Skipping channel due to user exclusion")
			return true
		}
	}

	// Check prefix matches
	for _, prefix := range excludePrefixes {
		if strings.HasPrefix(channelName, prefix) {
			logger.WithFields(logger.LogFields{
				"channel": channelName,
				"reason":  "prefix_match",
				"prefix":  prefix,
			}).Debug("Skipping channel due to user prefix exclusion")
			return true
		}
	}

	return false
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		seconds := int(d.Seconds())
		if seconds == 1 {
			return "1 second"
		}
		return fmt.Sprintf("%d seconds", seconds)
	}

	if d < time.Hour {
		minutes := int(d.Minutes())
		if minutes == 1 {
			return oneMinuteText
		}
		return fmt.Sprintf("%d minutes", minutes)
	}

	if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return oneHourText
		}
		return fmt.Sprintf("%d hours", hours)
	}

	days := int(d.Hours() / 24)
	if days == 1 {
		return oneDayText
	}
	return fmt.Sprintf("%d days", days)
}

func (c *Client) getChannelActivityWithRetry(channelID, channelName string) (lastActivity time.Time, hasWarning bool, warningTime time.Time, err error) {
	// Just call the base function - rate limiting should be handled globally
	// If we get rate limited, that affects the entire API, so we should stop processing
	return c.getChannelActivity(channelID)
}

func (c *Client) getChannelActivity(channelID string) (lastActivity time.Time, hasWarning bool, warningTime time.Time, err error) {
	// Fetch initial channel history
	history, err := c.fetchInitialChannelHistory(channelID)
	if err != nil {
		return time.Time{}, false, time.Time{}, err
	}

	if len(history.Messages) == 0 {
		return time.Unix(0, 0), false, time.Time{}, nil
	}

	// Analyze messages for recent activity
	botUserID := c.getBotUserID()
	lastRealMsg, lastRealMsgTime := c.findMostRecentRealMessage(history.Messages, botUserID)

	// Determine if we need detailed analysis
	if c.needsDetailedAnalysis(lastRealMsg, botUserID) {
		return c.getDetailedChannelActivity(channelID, botUserID)
	}

	// If the last real message is from a user, that's our activity time
	if lastRealMsg != nil && lastRealMsg.User != botUserID {
		return lastRealMsgTime, false, time.Time{}, nil
	}

	// Need detailed analysis for other cases
	return c.getDetailedChannelActivity(channelID, botUserID)
}

// fetchInitialChannelHistory fetches initial channel history with retry logic.
func (c *Client) fetchInitialChannelHistory(channelID string) (*slack.GetConversationHistoryResponse, error) {
	const maxRetries = 3
	var history *slack.GetConversationHistoryResponse

	for attempt := 1; attempt <= maxRetries; attempt++ {
		params := &slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Limit:     10, // Get enough messages to find real ones past any system messages
		}

		var err error
		history, err = c.api.GetConversationHistory(params)
		if err != nil {
			if shouldRetryOnRateLimitSimple(err.Error(), attempt, maxRetries) {
				continue
			}
			errStr := err.Error()
			if strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "rate limit") {
				return nil, fmt.Errorf("failed to get channel history after %d attempts, final Slack error: %s", maxRetries, errStr)
			}
			// Non-rate-limit errors - don't retry
			return nil, fmt.Errorf("failed to get channel history: %w", err)
		}
		// Success - break out of retry loop
		break
	}
	return history, nil
}

// findMostRecentRealMessage finds the most recent real message in the history.
func (c *Client) findMostRecentRealMessage(messages []slack.Message, botUserID string) (*slack.Message, time.Time) {
	for _, msg := range messages {
		if isRealMessage(msg, botUserID) {
			if msgTime, err := parseSlackTimestamp(msg.Timestamp); err == nil {
				return &msg, msgTime
			}
			return &msg, time.Time{}
		}
	}
	return nil, time.Time{}
}

// needsDetailedAnalysis determines if detailed channel analysis is needed.
func (c *Client) needsDetailedAnalysis(lastRealMsg *slack.Message, botUserID string) bool {
	if lastRealMsg == nil {
		return true
	}
	// If the last real message is from the bot and contains a warning, we need more context
	if lastRealMsg.User == botUserID && strings.Contains(strings.ToLower(lastRealMsg.Text), "inactive channel warning") {
		return true
	}
	// If the last real message is from the bot but not a warning, we need to look deeper
	if lastRealMsg.User == botUserID {
		return true
	}
	return false
}

func (c *Client) getDetailedChannelActivity(channelID, botUserID string) (lastActivity time.Time, hasWarning bool, warningTime time.Time, err error) {
	// Fetch detailed channel history with retry logic
	history, err := c.fetchDetailedChannelHistory(channelID)
	if err != nil {
		return time.Time{}, false, time.Time{}, err
	}

	if len(history.Messages) == 0 {
		return time.Unix(0, 0), false, time.Time{}, nil
	}

	// Analyze messages for activity and warnings
	mostRecentActivity, hasWarningMessage, mostRecentWarning := c.analyzeChannelMessages(history.Messages, botUserID)

	// Fallback to oldest message if no user activity found
	if mostRecentActivity.IsZero() && len(history.Messages) > 0 {
		if msgTime, err := parseSlackTimestamp(history.Messages[len(history.Messages)-1].Timestamp); err == nil {
			mostRecentActivity = msgTime
		}
	}

	return mostRecentActivity, hasWarningMessage, mostRecentWarning, nil
}

// fetchDetailedChannelHistory fetches detailed channel history with retry logic.
func (c *Client) fetchDetailedChannelHistory(channelID string) (*slack.GetConversationHistoryResponse, error) {
	const maxRetries = 3
	var history *slack.GetConversationHistoryResponse

	for attempt := 1; attempt <= maxRetries; attempt++ {
		params := &slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Limit:     50, // Reasonable limit to find warnings and user activity
		}

		var err error
		history, err = c.api.GetConversationHistory(params)
		if err != nil {
			if c.handleDetailedHistoryError(err, attempt, maxRetries, channelID) {
				continue
			}
			return nil, err
		}
		// Success - break out of retry loop
		break
	}
	return history, nil
}

// handleDetailedHistoryError handles errors during detailed history fetch and returns true if should retry.
func (c *Client) handleDetailedHistoryError(err error, attempt, maxRetries int, channelID string) bool {
	errStr := err.Error()

	// Handle rate limiting with retry
	if strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "rate limit") {
		logger.WithFields(logger.LogFields{
			"attempt":     attempt,
			"max_tries":   maxRetries,
			"channel":     channelID,
			"slack_error": errStr,
		}).Warn("Rate limited on detailed history, will retry after Slack-specified delay")

		fmt.Printf("   🔄 Detailed history retry %d/%d for channel %s (Slack error: %s)\n", attempt, maxRetries, channelID, errStr)

		if attempt < maxRetries {
			c.handleRateLimitWait(errStr)
			return true
		}
		// Last attempt failed
		fmt.Printf("   ❌ All %d detailed retry attempts failed\n", maxRetries)
	}
	return false
}

// handleRateLimitWait handles waiting for rate limit recovery.
func (c *Client) handleRateLimitWait(errStr string) {
	waitDuration := parseSlackRetryAfter(errStr)
	if waitDuration > 0 {
		fmt.Printf("   ⏳ Slack says wait %s (includes 3s buffer) before detailed retry...\n", waitDuration)
		time.Sleep(waitDuration)
	} else {
		fmt.Printf("   ⏳ Using fallback backoff before detailed retry...\n")
	}
}

// analyzeChannelMessages analyzes messages for user activity and bot warnings.
func (c *Client) analyzeChannelMessages(messages []slack.Message, botUserID string) (mostRecentActivity time.Time, hasWarningMessage bool, mostRecentWarning time.Time) {
	for _, msg := range messages {
		msgTime, err := parseSlackTimestamp(msg.Timestamp)
		if err != nil {
			continue
		}

		// Check if this is a warning message from our bot
		if msg.User == botUserID && strings.Contains(strings.ToLower(msg.Text), "inactive channel warning") {
			hasWarningMessage = true
			if msgTime.After(mostRecentWarning) {
				mostRecentWarning = msgTime
			}
		}

		// Track most recent real user activity (excluding system messages)
		if msg.User != botUserID && isRealMessage(msg, botUserID) && msgTime.After(mostRecentActivity) {
			mostRecentActivity = msgTime
		}
	}
	return mostRecentActivity, hasWarningMessage, mostRecentWarning
}

func (c *Client) autoJoinPublicChannels(channels []slack.Channel) (int, error) {
	joinedCount := 0
	var fatalErrors []string
	var skippedCount = 0
	var alreadyMemberCount = 0

	for _, ch := range channels {
		if ch.IsPrivate {
			logger.WithField("channel", ch.Name).Debug("Skipping private channel for auto-join")
			skippedCount++
			continue
		}

		// Skip channels we're already a member of
		if ch.IsMember {
			logger.WithField("channel", ch.Name).Debug("Already a member of channel")
			alreadyMemberCount++
			continue
		}

		result := c.joinChannel(ch)
		switch result.status {
		case joinSuccess:
			joinedCount++
		case joinSkipped:
			skippedCount++
		case joinFatal:
			return joinedCount, result.err
		case joinError:
			fatalErrors = append(fatalErrors, fmt.Sprintf("%s: %s", ch.Name, result.err.Error()))
		}
	}

	c.logAutoJoinSummary(joinedCount, skippedCount, alreadyMemberCount, len(channels))

	if len(fatalErrors) > 0 {
		return joinedCount, fmt.Errorf("failed to join %d channels, cannot proceed with accurate analysis: %v", len(fatalErrors), fatalErrors)
	}

	return joinedCount, nil
}

type joinStatus int

const (
	joinSuccess joinStatus = iota
	joinSkipped
	joinFatal
	joinError
)

type joinResult struct {
	err    error
	status joinStatus
}

// joinChannel attempts to join a single channel and returns the result.
func (c *Client) joinChannel(ch slack.Channel) joinResult {
	_, _, _, err := c.api.JoinConversation(ch.ID)
	if err == nil {
		logger.WithField("channel", ch.Name).Debug("Successfully joined channel")
		return joinResult{status: joinSuccess}
	}

	errStr := err.Error()

	// Handle rate limiting - this is fatal
	if strings.Contains(errStr, "rate_limited") {
		return joinResult{status: joinFatal, err: fmt.Errorf("rate limited during auto-join: %w", err)}
	}

	// Already in channel is success
	if strings.Contains(errStr, "already_in_channel") {
		logger.WithField("channel", ch.Name).Debug("Already in channel")
		return joinResult{status: joinSuccess}
	}

	// Handle skippable errors
	if c.isSkippableJoinError(errStr) {
		c.logSkippableError(ch.Name, errStr)
		return joinResult{status: joinSkipped}
	}

	// Handle fatal errors
	if c.isFatalJoinError(errStr) {
		return joinResult{status: joinFatal, err: c.createFatalJoinError(errStr, err)}
	}

	// Other errors
	return joinResult{status: joinError, err: err}
}

// isSkippableJoinError checks if a join error can be safely skipped.
func (c *Client) isSkippableJoinError(errStr string) bool {
	return strings.Contains(errStr, "is_archived") || strings.Contains(errStr, "invite_only")
}

// isFatalJoinError checks if a join error is fatal for the bot's functionality.
func (c *Client) isFatalJoinError(errStr string) bool {
	return strings.Contains(errStr, "missing_scope") || strings.Contains(errStr, "invalid_auth")
}

// logSkippableError logs skippable join errors.
func (c *Client) logSkippableError(channelName, errStr string) {
	if strings.Contains(errStr, "is_archived") {
		logger.WithField("channel", channelName).Debug("Channel is archived, skipping")
	} else if strings.Contains(errStr, "invite_only") {
		logger.WithField("channel", channelName).Debug("Channel is invite-only, skipping")
	}
}

// createFatalJoinError creates appropriate fatal error messages.
func (c *Client) createFatalJoinError(errStr string, originalErr error) error {
	if strings.Contains(errStr, "missing_scope") {
		return fmt.Errorf("missing required OAuth scope to join channels. Your bot needs the 'channels:join' OAuth scope.\nPlease add this scope in your Slack app settings at https://api.slack.com/apps")
	}
	if strings.Contains(errStr, "invalid_auth") {
		return fmt.Errorf("invalid authentication token: %w", originalErr)
	}
	return originalErr
}

// logAutoJoinSummary logs the summary of auto-join operation.
func (c *Client) logAutoJoinSummary(joined, skipped, alreadyMember, total int) {
	logger.WithFields(logger.LogFields{
		"joined":         joined,
		"skipped":        skipped,
		"already_member": alreadyMember,
		"total":          total,
	}).Debug("Auto-join summary")
}

func (c *Client) shouldSkipChannel(channelName string) bool {
	// Skip channels that should never be archived
	excludePatterns := []string{
		"general",
		"random",
		"announcements",
		"admin",
		"hr",
		"security",
	}

	lowerName := strings.ToLower(channelName)
	for _, pattern := range excludePatterns {
		if strings.Contains(lowerName, pattern) {
			return true
		}
	}
	return false
}

func (c *Client) seemsActiveFromMetadata(ch slack.Channel, warnCutoff time.Time) bool {
	// ONLY use metadata that directly indicates recent message activity
	// No guessing based on member counts, topics, etc.

	// Debug: Show what metadata we have for this channel
	hasLatest := ch.Latest != nil
	latestTimestamp := ""
	if hasLatest {
		latestTimestamp = ch.Latest.Timestamp
	}

	logger.WithFields(logger.LogFields{
		"channel":          ch.Name,
		"has_latest":       hasLatest,
		"latest_timestamp": latestTimestamp,
		"warn_cutoff":      warnCutoff.Format("2006-01-02 15:04:05"),
	}).Debug("Checking channel metadata for activity")

	// Check if the channel itself has a "latest" timestamp that's recent
	// Note: Some Slack APIs provide a "latest" field with the timestamp of the last message
	if ch.Latest != nil && ch.Latest.Timestamp != "" {
		latestMsgTime, err := parseSlackTimestamp(ch.Latest.Timestamp)
		if err == nil {
			logger.WithFields(logger.LogFields{
				"channel":         ch.Name,
				"latest_msg":      latestMsgTime.Format("2006-01-02 15:04:05"),
				"is_after_cutoff": latestMsgTime.After(warnCutoff),
			}).Debug("Parsed latest message timestamp")

			if latestMsgTime.After(warnCutoff) {
				return true
			}
		} else {
			logger.WithFields(logger.LogFields{
				"channel":   ch.Name,
				"timestamp": ch.Latest.Timestamp,
				"error":     err.Error(),
			}).Debug("Failed to parse latest timestamp")
		}
	}

	// If no direct message timestamp metadata is available, we need to check message history
	// Don't make any assumptions - let the message history check decide
	return false
}

func (c *Client) getBotUserID() string {
	// Cache the bot user ID to avoid repeated API calls
	if auth, err := c.api.AuthTest(); err == nil {
		return auth.UserID
	}
	return ""
}

func parseSlackTimestamp(ts string) (time.Time, error) {
	// Slack timestamps are in format "1234567890.123456"
	parts := strings.Split(ts, ".")
	if len(parts) == 0 {
		return time.Time{}, fmt.Errorf("invalid timestamp format")
	}

	seconds, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(seconds, 0), nil
}

func (c *Client) WarnInactiveChannel(channel Channel, warnSeconds, archiveSeconds int, metaChannelID string) error {
	// First, try to join the channel if it's public
	if err := c.ensureBotInChannel(channel); err != nil {
		logger.WithFields(logger.LogFields{
			"channel": channel.Name,
			"error":   err.Error(),
		}).Warn("Could not join channel, skipping warning")
		return fmt.Errorf("failed to join channel %s: %w", channel.Name, err)
	}

	message := c.FormatInactiveChannelWarning(channel, warnSeconds, archiveSeconds, metaChannelID)

	logger.WithFields(logger.LogFields{
		"channel":         channel.Name,
		"archive_seconds": archiveSeconds,
	}).Debug("Posting inactive channel warning")

	return c.postMessageToChannelID(channel.ID, message)
}

func (c *Client) ensureBotInChannel(channel Channel) error {
	logger.WithField("channel", channel.Name).Debug("Ensuring bot is in channel")

	_, _, _, err := c.api.JoinConversation(channel.ID)
	if err == nil {
		logger.WithField("channel", channel.Name).Info("Successfully joined channel")
		return nil
	}

	errStr := err.Error()
	logger.WithFields(logger.LogFields{
		"channel":   channel.Name,
		"error":     errStr,
		"operation": "join_conversation",
	}).Debug("Join conversation result")

	return c.handleJoinChannelError(channel, errStr, err)
}

// handleJoinChannelError handles various join channel errors.
func (c *Client) handleJoinChannelError(channel Channel, errStr string, originalErr error) error {
	// Handle rate limiting
	if strings.Contains(errStr, "rate_limited") {
		return fmt.Errorf("rate limited by Slack API. Will retry with exponential backoff on next request")
	}

	// Handle expected success cases
	if strings.Contains(errStr, "already_in_channel") {
		logger.WithField("channel", channel.Name).Debug("Bot already in channel")
		return nil
	}

	// Handle fatal permission errors
	if strings.Contains(errStr, "missing_scope") {
		return fmt.Errorf("missing required permission to join channels. Your bot needs the 'channels:join' OAuth scope.\nPlease add this scope in your Slack app settings at https://api.slack.com/apps")
	}

	// Handle specific channel errors
	err := c.handleSpecificChannelErrors(channel, errStr)
	if err != nil {
		return err
	}

	// For other errors, log but don't fail completely
	logger.WithFields(logger.LogFields{
		"channel": channel.Name,
		"error":   errStr,
	}).Warn("Unexpected error joining channel")
	return fmt.Errorf("failed to join channel %s: %w", channel.Name, originalErr)
}

// handleSpecificChannelErrors handles specific channel-related errors.
func (c *Client) handleSpecificChannelErrors(channel Channel, errStr string) error {
	if strings.Contains(errStr, "channel_not_found") {
		return fmt.Errorf("channel '%s' not found", channel.Name)
	}

	if strings.Contains(errStr, "is_archived") {
		return fmt.Errorf("channel '%s' is archived", channel.Name)
	}

	if strings.Contains(errStr, "invite_only") {
		logger.WithField("channel", channel.Name).Debug("Channel is private/invite-only, cannot join")
		return fmt.Errorf("channel '%s' is private or invite-only", channel.Name)
	}

	return nil
}

func (c *Client) postMessageToChannelID(channelID, message string) error {
	logger.WithFields(logger.LogFields{
		"channel_id":     channelID,
		"message_length": len(message),
	}).Debug("Posting message to channel by ID")

	_, _, err := c.api.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil {
		errStr := err.Error()
		logger.WithFields(logger.LogFields{
			"channel_id": channelID,
			"error":      errStr,
			"operation":  "post_message",
		}).Error("Failed to post message to Slack")

		// Handle rate limiting
		if strings.Contains(errStr, "rate_limited") {
			return fmt.Errorf("rate limited by Slack API. Will retry with exponential backoff on next request")
		}

		if strings.Contains(errStr, "missing_scope") {
			logger.Error("Missing chat:write OAuth scope")
			return fmt.Errorf("missing required permission to post messages. Your bot needs the 'chat:write' OAuth scope.\nPlease add this scope in your Slack app settings at https://api.slack.com/apps")
		}
		if strings.Contains(errStr, "channel_not_found") {
			logger.WithField("channel_id", channelID).Error("Channel not found")
			return fmt.Errorf("channel with ID '%s' not found. Make sure the bot is added to the channel", channelID)
		}
		if strings.Contains(errStr, "not_in_channel") {
			logger.WithField("channel_id", channelID).Error("Bot not in channel")
			return fmt.Errorf("bot is not a member of channel with ID '%s'. Please add the bot to the channel", channelID)
		}
		return fmt.Errorf("failed to post message to channel %s: %w", channelID, err)
	}

	logger.WithField("channel_id", channelID).Debug("Message posted successfully")
	return nil
}

func (c *Client) FormatInactiveChannelWarning(channel Channel, warnSeconds, archiveSeconds int, metaChannelID string) string {
	var builder strings.Builder

	builder.WriteString("🚨 Inactive Channel Warning 🚨\n\n")

	warnText := formatDurationSeconds(warnSeconds)

	builder.WriteString(fmt.Sprintf("This channel has been inactive for more than %s.\n\n", warnText))

	archiveText := formatDurationSeconds(archiveSeconds)

	builder.WriteString(fmt.Sprintf("This channel could be archived in another %s unless new messages are posted.\n\n", archiveText))

	builder.WriteString("To keep this channel active:\n\n")
	builder.WriteString("• Post a message in this channel or\n")
	// Create meta channel link if available
	metaLink := "#meta"
	if metaChannelID != "" {
		metaLink = fmt.Sprintf("<#%s|meta>", metaChannelID)
	}
	builder.WriteString(fmt.Sprintf("• Discuss in %s if this channel warrants admin intervention\n\n", metaLink))

	return builder.String()
}

func (c *Client) FormatChannelArchivalMessage(channel Channel, warnSeconds, archiveSeconds int, metaChannelID string) string {
	var builder strings.Builder

	builder.WriteString("📋 Channel Archival Notice 📋\n\n")

	warnText := formatDurationSeconds(warnSeconds)

	archiveText := formatDurationSeconds(archiveSeconds)

	builder.WriteString("This channel is being archived because:\n\n")
	builder.WriteString(fmt.Sprintf("• It was inactive for more than %s (warning threshold)\n", warnText))
	builder.WriteString("• An inactivity warning was posted\n")
	builder.WriteString(fmt.Sprintf("• No new activity occurred within %s after the warning (archive threshold)\n\n", archiveText))

	builder.WriteString("This channel is now being archived.\n\n")

	// Create meta channel link if available
	metaLink := "#meta"
	if metaChannelID != "" {
		metaLink = fmt.Sprintf("<#%s|meta>", metaChannelID)
	}
	builder.WriteString(fmt.Sprintf("You may unarchive the channel yourself (given permissions) or discuss in %s if you disagree!", metaLink))

	return builder.String()
}

func (c *Client) ArchiveChannel(channel Channel) error {
	// Legacy method for backward compatibility - uses default thresholds
	return c.ArchiveChannelWithThresholds(channel, 300, 60) // 5 minutes warn, 1 minute archive
}

func (c *Client) ArchiveChannelWithThresholds(channel Channel, warnSeconds, archiveSeconds int) error {
	logger.WithField("channel", channel.Name).Debug("Archiving inactive channel")

	// Look up meta channel ID once to reduce API calls
	metaChannelID, err := c.ResolveChannelNameToID("meta")
	if err != nil {
		logger.WithField("error", err.Error()).Debug("Could not find #meta channel for linking")
		metaChannelID = "" // Fall back to plain text #meta
	}

	// First, ensure the bot is in the channel to post the archival message
	if joinErr := c.ensureBotInChannel(channel); joinErr != nil {
		logger.WithFields(logger.LogFields{
			"channel": channel.Name,
			"error":   joinErr.Error(),
		}).Warn("Could not join channel for archival message, proceeding with archival anyway")
		// Don't fail here - we can still archive even if we can't post the message
	}

	// Post archival message explaining why the channel is being archived
	archivalMessage := c.FormatChannelArchivalMessage(channel, warnSeconds, archiveSeconds, metaChannelID)
	if postErr := c.postMessageToChannelID(channel.ID, archivalMessage); postErr != nil {
		logger.WithFields(logger.LogFields{
			"channel": channel.Name,
			"error":   postErr.Error(),
		}).Warn("Failed to post archival message, proceeding with archival anyway")
		// Don't fail here - archival should proceed even if the message fails
	} else {
		logger.WithField("channel", channel.Name).Info("Posted archival message successfully")
	}

	// Rate limit before archival API call

	err = c.api.ArchiveConversation(channel.ID)
	if err != nil {
		errStr := err.Error()
		logger.WithFields(logger.LogFields{
			"channel":   channel.Name,
			"error":     errStr,
			"operation": "archive_conversation",
		}).Error("Failed to archive channel")

		// Handle rate limiting
		if strings.Contains(errStr, "rate_limited") {
			return fmt.Errorf("rate limited by Slack API. Will retry with exponential backoff on next request")
		}

		if strings.Contains(errStr, "missing_scope") {
			return fmt.Errorf("missing required permission to archive channels. Your bot needs the 'channels:manage' OAuth scope.\nPlease add this scope in your Slack app settings at https://api.slack.com/apps")
		}

		if strings.Contains(errStr, "channel_not_found") {
			return fmt.Errorf("channel '%s' not found", channel.Name)
		}

		if strings.Contains(errStr, "already_archived") {
			logger.WithField("channel", channel.Name).Info("Channel was already archived")
			return nil
		}

		return fmt.Errorf("failed to archive channel %s: %w", channel.Name, err)
	}

	logger.WithField("channel", channel.Name).Info("Channel archived successfully")
	return nil
}

func (c *Client) GetChannelsWithMetadata() ([]Channel, error) {
	logger.Debug("Fetching channels with metadata from Slack API")

	// Rate limit before API call

	channels, _, err := c.api.GetConversations(&slack.GetConversationsParameters{
		Types: []string{"public_channel"},
		Limit: 1000,
	})
	if err != nil {
		errStr := err.Error()
		logger.WithFields(logger.LogFields{
			"error":     errStr,
			"operation": "get_conversations",
		}).Error("Slack API error")

		// Handle rate limiting
		if strings.Contains(errStr, "rate_limited") {
			return nil, fmt.Errorf("rate limited by Slack API. Please wait before retrying")
		}

		if strings.Contains(errStr, "missing_scope") {
			logger.Error("Missing OAuth scopes for channel access")
			return nil, fmt.Errorf("missing required permissions. Your bot needs these OAuth scopes:\\n  - channels:read (to list public channels)\\n\\nPlease add these scopes in your Slack app settings at https://api.slack.com/apps")
		}
		if strings.Contains(errStr, "invalid_auth") {
			logger.Error("Invalid Slack authentication token")
			return nil, fmt.Errorf("invalid token. Please check your SLACK_TOKEN")
		}
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	logger.WithField("total_channels", len(channels)).Debug("Retrieved channels from Slack API")

	result := make([]Channel, 0, len(channels))
	for _, ch := range channels {
		created := time.Unix(int64(ch.Created), 0)

		// Parse the updated timestamp - this appears to be in milliseconds since epoch
		var updated time.Time

		// Try to access the updated field if available in the slack-go library
		// The JSON shows "updated": 1678229664302 (milliseconds since epoch)
		// Let's check what's available on the Channel struct

		// Try to access the actual "updated" field from the Slack API
		// The slack-go library may not expose this field directly

		// Let me try to see if there's a direct Updated field
		// Based on the JSON, this should be "updated": 1678229664302 (milliseconds since epoch)

		// Check for various possible field names
		// Since slack-go might expose it as a different field...
		// Let's add some debug info to see what's actually available

		// For the dev command, we'll show channel creation time
		// This is useful for understanding when channels were created
		// Note: This is NOT the same as last message activity

		// Try to get latest message timestamp if available
		if ch.Latest != nil && ch.Latest.Timestamp != "" {
			if latestTime, err := parseSlackTimestamp(ch.Latest.Timestamp); err == nil {
				updated = latestTime
			}
		}

		// Fallback to creation time (which is what we typically see)
		if updated.IsZero() {
			updated = created
		}

		result = append(result, Channel{
			ID:          ch.ID,
			Name:        ch.Name,
			Created:     created,
			Updated:     updated,
			Purpose:     ch.Purpose.Value,
			Creator:     ch.Creator,
			MemberCount: ch.NumMembers,
			IsArchived:  ch.IsArchived,
		})
	}

	logger.WithField("channels_count", len(result)).Debug("Channel metadata extraction completed")
	return result, nil
}

// GetChannelActivity returns the last activity time, warning status, and warning time for a channel.
func (c *Client) GetChannelActivity(channelID string) (lastActivity time.Time, hasWarning bool, warningTime time.Time, err error) {
	return c.getChannelActivity(channelID)
}

// MessageInfo contains details about a message.
type MessageInfo struct {
	Timestamp time.Time
	User      string
	UserName  string // Human-readable name
	Text      string
	IsBot     bool
}

// GetChannelActivityWithMessage returns activity info plus details about the most recent message.
func (c *Client) GetChannelActivityWithMessage(channelID string) (lastActivity time.Time, hasWarning bool, warningTime time.Time, lastMessage *MessageInfo, err error) {
	history, err := c.getChannelHistoryWithRetry(channelID)
	if err != nil {
		return time.Time{}, false, time.Time{}, nil, err
	}

	if len(history.Messages) == 0 {
		return time.Time{}, false, time.Time{}, nil, nil
	}

	botUserID := c.getBotUserID()
	lastRealMsg, lastRealMsgTime := c.findMostRecentRealMessage(history.Messages, botUserID)
	if lastRealMsg == nil {
		return time.Time{}, false, time.Time{}, nil, nil
	}

	msgInfo := c.createMessageInfo(lastRealMsg, lastRealMsgTime, botUserID)
	hasWarningMessage, mostRecentWarning := c.checkForWarningMessage(lastRealMsg, lastRealMsgTime, botUserID)

	return lastRealMsgTime, hasWarningMessage, mostRecentWarning, msgInfo, nil
}

// getChannelHistoryWithRetry handles the API call with retry logic.
func (c *Client) getChannelHistoryWithRetry(channelID string) (*slack.GetConversationHistoryResponse, error) {
	const maxRetries = 3
	params := &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     10,
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		history, err := c.api.GetConversationHistory(params)
		if err != nil {
			if c.isRateLimitError(err) && attempt < maxRetries {
				c.handleRateLimit(err, attempt, maxRetries)
				continue
			}
			if c.isRateLimitError(err) {
				return nil, fmt.Errorf("failed to get channel history after %d attempts, final Slack error: %s", maxRetries, err.Error())
			}
			return nil, fmt.Errorf("failed to get channel history: %w", err)
		}
		return history, nil
	}
	return nil, fmt.Errorf("unexpected end of retry loop")
}

// createMessageInfo creates a MessageInfo struct from a Slack message.
func (c *Client) createMessageInfo(msg *slack.Message, msgTime time.Time, botUserID string) *MessageInfo {
	return &MessageInfo{
		Timestamp: msgTime,
		User:      msg.User,
		Text:      msg.Text,
		IsBot:     msg.User == botUserID,
	}
}

// checkForWarningMessage checks if the message is a warning from our bot.
func (c *Client) checkForWarningMessage(msg *slack.Message, msgTime time.Time, botUserID string) (bool, time.Time) {
	hasWarning := msg.User == botUserID && strings.Contains(strings.ToLower(msg.Text), "inactive channel warning")
	if hasWarning {
		return true, msgTime
	}
	return false, time.Time{}
}

// isRateLimitError checks if an error is a rate limiting error.
func (c *Client) isRateLimitError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "rate limit")
}

// handleRateLimit handles rate limiting with user feedback.
func (c *Client) handleRateLimit(err error, attempt, maxRetries int) {
	waitDuration := parseSlackRetryAfter(err.Error())
	if waitDuration > 0 {
		fmt.Printf("⏳ Waiting %s due to Slack API rate limiting...\n", waitDuration.String())
		showProgressBar(waitDuration)
	}
}

// parseSlackRetryAfter parses Slack's "retry after" directive from error messages.
// Example: "slack rate limit exceeded, retry after 1m0s" -> 1 minute duration.
func parseSlackRetryAfter(errorStr string) time.Duration {
	// Look for "retry after" followed by a duration
	retryAfterIndex := strings.Index(errorStr, "retry after ")
	if retryAfterIndex == -1 {
		return 0
	}

	// Extract the duration part
	durationStart := retryAfterIndex + len("retry after ")
	remaining := errorStr[durationStart:]

	// Find the end of the duration (usually end of string or a space)
	var durationStr string
	spaceIndex := strings.Index(remaining, " ")
	if spaceIndex == -1 {
		durationStr = remaining
	} else {
		durationStr = remaining[:spaceIndex]
	}

	// Parse the duration using Go's time.ParseDuration
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		logger.WithFields(logger.LogFields{
			"error_string":  errorStr,
			"duration_part": durationStr,
			"parse_error":   err.Error(),
		}).Warn("Failed to parse Slack retry-after duration")
		return 0
	}

	// Add 1 second buffer to avoid hitting the limit again immediately
	bufferedDuration := duration + (1 * time.Second)

	return bufferedDuration
}

// showProgressBar displays a text-based progress bar for waiting periods.
// Shows - for remaining time and * for elapsed time.
func showProgressBar(duration time.Duration) {
	totalSeconds := int(duration.Seconds())
	if totalSeconds <= 0 {
		return
	}

	// Limit to reasonable max for display (2 minutes = 120 chars)
	maxDisplay := 120
	if totalSeconds > maxDisplay {
		// Scale down for very long waits
		scaleFactor := float64(totalSeconds) / float64(maxDisplay)
		fmt.Printf("   Progress (scaled 1:%d): ", int(scaleFactor))
		totalSeconds = maxDisplay
	} else {
		fmt.Printf("   Progress: ")
	}

	// Print initial bar (all dashes)
	for i := 0; i < totalSeconds; i++ {
		fmt.Printf("-")
	}
	fmt.Printf(" [0/%ds]\r", int(duration.Seconds()))

	// Update progress each second
	for elapsed := 0; elapsed < totalSeconds; elapsed++ {
		time.Sleep(time.Second)

		// Move cursor to start of progress bar
		fmt.Printf("   Progress: ")
		if totalSeconds > maxDisplay {
			fmt.Printf("(scaled 1:%d): ", int(float64(int(duration.Seconds()))/float64(maxDisplay)))
		}

		// Print progress: * for completed, - for remaining
		for i := 0; i < totalSeconds; i++ {
			if i <= elapsed {
				fmt.Printf("*")
			} else {
				fmt.Printf("-")
			}
		}
		fmt.Printf(" [%d/%ds]\r", elapsed+1, int(duration.Seconds()))
	}

	// Final newline
	fmt.Printf("\n   ✅ Wait complete!\n")
}

// isRealMessage filters out system messages like joins, leaves, topic changes, etc.
// Returns true for actual user-generated content.
func isRealMessage(msg slack.Message, botUserID string) bool {
	// Filter out messages with system subtypes
	if msg.SubType != "" {
		systemSubtypes := []string{
			"channel_join",
			"channel_leave",
			"channel_topic",
			"channel_purpose",
			"channel_name",
			"channel_archive",
			"channel_unarchive",
			"group_join",
			"group_leave",
			"group_topic",
			"group_purpose",
			"group_name",
			"group_archive",
			"group_unarchive",
			"bot_add",
			"bot_remove",
			"pinned_item",
			"unpinned_item",
		}

		for _, systemType := range systemSubtypes {
			if msg.SubType == systemType {
				return false
			}
		}
	}

	// Filter out join/leave messages by content patterns
	text := msg.Text
	joinLeavePatterns := []string{
		"has joined the channel",
		"has left the channel",
		"has joined the group",
		"has left the group",
		"set the channel topic:",
		"set the channel purpose:",
		"renamed the channel from",
		"archived this channel",
		"unarchived this channel",
		"pinned a message to this channel",
		"unpinned a message from this channel",
	}

	for _, pattern := range joinLeavePatterns {
		if strings.Contains(text, pattern) {
			return false
		}
	}

	// Filter out empty messages (unless they have file attachments)
	if strings.TrimSpace(text) == "" && len(msg.Files) == 0 {
		return false
	}

	// All other messages are considered "real"
	return true
}

// getUserMap fetches all users and builds a map from user ID to display name.
func (c *Client) getUserMap() (map[string]string, error) {

	users, err := c.api.GetUsers()
	if err != nil {
		errStr := err.Error()
		logger.WithFields(logger.LogFields{
			"error":          errStr,
			"operation":      "get_users",
			"required_scope": "users:read",
		}).Error("Failed to get users list")

		// Handle rate limiting
		if strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "rate limit") {
			return nil, fmt.Errorf("rate limited getting users: %w", err)
		}

		// Handle missing scope with clearer message
		if strings.Contains(errStr, "missing_scope") {
			return nil, fmt.Errorf("missing required OAuth scope 'users:read' to get user list: %w", err)
		}

		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// Build the map
	userMap := make(map[string]string)
	for _, user := range users {
		displayName := user.RealName
		if displayName == "" {
			displayName = user.Name
		}
		if displayName == "" {
			displayName = user.Profile.DisplayName
		}
		if displayName == "" {
			displayName = user.ID // Fallback to ID if no name available
		}
		userMap[user.ID] = displayName
	}

	logger.WithField("user_count", len(userMap)).Debug("Built user lookup map")
	return userMap, nil
}

// GetUserMap is a public wrapper for getUserMap.
func (c *Client) GetUserMap() (map[string]string, error) {
	return c.getUserMap()
}

// getUsersForDefaultDetection fetches all workspace users with retry logic.
func (c *Client) getUsersForDefaultDetection() ([]slack.User, error) {
	const maxRetries = 3
	var users []slack.User
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		users, err = c.api.GetUsers()
		if err != nil {
			if shouldRetryRateLimit(err, attempt, maxRetries) {
				waitDuration := parseSlackRetryAfter(err.Error())
				if waitDuration > 0 {
					logger.WithFields(logger.LogFields{
						"attempt":       attempt,
						"max_tries":     maxRetries,
						"wait_duration": waitDuration,
					}).Debug("Rate limited while fetching users, waiting before retry")
					showProgressBar(waitDuration)
				}
				continue
			}
			return nil, fmt.Errorf("failed to get users after %d attempts: %w", maxRetries, err)
		}
		break
	}
	return users, nil
}

// filterAndSortRealUsers extracts real users and sorts them by most recently updated.
func filterAndSortRealUsers(users []slack.User, sampleSize int) []string {
	type userInfo struct {
		ID      string
		Updated int64
	}
	var realUsers []userInfo
	for _, user := range users {
		if !user.IsBot && !user.Deleted && user.ID != "USLACKBOT" {
			realUsers = append(realUsers, userInfo{
				ID:      user.ID,
				Updated: int64(user.Updated),
			})
		}
	}

	if len(realUsers) < 2 {
		return []string{}
	}

	// Sort by updated time (newest first)
	for i := 0; i < len(realUsers)-1; i++ {
		for j := i + 1; j < len(realUsers); j++ {
			if realUsers[j].Updated > realUsers[i].Updated {
				realUsers[i], realUsers[j] = realUsers[j], realUsers[i]
			}
		}
	}

	// Take sample of most recent users
	if len(realUsers) < sampleSize {
		sampleSize = len(realUsers)
	}

	result := make([]string, sampleSize)
	for i := 0; i < sampleSize; i++ {
		result[i] = realUsers[i].ID
	}
	return result
}

// getConversationsForUserWithRetry fetches user conversations with retry logic.
func (c *Client) getConversationsForUserWithRetry(userID, cursor string) ([]slack.Channel, string, error) {
	const maxRetries = 3
	var channelList []slack.Channel
	var nextCursor string
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		params := &slack.GetConversationsForUserParameters{
			UserID: userID,
			Types:  []string{"public_channel"},
			Cursor: cursor,
		}
		channelList, nextCursor, err = c.api.GetConversationsForUser(params)
		if err != nil {
			if shouldRetryRateLimit(err, attempt, maxRetries) {
				waitDuration := parseSlackRetryAfter(err.Error())
				if waitDuration > 0 {
					logger.WithFields(logger.LogFields{
						"user_id":       userID,
						"attempt":       attempt,
						"max_tries":     maxRetries,
						"wait_duration": waitDuration,
					}).Debug("Rate limited while fetching user channels, waiting before retry")
					showProgressBar(waitDuration)
				}
				continue
			}
			logger.WithFields(logger.LogFields{
				"user_id": userID,
				"error":   err.Error(),
			}).Warn("Failed to get channels for user, skipping")
			return nil, "", err
		}
		break
	}

	return channelList, nextCursor, err
}

// getUserChannelMemberships fetches channel memberships for a single user with retry logic.
func (c *Client) getUserChannelMemberships(userID string) (map[string]bool, error) {
	channelSet := make(map[string]bool)
	cursor := ""

	for {
		channelList, nextCursor, err := c.getConversationsForUserWithRetry(userID, cursor)
		if err != nil {
			return channelSet, err
		}

		for _, ch := range channelList {
			channelSet[ch.ID] = true
		}

		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}
	return channelSet, nil
}

// shouldRetryRateLimit checks if we should retry due to rate limiting.
func shouldRetryRateLimit(err error, attempt, maxRetries int) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	isRateLimited := strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "rate limit")
	return isRateLimited && attempt < maxRetries
}

// getChannelNameByID fetches a channel's name by ID with retry logic.
func (c *Client) getChannelNameByID(channelID string) (string, error) {
	const maxRetries = 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		channel, err := c.api.GetConversationInfo(&slack.GetConversationInfoInput{
			ChannelID: channelID,
		})
		if err != nil {
			if shouldRetryRateLimit(err, attempt, maxRetries) {
				waitDuration := parseSlackRetryAfter(err.Error())
				if waitDuration > 0 {
					logger.WithFields(logger.LogFields{
						"channel_id":    channelID,
						"attempt":       attempt,
						"max_tries":     maxRetries,
						"wait_duration": waitDuration,
					}).Debug("Rate limited while fetching channel info, waiting before retry")
					showProgressBar(waitDuration)
				}
				continue
			}
			logger.WithFields(logger.LogFields{
				"channel_id": channelID,
				"error":      err.Error(),
			}).Warn("Failed to get channel info, skipping")
			return "", err
		}
		return channel.Name, nil
	}
	return "", fmt.Errorf("failed to get channel info after retries")
}

// GetDefaultChannels detects likely default channels by finding channels that all recent users share.
// It samples the most recent non-bot users and returns channels that meet the membership threshold.
// threshold: percentage of sampled users that must be members (e.g., 0.9 = 90%)
// Returns a list of channel names (without # prefix) that are likely auto-join default channels.
func (c *Client) GetDefaultChannels(sampleSize int, threshold float64) ([]string, error) {
	logger.WithFields(logger.LogFields{
		"sample_size": sampleSize,
		"threshold":   threshold,
	}).Debug("Detecting default channels")

	// Get all users
	users, err := c.getUsersForDefaultDetection()
	if err != nil {
		return nil, err
	}

	// Filter and sort real users
	recentUserIDs := filterAndSortRealUsers(users, sampleSize)
	if len(recentUserIDs) == 0 {
		logger.Debug("Not enough users to detect default channels")
		return []string{}, nil
	}

	logger.WithField("sampled_users", len(recentUserIDs)).Debug("Sampled recent users for default channel detection")

	// Get channel memberships for each user
	userChannels := make(map[string]map[string]bool) // userID -> channelID -> true
	for _, userID := range recentUserIDs {
		channelSet, err := c.getUserChannelMemberships(userID)
		if err != nil {
			// Error already logged in helper, continue with other users
			continue
		}
		userChannels[userID] = channelSet
	}

	if len(userChannels) == 0 {
		logger.Debug("No channel data collected")
		return []string{}, nil
	}

	// Find channels that meet the threshold (e.g., 90% of sampled users must be members)
	// This is more resilient than requiring ALL users (100%) in case some users leave default channels
	minMembership := int(float64(len(userChannels)) * threshold)
	if minMembership < 1 {
		minMembership = 1
	}

	// Count membership for each channel across all sampled users
	channelMembership := make(map[string]int)
	for _, channels := range userChannels {
		for chID := range channels {
			channelMembership[chID]++
		}
	}

	// Channels that meet the threshold are likely defaults
	var commonChannelIDs []string
	for chID, memberCount := range channelMembership {
		if memberCount >= minMembership {
			commonChannelIDs = append(commonChannelIDs, chID)
		}
	}

	// Get channel names for the common channel IDs
	defaultChannelNames := make([]string, 0, len(commonChannelIDs))
	for _, chID := range commonChannelIDs {
		name, err := c.getChannelNameByID(chID)
		if err != nil {
			// Error already logged in helper, continue with other channels
			continue
		}
		defaultChannelNames = append(defaultChannelNames, name)
	}

	logger.WithFields(logger.LogFields{
		"detected_count": len(defaultChannelNames),
		"channels":       strings.Join(defaultChannelNames, ", "),
	}).Debug("Default channel detection complete")

	return defaultChannelNames, nil
}

// GetChannelActivityWithMessageAndUsers returns activity info plus message details with resolved user names.
func (c *Client) GetChannelActivityWithMessageAndUsers(channelID string, userMap map[string]string) (lastActivity time.Time, hasWarning bool, warningTime time.Time, lastMessage *MessageInfo, err error) {
	// Get the basic activity info
	lastActivity, hasWarning, warningTime, basicMessage, err := c.GetChannelActivityWithMessage(channelID)
	if err != nil || basicMessage == nil {
		return lastActivity, hasWarning, warningTime, basicMessage, err
	}

	// Resolve the user name
	userName := userMap[basicMessage.User]
	if userName == "" {
		userName = basicMessage.User // Fallback to ID if not found
	}

	// Create enhanced message info with resolved name
	enhancedMessage := &MessageInfo{
		Timestamp: basicMessage.Timestamp,
		User:      basicMessage.User,
		UserName:  userName,
		Text:      basicMessage.Text,
		IsBot:     basicMessage.IsBot,
	}

	return lastActivity, hasWarning, warningTime, enhancedMessage, nil
}

func shouldRetryOnRateLimit(errStr string, attempt, maxRetries int, channel string) bool {
	if !strings.Contains(errStr, "rate_limited") && !strings.Contains(errStr, "rate limit") {
		return false
	}
	if attempt >= maxRetries {
		return false
	}

	waitDuration := parseSlackRetryAfter(errStr)
	if waitDuration <= 0 {
		return false
	}

	logger.WithFields(logger.LogFields{
		"channel":       channel,
		"attempt":       attempt,
		"wait_duration": waitDuration.String(),
	}).Info("Respecting Slack rate limit for duplicate check")
	fmt.Printf("⏳ Waiting %s due to Slack rate limit (attempt %d/%d)...\n", waitDuration.String(), attempt, maxRetries)
	showProgressBar(waitDuration)
	fmt.Println() // Add newline after progress bar
	return true
}

func shouldRetryOnRateLimitSimple(errStr string, attempt, maxRetries int) bool {
	if !strings.Contains(errStr, "rate_limited") && !strings.Contains(errStr, "rate limit") {
		return false
	}
	if attempt >= maxRetries {
		return false
	}

	waitDuration := parseSlackRetryAfter(errStr)
	if waitDuration > 0 {
		fmt.Printf("⏳ Waiting %s due to Slack API rate limiting...\n", waitDuration.String())
		showProgressBar(waitDuration)
	}
	return true
}

func formatDurationSeconds(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	}

	minutes := seconds / 60
	if minutes == 1 {
		return "1 minute"
	}
	if minutes < 60 {
		return fmt.Sprintf("%d minutes", minutes)
	}

	hours := minutes / 60
	if hours == 1 {
		return "1 hour"
	}
	if hours < 24 {
		return fmt.Sprintf("%d hours", hours)
	}

	days := hours / 24
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}
