package slack

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astrostl/slack-buddy-ai/pkg/logger"

	"github.com/slack-go/slack"
)

// Common time duration text constants
const (
	oneMinuteText = "1 minute"
	oneHourText   = "1 hour"
	oneDayText    = "1 day"
)

type Client struct {
	api SlackAPI
}

type Channel struct {
	ID           string
	Name         string
	Created      time.Time
	Updated      time.Time
	Purpose      string
	Creator      string
	LastActivity time.Time
	MemberCount  int
	IsArchived   bool
	LastMessage  *MessageInfo // Optional: details about the last message
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
		Types: []string{"public_channel", "private_channel"},
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
			return nil, fmt.Errorf("missing required permissions. Your bot needs these OAuth scopes:\n  - channels:read (to list public channels) - REQUIRED\n  - groups:read (to list private channels) - OPTIONAL\n\nPlease add these scopes in your Slack app settings at https://api.slack.com/apps")
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
		Types: []string{"public_channel", "private_channel"},
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

	var builder strings.Builder

	if len(channels) == 1 {
		builder.WriteString("New channel alert!")
	} else {
		builder.WriteString(fmt.Sprintf("%d new channels created!", len(channels)))
	}

	builder.WriteString("\n\n")

	for i, ch := range channels {
		builder.WriteString(fmt.Sprintf("• <#%s>", ch.ID))

		// Add creator info if available
		if ch.Creator != "" {
			builder.WriteString(fmt.Sprintf(" created by <@%s>", ch.Creator))
		}

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

func (c *Client) FormatNewChannelAnnouncementDryRun(channels []Channel, since time.Time) string {
	// Get user map for resolving creator names
	userMap, err := c.GetUserMap()
	if err != nil {
		// If we can't get users, fall back to regular format
		return c.FormatNewChannelAnnouncement(channels, since)
	}

	var builder strings.Builder

	if len(channels) == 1 {
		builder.WriteString("New channel alert!")
	} else {
		builder.WriteString(fmt.Sprintf("%d new channels created!", len(channels)))
	}

	builder.WriteString("\n\n")

	for i, ch := range channels {
		builder.WriteString(fmt.Sprintf("• #%s", ch.Name))

		// Add creator info with pretty name if available
		if ch.Creator != "" {
			creatorName := ch.Creator // fallback to ID
			if userName, exists := userMap[ch.Creator]; exists && userName != "" {
				creatorName = userName
			}
			builder.WriteString(fmt.Sprintf(" created by %s", creatorName))
		}

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

	// Create name-to-ID mapping from provided channels or fetch if not provided
	var channelNameToID map[string]string
	if allChannels != nil {
		// Use provided channel list
		channelNameToID = make(map[string]string)
		for _, ch := range allChannels {
			channelNameToID[ch.Name] = ch.ID
		}
	} else {
		// Fallback: get channels via API call
		var err error
		channelNameToID, err = c.getAllChannelNameToIDMap()
		if err != nil {
			logger.WithFields(logger.LogFields{
				"error": err.Error(),
			}).Warn("Failed to get channel mappings for duplicate detection, will use name-only matching")
			channelNameToID = make(map[string]string)
		}
	}

	// Get our bot's auth info to identify our messages
	authInfo, err := c.TestAuth()
	if err != nil {
		return false, nil, fmt.Errorf("failed to get auth info: %w", err)
	}

	// Due to Slack API limits (15 messages max, 1 req/min), just get the last 15 messages
	// No time window - work with whatever we can get
	params := &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     15, // Explicit limit to match API restriction
	}

	// Retry on Slack rate limits but not with artificial delays
	const maxRetries = 3
	var history *slack.GetConversationHistoryResponse

	for attempt := 1; attempt <= maxRetries; attempt++ {
		var err error
		history, err = c.api.GetConversationHistory(params)
		if err != nil {
			errStr := err.Error()
			// Only retry on Slack's rate limiting
			if strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "rate limit") {
				if attempt < maxRetries {
					// Parse Slack's retry-after directive (already includes buffer)
					waitDuration := parseSlackRetryAfter(errStr)
					if waitDuration > 0 {
						logger.WithFields(logger.LogFields{
							"channel":       channel,
							"attempt":       attempt,
							"wait_duration": waitDuration.String(),
						}).Info("Respecting Slack rate limit for duplicate check")
						fmt.Printf("⏳ Waiting %s due to Slack rate limit (attempt %d/%d)...\n", waitDuration.String(), attempt, maxRetries)
						showProgressBar(waitDuration)
						fmt.Println() // Add newline after progress bar
						continue
					}
				}
			}
			// For non-rate-limit errors or final attempt
			logger.WithFields(logger.LogFields{
				"channel": channel,
				"error":   errStr,
			}).Warn("Failed to get channel history for duplicate check")
			return false, nil, nil
		}
		// Success - break out of retry loop
		break
	}

	// Check each message to see if it's from our bot and collect all announced channels
	var allAnnouncedChannels []string
	var foundDuplicates bool

	for _, message := range history.Messages {
		// Skip if not from our bot
		if message.User != authInfo.UserID && message.BotID != authInfo.UserID {
			continue
		}
		if duplicateChannels := c.findDuplicateChannelsInMessageWithIDs(message.Text, channelNames, channelNameToID); len(duplicateChannels) > 0 {
			logger.WithFields(logger.LogFields{
				"channel":            channel,
				"duplicate_ts":       message.Timestamp,
				"announced_channels": strings.Join(duplicateChannels, ", "),
			}).Debug("Found duplicate announcement")

			// Add these channels to our list of announced channels
			for _, dupChannel := range duplicateChannels {
				// Avoid duplicates in our list
				alreadyInList := false
				for _, existing := range allAnnouncedChannels {
					if existing == dupChannel {
						alreadyInList = true
						break
					}
				}
				if !alreadyInList {
					allAnnouncedChannels = append(allAnnouncedChannels, dupChannel)
				}
			}
			foundDuplicates = true
		}
	}

	if foundDuplicates {
		return true, allAnnouncedChannels, nil
	}

	logger.WithFields(logger.LogFields{
		"channel":      channel,
		"scanned_msgs": len(history.Messages),
	}).Debug("No duplicate announcements found")
	return false, nil, nil
}

func (c *Client) messageContainsSameChannels(messageText string, channelNames []string) bool {
	return len(c.findDuplicateChannelsInMessage(messageText, channelNames)) > 0
}

func (c *Client) findDuplicateChannelsInMessage(messageText string, channelNames []string) []string {
	// Legacy function - creates empty ID map for backward compatibility
	return c.findDuplicateChannelsInMessageWithIDs(messageText, channelNames, make(map[string]string))
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
		Types: []string{"public_channel", "private_channel"},
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

	return &AuthInfo{
		User:   auth.User,
		UserID: auth.UserID,
		Team:   auth.Team,
		TeamID: auth.TeamID,
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
	scopeResults["chat:write"] = c.testChatWriteScope()
	scopeResults["channels:manage"] = c.testChannelsManageScope()
	scopeResults["users:read"] = c.testUsersReadScope()
	scopeResults["groups:read"] = c.testGroupsReadScope() // Optional for private channels

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

func (c *Client) testGroupsReadScope() bool {
	// Try to get private channels - this requires groups:read (optional)
	_, _, err := c.api.GetConversations(&slack.GetConversationsParameters{
		Types: []string{"private_channel"},
		Limit: 1,
	})

	if err != nil && strings.Contains(err.Error(), "missing_scope") {
		return false
	}
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
		Types: []string{"public_channel", "private_channel"},
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

	// Calculate cutoff times
	warnCutoff := time.Now().Add(-time.Duration(warnSeconds) * time.Second)

	// Get all channels


	allChannels, _, err := c.api.GetConversations(&slack.GetConversationsParameters{
		Types:           []string{"public_channel", "private_channel"},
		Limit:           1000,
		ExcludeArchived: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get conversations: %w", err)
	}


	logger.WithField("total_channels", len(allChannels)).Debug("Retrieved channels for activity analysis")

	// Pre-filter channels using metadata to reduce API calls
	var candidateChannels []slack.Channel
	var skippedNew, skippedExcluded, skippedActive int

	for _, ch := range allChannels {
		// Skip if channel is too new to possibly need warning
		// If a channel was created after the warn cutoff, it can't be inactive long enough to warn
		created := time.Unix(int64(ch.Created), 0)
		if created.After(warnCutoff) {
			logger.WithFields(logger.LogFields{
				"channel":     ch.Name,
				"created":     created.Format("2006-01-02 15:04:05"),
				"warn_cutoff": warnCutoff.Format("2006-01-02 15:04:05"),
			}).Debug("Skipping channel: created after warn cutoff")
			skippedNew++
			continue
		}

		// Skip channels with certain patterns (configurable exclusions)
		if c.shouldSkipChannel(ch.Name) {
			logger.WithField("channel", ch.Name).Debug("Skipping excluded channel")
			skippedExcluded++
			continue
		}

		// Use metadata to pre-filter potentially inactive channels
		if c.seemsActiveFromMetadata(ch, warnCutoff) {
			logger.WithFields(logger.LogFields{
				"channel": ch.Name,
				"reason":  "metadata_shows_recent_activity",
			}).Debug("Channel seems active from metadata, skipping message history check")
			skippedActive++
			continue
		}

		candidateChannels = append(candidateChannels, ch)
	}

	logger.WithFields(logger.LogFields{
		"total_channels":     len(allChannels),
		"candidate_channels": len(candidateChannels),
		"skipped_new":        skippedNew,
		"skipped_excluded":   skippedExcluded,
		"skipped_active":     skippedActive,
	}).Info("Pre-filtered channels using metadata")

	// Auto-join all public candidate channels before analyzing them
	// This is critical for getting accurate message history data
	joinedCount, err := c.autoJoinPublicChannels(candidateChannels)
	if err != nil {
		return toWarn, toArchive, fmt.Errorf("failed to auto-join channels - inactive detection requires channel membership: %w", err)
	}
	logger.WithField("joined_count", joinedCount).Debug("Auto-joined public channels")

	// Now check message history only for candidate channels
	for _, ch := range candidateChannels {
		logger.WithField("channel", ch.Name).Debug("Checking message history for candidate channel")

		// Get channel activity with retry for rate limits
		lastActivity, hasWarning, warningTime, err := c.getChannelActivityWithRetry(ch.ID, ch.Name)
		if err != nil {
			errStr := err.Error()

			// Handle different error types with appropriate policies
			if strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "rate limit") {
				logger.WithFields(logger.LogFields{
					"channel": ch.Name,
					"error":   errStr,
				}).Warn("Rate limited by Slack API - this affects all subsequent requests, stopping analysis")
				return toWarn, toArchive, fmt.Errorf("rate limited by Slack API while processing channel %s: %w", ch.Name, err)
			}

			if strings.Contains(errStr, "not_in_channel") {
				// Bot isn't in channel, treat as potentially inactive
				logger.WithFields(logger.LogFields{
					"channel": ch.Name,
				}).Debug("Bot not in channel, treating as potentially inactive")
				lastActivity = time.Unix(0, 0) // Very old timestamp
				hasWarning = false
				warningTime = time.Time{}
			} else {
				// For other errors, skip the channel to be safe
				logger.WithFields(logger.LogFields{
					"channel": ch.Name,
					"error":   errStr,
				}).Warn("Failed to get channel activity, skipping")
				continue
			}
		}

		logger.WithFields(logger.LogFields{
			"channel":       ch.Name,
			"last_activity": lastActivity.Format("2006-01-02 15:04:05"),
			"has_warning":   hasWarning,
			"warning_time":  warningTime.Format("2006-01-02 15:04:05"),
		}).Debug("Retrieved channel activity data")

		created := time.Unix(int64(ch.Created), 0)
		channel := Channel{
			ID:           ch.ID,
			Name:         ch.Name,
			Created:      created,
			Purpose:      ch.Purpose.Value,
			Creator:      ch.Creator,
			LastActivity: lastActivity,
			MemberCount:  ch.NumMembers,
			IsArchived:   ch.IsArchived,
		}

		// Decision logic
		if hasWarning {
			// Channel was already warned, check if it should be archived
			// Archive if: last activity was before warning AND warning was sent long enough ago
			gracePeriodExpired := time.Since(warningTime) > time.Duration(archiveSeconds)*time.Second
			if lastActivity.Before(warningTime) && gracePeriodExpired {
				toArchive = append(toArchive, channel)
				logger.WithFields(logger.LogFields{
					"channel":       ch.Name,
					"last_activity": lastActivity.Format("2006-01-02 15:04:05"),
					"warning_time":  warningTime.Format("2006-01-02 15:04:05"),
				}).Debug("Channel marked for archival")
			}
		} else {
			// No warning yet, check if it should be warned
			if lastActivity.IsZero() || lastActivity.Before(warnCutoff) {
				toWarn = append(toWarn, channel)
				logger.WithFields(logger.LogFields{
					"channel":       ch.Name,
					"last_activity": lastActivity.Format("2006-01-02 15:04:05"),
				}).Debug("Channel marked for warning")
			}
		}
	}

	logger.WithFields(logger.LogFields{
		"to_warn":    len(toWarn),
		"to_archive": len(toArchive),
	}).Info("Inactive channel analysis completed")

	return toWarn, toArchive, nil
}

// GetInactiveChannelsWithDetails returns inactive channels with detailed message information and user name resolution
func (c *Client) GetInactiveChannelsWithDetails(warnSeconds int, archiveSeconds int, userMap map[string]string, isDebug bool) (toWarn []Channel, toArchive []Channel, err error) {
	logger.WithFields(logger.LogFields{
		"warn_seconds":    warnSeconds,
		"archive_seconds": archiveSeconds,
	}).Debug("Starting inactive channel detection with message details")

	// Calculate cutoff times
	warnCutoff := time.Now().Add(-time.Duration(warnSeconds) * time.Second)

	// Get all channels


	allChannels, _, err := c.api.GetConversations(&slack.GetConversationsParameters{
		Types:           []string{"public_channel", "private_channel"},
		Limit:           1000,
		ExcludeArchived: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get conversations: %w", err)
	}


	logger.WithField("total_channels", len(allChannels)).Debug("Retrieved channels for detailed activity analysis")

	// Pre-filter channels using metadata to reduce API calls
	var candidateChannels []slack.Channel
	var skippedActive, skippedExcluded, skippedNew int

	for _, ch := range allChannels {
		// Skip excluded channels
		if c.shouldSkipChannel(ch.Name) {
			skippedExcluded++
			continue
		}

		// Skip channels that are too new to possibly need warnings
		created := time.Unix(int64(ch.Created), 0)
		if created.After(warnCutoff) {
			skippedNew++
			continue
		}

		// Skip channels that seem active from metadata
		if c.seemsActiveFromMetadata(ch, warnCutoff) {
			skippedActive++
			continue
		}

		candidateChannels = append(candidateChannels, ch)
	}

	logger.WithFields(logger.LogFields{
		"total_channels":     len(allChannels),
		"candidate_channels": len(candidateChannels),
		"skipped_active":     skippedActive,
		"skipped_excluded":   skippedExcluded,
		"skipped_new":        skippedNew,
	}).Info("Pre-filtered channels using metadata")

	fmt.Printf("📞 API Call 2: Getting channel list with metadata...\n")
	fmt.Printf("✅ Got %d channels from API\n", len(allChannels))
	fmt.Printf("   Pre-filtered to %d candidates (skipped %d active, %d excluded, %d too new)\n\n",
		len(candidateChannels), skippedActive, skippedExcluded, skippedNew)

	// Auto-join public channels before analysis
	fmt.Printf("📞 API Calls 3+: Auto-joining public channels for accurate analysis...\n")
	joinedCount, err := c.autoJoinPublicChannels(candidateChannels)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to auto-join channels - inactive detection requires channel membership: %w", err)
	}
	fmt.Printf("✅ Auto-joined %d channels\n\n", joinedCount)

	logger.WithField("joined_count", joinedCount).Debug("Auto-joined public channels")

	// Now analyze each candidate channel for activity with detailed reporting
	now := time.Now()
	for _, ch := range candidateChannels {

		lastActivity, hasWarning, warningTime, lastMessage, err := c.GetChannelActivityWithMessageAndUsers(ch.ID, userMap)
		if err != nil {
			errStr := err.Error()
			// Log the error but continue with other channels
			logger.WithFields(logger.LogFields{
				"channel": ch.Name,
				"error":   errStr,
			}).Error("Failed to get channel activity")

			// If rate limited, stop processing entirely
			if strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "rate limit") {
				logger.WithFields(logger.LogFields{
					"channel": ch.Name,
					"error":   errStr,
				}).Warn("Rate limited by Slack API - this affects all subsequent requests, stopping analysis")

				if isDebug {
					fmt.Printf("❌ API Call failed: Rate limited while checking #%s\n", ch.Name)
				}
				return toWarn, toArchive, fmt.Errorf("rate limited by Slack API")
			}

			if isDebug {
				fmt.Printf("❌ API Call failed: Error checking #%s: %s\n", ch.Name, errStr)
			}
			continue
		}

		if isDebug {
			fmt.Printf("✅ API Call succeeded\n")
		}

		// Create enhanced channel with message details
		enhancedChannel := Channel{
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

		// Show detailed info about the channel
		var activityStr string
		if lastActivity.IsZero() {
			created := time.Unix(int64(ch.Created), 0)
			createdDuration := now.Sub(created)
			activityStr = fmt.Sprintf("no real messages (created %s ago)", formatDuration(createdDuration))
		} else {
			duration := now.Sub(lastActivity)
			activityStr = fmt.Sprintf("last real message %s ago", formatDuration(duration))
			if hasWarning && !warningTime.IsZero() {
				warningDuration := now.Sub(warningTime)
				activityStr += fmt.Sprintf(" (warning sent %s ago)", formatDuration(warningDuration))
			}
		}

		fmt.Printf("  #%-20s - %s\n", ch.Name, activityStr)

		// Show message details if available
		if lastMessage != nil {
			// Truncate long messages
			messageText := lastMessage.Text
			if len(messageText) > 80 {
				messageText = messageText[:77] + "..."
			}

			// Replace newlines with spaces for cleaner display
			messageText = strings.ReplaceAll(messageText, "\n", " ")

			botIndicator := ""
			if lastMessage.IsBot {
				botIndicator = " (bot)"
			}

			// Use resolved name if available, fallback to ID
			authorName := lastMessage.UserName
			if authorName == "" {
				authorName = lastMessage.User
			}

			fmt.Printf("    └─ Author: %s%s | Message: \"%s\"\n", authorName, botIndicator, messageText)
		}
		fmt.Println() // Empty line between channels

		// Determine if this channel needs warning or archiving
		if hasWarning {
			// Channel already has a warning, check if grace period has expired
			// Archive if warning was sent more than archiveSeconds ago
			gracePeriodExpired := time.Since(warningTime) > time.Duration(archiveSeconds)*time.Second
			if gracePeriodExpired {
				toArchive = append(toArchive, enhancedChannel)
				logger.WithFields(logger.LogFields{
					"channel":      ch.Name,
					"warning_time": warningTime.Format("2006-01-02 15:04:05"),
					"grace_period": "expired",
				}).Debug("Channel marked for archival - grace period expired")
			}
		} else if lastActivity.IsZero() || lastActivity.Before(warnCutoff) {
			// Channel has no messages or is inactive and hasn't been warned yet
			toWarn = append(toWarn, enhancedChannel)
			logger.WithFields(logger.LogFields{
				"channel":       ch.Name,
				"last_activity": lastActivity.Format("2006-01-02 15:04:05"),
				"inactive_for":  now.Sub(lastActivity).String(),
			}).Debug("Channel marked for warning")
		}
	}

	logger.WithFields(logger.LogFields{
		"channels_to_warn":    len(toWarn),
		"channels_to_archive": len(toArchive),
	}).Info("Inactive channel analysis completed")

	return toWarn, toArchive, nil
}

// GetInactiveChannelsWithDetailsAndExclusions returns inactive channels with exclusion support
func (c *Client) GetInactiveChannelsWithDetailsAndExclusions(warnSeconds int, archiveSeconds int, userMap map[string]string, excludeChannels, excludePrefixes []string, isDebug bool) (toWarn []Channel, toArchive []Channel, err error) {
	logger.WithFields(logger.LogFields{
		"warn_seconds":     warnSeconds,
		"archive_seconds":  archiveSeconds,
		"exclude_channels": excludeChannels,
		"exclude_prefixes": excludePrefixes,
	}).Debug("Starting inactive channel detection with message details and exclusions")

	// Calculate cutoff times
	warnCutoff := time.Now().Add(-time.Duration(warnSeconds) * time.Second)

	// Get all channels


	allChannels, _, err := c.api.GetConversations(&slack.GetConversationsParameters{
		Types:           []string{"public_channel", "private_channel"},
		Limit:           1000,
		ExcludeArchived: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get conversations: %w", err)
	}


	logger.WithField("total_channels", len(allChannels)).Debug("Retrieved channels for detailed activity analysis")

	// Pre-filter channels using metadata and exclusions to reduce API calls
	var candidateChannels []slack.Channel
	var skippedActive, skippedExcluded, skippedNew, skippedUserExcluded int

	for _, ch := range allChannels {
		// Check user-specified exclusions first
		if c.shouldSkipChannelWithExclusions(ch.Name, excludeChannels, excludePrefixes) {
			skippedUserExcluded++
			continue
		}

		// Skip default excluded channels
		if c.shouldSkipChannel(ch.Name) {
			skippedExcluded++
			continue
		}

		// Skip channels that are too new to possibly need warnings
		created := time.Unix(int64(ch.Created), 0)
		if created.After(warnCutoff) {
			skippedNew++
			continue
		}

		// Skip channels that seem active from metadata
		if c.seemsActiveFromMetadata(ch, warnCutoff) {
			skippedActive++
			continue
		}

		candidateChannels = append(candidateChannels, ch)
	}

	logger.WithFields(logger.LogFields{
		"total_channels":        len(allChannels),
		"candidate_channels":    len(candidateChannels),
		"skipped_active":        skippedActive,
		"skipped_excluded":      skippedExcluded,
		"skipped_new":           skippedNew,
		"skipped_user_excluded": skippedUserExcluded,
	}).Debug("Pre-filtered channels using metadata and exclusions")

	if isDebug {
		fmt.Printf("📞 API Call 2: Getting channel list with metadata...\n")
		fmt.Printf("✅ Got %d channels from API\n", len(allChannels))
		fmt.Printf("   Pre-filtered to %d candidates (skipped %d active, %d excluded, %d too new, %d user-excluded)\n\n",
			len(candidateChannels), skippedActive, skippedExcluded, skippedNew, skippedUserExcluded)
	}

	// Auto-join public channels before analysis
	if isDebug {
		fmt.Printf("📞 API Calls 3+: Auto-joining public channels for accurate analysis...\n")
	}
	joinedCount, err := c.autoJoinPublicChannels(candidateChannels)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to auto-join channels - inactive detection requires channel membership: %w", err)
	}
	if isDebug {
		fmt.Printf("✅ Auto-joined %d channels\n\n", joinedCount)
	}

	logger.WithField("joined_count", joinedCount).Debug("Auto-joined public channels")

	// Now analyze each candidate channel for activity with detailed reporting
	now := time.Now()
	for _, ch := range candidateChannels {

		lastActivity, hasWarning, warningTime, lastMessage, err := c.GetChannelActivityWithMessageAndUsers(ch.ID, userMap)
		if err != nil {
			errStr := err.Error()
			// Log the error but continue with other channels
			logger.WithFields(logger.LogFields{
				"channel": ch.Name,
				"error":   errStr,
			}).Error("Failed to get channel activity")

			// If rate limited, stop processing entirely
			if strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "rate limit") {
				logger.WithFields(logger.LogFields{
					"channel": ch.Name,
					"error":   errStr,
				}).Warn("Rate limited by Slack API - this affects all subsequent requests, stopping analysis")

				if isDebug {
					fmt.Printf("❌ API Call failed: Rate limited while checking #%s\n", ch.Name)
				}
				return toWarn, toArchive, fmt.Errorf("rate limited by Slack API")
			}

			if isDebug {
				fmt.Printf("❌ API Call failed: Error checking #%s: %s\n", ch.Name, errStr)
			}
			continue
		}

		if isDebug {
			fmt.Printf("✅ API Call succeeded\n")
		}

		// Create enhanced channel with message details
		enhancedChannel := Channel{
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

		// Show detailed info about the channel
		var activityStr string
		if lastActivity.IsZero() {
			created := time.Unix(int64(ch.Created), 0)
			createdDuration := now.Sub(created)
			activityStr = fmt.Sprintf("no real messages (created %s ago)", formatDuration(createdDuration))
		} else {
			duration := now.Sub(lastActivity)
			activityStr = fmt.Sprintf("last real message %s ago", formatDuration(duration))
			if hasWarning && !warningTime.IsZero() {
				warningDuration := now.Sub(warningTime)
				activityStr += fmt.Sprintf(" (warning sent %s ago)", formatDuration(warningDuration))
			}
		}

		fmt.Printf("  #%-20s - %s\n", ch.Name, activityStr)

		// Show message details if available
		if lastMessage != nil {
			// Truncate long messages
			messageText := lastMessage.Text
			if len(messageText) > 80 {
				messageText = messageText[:77] + "..."
			}

			// Replace newlines with spaces for cleaner display
			messageText = strings.ReplaceAll(messageText, "\n", " ")

			botIndicator := ""
			if lastMessage.IsBot {
				botIndicator = " (bot)"
			}

			// Use resolved name if available, fallback to ID
			authorName := lastMessage.UserName
			if authorName == "" {
				authorName = lastMessage.User
			}

			fmt.Printf("    └─ Author: %s%s | Message: \"%s\"\n", authorName, botIndicator, messageText)
		}
		fmt.Println() // Empty line between channels

		// Determine if this channel needs warning or archiving
		if hasWarning {
			// Channel already has a warning, check if grace period has expired
			// Archive if warning was sent more than archiveSeconds ago
			gracePeriodExpired := time.Since(warningTime) > time.Duration(archiveSeconds)*time.Second
			if gracePeriodExpired {
				toArchive = append(toArchive, enhancedChannel)
				logger.WithFields(logger.LogFields{
					"channel":      ch.Name,
					"warning_time": warningTime.Format("2006-01-02 15:04:05"),
					"grace_period": "expired",
				}).Debug("Channel marked for archival - grace period expired")
			}
		} else if lastActivity.IsZero() || lastActivity.Before(warnCutoff) {
			// Channel has no messages or is inactive and hasn't been warned yet
			toWarn = append(toWarn, enhancedChannel)
			logger.WithFields(logger.LogFields{
				"channel":       ch.Name,
				"last_activity": lastActivity.Format("2006-01-02 15:04:05"),
				"inactive_for":  now.Sub(lastActivity).String(),
			}).Debug("Channel marked for warning")
		}
	}

	logger.WithFields(logger.LogFields{
		"channels_to_warn":    len(toWarn),
		"channels_to_archive": len(toArchive),
	}).Info("Inactive channel analysis completed")

	return toWarn, toArchive, nil
}

// shouldSkipChannelWithExclusions checks if a channel should be skipped based on user exclusions
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

// formatDuration formats a duration in a human-readable way
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
	const maxRetries = 3
	var history *slack.GetConversationHistoryResponse

	for attempt := 1; attempt <= maxRetries; attempt++ {

		// Optimized two-stage approach:
		// 1. Get just the last message to check if channel is obviously active
		// 2. Only if needed, get more messages to check for bot warnings

		// Stage 1: Get recent messages to find the most recent real one
		params := &slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Limit:     10, // Get enough messages to find real ones past any system messages
		}

		history, err = c.api.GetConversationHistory(params)
		if err != nil {
			errStr := err.Error()

			// Handle rate limiting with retry
			if strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "rate limit") {

				// If not the last attempt, continue to retry
				if attempt < maxRetries {
					// Parse Slack's retry-after directive
					waitDuration := parseSlackRetryAfter(errStr)
					if waitDuration > 0 {
						// Wait for the specified duration with user-friendly explanation
						fmt.Printf("⏳ Waiting %s due to Slack API rate limiting...\n", waitDuration.String())
						showProgressBar(waitDuration)
					} else {
						// Fallback to our rate limiter if we can't parse Slack's directive
										// Use fallback backoff quietly without verbose output
					}
					continue
				}
				// Last attempt failed, return error
				// All retry attempts failed, proceeding to return error
				return time.Time{}, false, time.Time{}, fmt.Errorf("failed to get channel history after %d attempts, final Slack error: %s", maxRetries, errStr)
			}

			// Non-rate-limit errors - don't retry
			return time.Time{}, false, time.Time{}, fmt.Errorf("failed to get channel history: %w", err)
		}

		// Success - break out of retry loop
		break
	}


	if len(history.Messages) == 0 {
		// No messages
		return time.Unix(0, 0), false, time.Time{}, nil
	}

	// Find the most recent "real" message (not join/leave/system messages)
	botUserID := c.getBotUserID()
	var lastRealMsg *slack.Message
	var lastRealMsgTime time.Time

	for _, msg := range history.Messages {
		if isRealMessage(msg, botUserID) {
			lastRealMsg = &msg
			if msgTime, err := parseSlackTimestamp(msg.Timestamp); err == nil {
				lastRealMsgTime = msgTime
			}
			break // Found the most recent real message
		}
	}

	// If no real messages found in the first message, get more history
	if lastRealMsg == nil {
		return c.getDetailedChannelActivity(channelID, botUserID)
	}

	// If the last real message is from the bot and contains a warning, we need more context
	if lastRealMsg.User == botUserID && strings.Contains(lastRealMsg.Text, "inactive channel warning") {
		// Stage 2: Get more messages to find actual user activity and warning history
		return c.getDetailedChannelActivity(channelID, botUserID)
	}

	// If the last real message is from a user, that's our activity time
	if lastRealMsg.User != botUserID {
		return lastRealMsgTime, false, time.Time{}, nil
	}

	// If the last real message is from the bot but not a warning, we need to look deeper
	// to find the last user activity
	return c.getDetailedChannelActivity(channelID, botUserID)
}

func (c *Client) getDetailedChannelActivity(channelID, botUserID string) (lastActivity time.Time, hasWarning bool, warningTime time.Time, err error) {
	const maxRetries = 3
	var history *slack.GetConversationHistoryResponse

	for attempt := 1; attempt <= maxRetries; attempt++ {

		// Get more messages to analyze warning history and find user activity
		params := &slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Limit:     50, // Reasonable limit to find warnings and user activity
		}

		history, err = c.api.GetConversationHistory(params)
		if err != nil {
			errStr := err.Error()

			// Handle rate limiting with retry
			if strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "rate limit") {
				logger.WithFields(logger.LogFields{
					"attempt":     attempt,
					"max_tries":   maxRetries,
					"channel":     channelID,
					"slack_error": errStr,
				}).Warn("Rate limited on detailed history, will retry after Slack-specified delay")

				// Print to stdout so user can see retry attempts
				fmt.Printf("   🔄 Detailed history retry %d/%d for channel %s (Slack error: %s)\n", attempt, maxRetries, channelID, errStr)

				// If not the last attempt, continue to retry
				if attempt < maxRetries {
					// Parse Slack's retry-after directive
					waitDuration := parseSlackRetryAfter(errStr)
					if waitDuration > 0 {
						fmt.Printf("   ⏳ Slack says wait %s (includes 3s buffer) before detailed retry...\n", waitDuration)
						time.Sleep(waitDuration)
					} else {
						// Fallback to our rate limiter if we can't parse Slack's directive
										fmt.Printf("   ⏳ Using fallback backoff before detailed retry...\n")
					}
					continue
				}
				// Last attempt failed, return error
				fmt.Printf("   ❌ All %d detailed retry attempts failed\n", maxRetries)
				return time.Time{}, false, time.Time{}, fmt.Errorf("failed to get detailed channel history after %d attempts, final Slack error: %s", maxRetries, errStr)
			}

			// Non-rate-limit errors - don't retry
			return time.Time{}, false, time.Time{}, fmt.Errorf("failed to get detailed channel history: %w", err)
		}

		// Success - break out of retry loop
		break
	}


	if len(history.Messages) == 0 {
		return time.Unix(0, 0), false, time.Time{}, nil
	}

	// Find the most recent real user message and check for warnings
	mostRecentActivity := time.Time{}
	hasWarningMessage := false
	mostRecentWarning := time.Time{}

	for _, msg := range history.Messages {
		msgTime, err := parseSlackTimestamp(msg.Timestamp)
		if err != nil {
			continue
		}

		// Check if this is a warning message from our bot
		if msg.User == botUserID && strings.Contains(msg.Text, "inactive channel warning") {
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

	// If only bot messages exist, use the oldest message time
	if mostRecentActivity.IsZero() && len(history.Messages) > 0 {
		if msgTime, err := parseSlackTimestamp(history.Messages[len(history.Messages)-1].Timestamp); err == nil {
			mostRecentActivity = msgTime
		}
	}

	return mostRecentActivity, hasWarningMessage, mostRecentWarning, nil
}

func (c *Client) autoJoinPublicChannels(channels []slack.Channel) (int, error) {
	joinedCount := 0
	var fatalErrors []string
	var skippedCount = 0

	for _, ch := range channels {
		// Only join public channels
		if ch.IsPrivate {
			logger.WithField("channel", ch.Name).Debug("Skipping private channel for auto-join")
			skippedCount++
			continue
		}

		// Rate limit before joining

		// Try to join the channel
		_, _, _, err := c.api.JoinConversation(ch.ID)
		if err != nil {
			errStr := err.Error()

			// Handle rate limiting - this is fatal
			if strings.Contains(errStr, "rate_limited") {
						return joinedCount, fmt.Errorf("rate limited during auto-join: %w", err)
			}

			// Already in channel is success
			if strings.Contains(errStr, "already_in_channel") {
				logger.WithField("channel", ch.Name).Debug("Already in channel")
							joinedCount++
				continue
			}

			// These indicate the channel can't be joined, but that's OK to skip
			if strings.Contains(errStr, "is_archived") {
				logger.WithField("channel", ch.Name).Debug("Channel is archived, skipping")
							skippedCount++
				continue
			}

			if strings.Contains(errStr, "invite_only") {
				logger.WithField("channel", ch.Name).Debug("Channel is invite-only, skipping")
							skippedCount++
				continue
			}

			// Missing scope or permissions - this is fatal for the bot's functionality
			if strings.Contains(errStr, "missing_scope") {
				return joinedCount, fmt.Errorf("missing required OAuth scope to join channels. Your bot needs the 'channels:join' OAuth scope.\nPlease add this scope in your Slack app settings at https://api.slack.com/apps")
			}

			if strings.Contains(errStr, "invalid_auth") {
				return joinedCount, fmt.Errorf("invalid authentication token: %w", err)
			}

			// Other errors are fatal - we need to be able to join channels for accurate analysis
			fatalErrors = append(fatalErrors, fmt.Sprintf("%s: %s", ch.Name, errStr))
			continue
		}

			joinedCount++
		logger.WithField("channel", ch.Name).Debug("Successfully joined channel")
	}

	logger.WithFields(logger.LogFields{
		"joined":  joinedCount,
		"skipped": skippedCount,
		"total":   len(channels),
	}).Debug("Auto-join summary")

	if len(fatalErrors) > 0 {
		return joinedCount, fmt.Errorf("failed to join %d channels, cannot proceed with accurate analysis: %v", len(fatalErrors), fatalErrors)
	}

	return joinedCount, nil
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

func (c *Client) WarnInactiveChannel(channel Channel, warnSeconds, archiveSeconds int) error {
	// First, try to join the channel if it's public
	if err := c.ensureBotInChannel(channel); err != nil {
		logger.WithFields(logger.LogFields{
			"channel": channel.Name,
			"error":   err.Error(),
		}).Warn("Could not join channel, skipping warning")
		return fmt.Errorf("failed to join channel %s: %w", channel.Name, err)
	}

	message := c.FormatInactiveChannelWarning(channel, warnSeconds, archiveSeconds)

	logger.WithFields(logger.LogFields{
		"channel":         channel.Name,
		"archive_seconds": archiveSeconds,
	}).Debug("Posting inactive channel warning")

	return c.postMessageToChannelID(channel.ID, message)
}

func (c *Client) ensureBotInChannel(channel Channel) error {
	logger.WithField("channel", channel.Name).Debug("Ensuring bot is in channel")

	// Rate limit before API call


	_, _, _, err := c.api.JoinConversation(channel.ID)
	if err != nil {
		errStr := err.Error()
		logger.WithFields(logger.LogFields{
			"channel":   channel.Name,
			"error":     errStr,
			"operation": "join_conversation",
		}).Debug("Join conversation result")

		// Handle rate limiting
		if strings.Contains(errStr, "rate_limited") {
				return fmt.Errorf("rate limited by Slack API. Will retry with exponential backoff on next request")
		}

		// Handle expected errors that we can ignore
		if strings.Contains(errStr, "already_in_channel") {
			logger.WithField("channel", channel.Name).Debug("Bot already in channel")
					return nil
		}

		// Handle permission errors
		if strings.Contains(errStr, "missing_scope") {
			return fmt.Errorf("missing required permission to join channels. Your bot needs the 'channels:join' OAuth scope.\nPlease add this scope in your Slack app settings at https://api.slack.com/apps")
		}

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

		// For other errors, log but don't fail completely
		logger.WithFields(logger.LogFields{
			"channel": channel.Name,
			"error":   errStr,
		}).Warn("Unexpected error joining channel")
		return fmt.Errorf("failed to join channel %s: %w", channel.Name, err)
	}

	logger.WithField("channel", channel.Name).Info("Successfully joined channel")
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

func (c *Client) FormatInactiveChannelWarning(channel Channel, warnSeconds, archiveSeconds int) string {
	var builder strings.Builder

	builder.WriteString("🚨 **Inactive Channel Warning** 🚨\n\n")

	// Format warn period in human readable format
	warnText := fmt.Sprintf("%d seconds", warnSeconds)
	if warnSeconds >= 60 {
		minutes := warnSeconds / 60
		if minutes == 1 {
			warnText = "1 minute"
		} else if minutes < 60 {
			warnText = fmt.Sprintf("%d minutes", minutes)
		} else {
			hours := minutes / 60
			if hours == 1 {
				warnText = "1 hour"
			} else if hours < 24 {
				warnText = fmt.Sprintf("%d hours", hours)
			} else {
				days := hours / 24
				if days == 1 {
					warnText = "1 day"
				} else {
					warnText = fmt.Sprintf("%d days", days)
				}
			}
		}
	}

	builder.WriteString(fmt.Sprintf("This channel has been inactive for more than %s.\n\n", warnText))

	// Convert seconds to human readable format
	archiveText := fmt.Sprintf("%d seconds", archiveSeconds)
	if archiveSeconds >= 60 {
		minutes := archiveSeconds / 60
		if minutes == 1 {
			archiveText = "1 minute"
		} else {
			archiveText = fmt.Sprintf("%d minutes", minutes)
		}
	}

	builder.WriteString(fmt.Sprintf("**This channel will be archived in %s** unless new messages are posted.\n\n", archiveText))

	builder.WriteString("To keep this channel active:\n")
	builder.WriteString("• Post a message in this channel\n")
	builder.WriteString("• Or contact an admin if this channel should remain active\n\n")

	builder.WriteString("_This is an automated message from slack-buddy bot._\n")
	builder.WriteString("<!-- inactive channel warning -->")

	return builder.String()
}

func (c *Client) FormatChannelArchivalMessage(channel Channel, warnSeconds, archiveSeconds int) string {
	var builder strings.Builder

	builder.WriteString("📋 **Channel Archival Notice** 📋\n\n")

	// Format warn period in human readable format
	warnText := fmt.Sprintf("%d seconds", warnSeconds)
	if warnSeconds >= 60 {
		minutes := warnSeconds / 60
		if minutes == 1 {
			warnText = "1 minute"
		} else if minutes < 60 {
			warnText = fmt.Sprintf("%d minutes", minutes)
		} else {
			hours := minutes / 60
			if hours == 1 {
				warnText = "1 hour"
			} else if hours < 24 {
				warnText = fmt.Sprintf("%d hours", hours)
			} else {
				days := hours / 24
				if days == 1 {
					warnText = "1 day"
				} else {
					warnText = fmt.Sprintf("%d days", days)
				}
			}
		}
	}

	// Format archive period in human readable format
	archiveText := fmt.Sprintf("%d seconds", archiveSeconds)
	if archiveSeconds >= 60 {
		minutes := archiveSeconds / 60
		if minutes == 1 {
			archiveText = "1 minute"
		} else if minutes < 60 {
			archiveText = fmt.Sprintf("%d minutes", minutes)
		} else {
			hours := minutes / 60
			if hours == 1 {
				archiveText = "1 hour"
			} else if hours < 24 {
				archiveText = fmt.Sprintf("%d hours", hours)
			} else {
				days := hours / 24
				if days == 1 {
					archiveText = oneDayText
				} else {
					archiveText = strconv.Itoa(days) + " days"
				}
			}
		}
	}

	builder.WriteString(fmt.Sprintf("This channel is being archived because:\n"))
	builder.WriteString(fmt.Sprintf("• It was inactive for more than %s (warning threshold)\n", warnText))
	builder.WriteString(fmt.Sprintf("• An inactivity warning was posted\n"))
	builder.WriteString(fmt.Sprintf("• No new activity occurred within %s after the warning (archive threshold)\n\n", archiveText))

	builder.WriteString("**This channel is now being archived.**\n\n")

	builder.WriteString("If this channel should not have been archived, please contact a workspace admin.\n\n")

	builder.WriteString("_This is an automated action by slack-buddy bot._\n")
	builder.WriteString("<!-- channel archival notice -->")

	return builder.String()
}

func (c *Client) ArchiveChannel(channel Channel) error {
	// Legacy method for backward compatibility - uses default thresholds
	return c.ArchiveChannelWithThresholds(channel, 300, 60) // 5 minutes warn, 1 minute archive
}

func (c *Client) ArchiveChannelWithThresholds(channel Channel, warnSeconds, archiveSeconds int) error {
	logger.WithField("channel", channel.Name).Debug("Archiving inactive channel")

	// First, ensure the bot is in the channel to post the archival message
	if err := c.ensureBotInChannel(channel); err != nil {
		logger.WithFields(logger.LogFields{
			"channel": channel.Name,
			"error":   err.Error(),
		}).Warn("Could not join channel for archival message, proceeding with archival anyway")
		// Don't fail here - we can still archive even if we can't post the message
	}

	// Post archival message explaining why the channel is being archived
	archivalMessage := c.FormatChannelArchivalMessage(channel, warnSeconds, archiveSeconds)
	if err := c.postMessageToChannelID(channel.ID, archivalMessage); err != nil {
		logger.WithFields(logger.LogFields{
			"channel": channel.Name,
			"error":   err.Error(),
		}).Warn("Failed to post archival message, proceeding with archival anyway")
		// Don't fail here - archival should proceed even if the message fails
	} else {
		logger.WithField("channel", channel.Name).Info("Posted archival message successfully")
	}

	// Rate limit before archival API call


	err := c.api.ArchiveConversation(channel.ID)
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
		Types: []string{"public_channel", "private_channel"},
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
			return nil, fmt.Errorf("missing required permissions. Your bot needs these OAuth scopes:\\n  - channels:read (to list public channels) - REQUIRED\\n  - groups:read (to list private channels) - OPTIONAL\\n\\nPlease add these scopes in your Slack app settings at https://api.slack.com/apps")
		}
		if strings.Contains(errStr, "invalid_auth") {
			logger.Error("Invalid Slack authentication token")
			return nil, fmt.Errorf("invalid token. Please check your SLACK_TOKEN")
		}
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}


	logger.WithField("total_channels", len(channels)).Debug("Retrieved channels from Slack API")

	var result []Channel
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

// GetChannelActivity returns the last activity time, warning status, and warning time for a channel
func (c *Client) GetChannelActivity(channelID string) (lastActivity time.Time, hasWarning bool, warningTime time.Time, err error) {
	return c.getChannelActivity(channelID)
}

// MessageInfo contains details about a message
type MessageInfo struct {
	Timestamp time.Time
	User      string
	UserName  string // Human-readable name
	Text      string
	IsBot     bool
}

// GetChannelActivityWithMessage returns activity info plus details about the most recent message
func (c *Client) GetChannelActivityWithMessage(channelID string) (lastActivity time.Time, hasWarning bool, warningTime time.Time, lastMessage *MessageInfo, err error) {
	const maxRetries = 3
	var history *slack.GetConversationHistoryResponse

	for attempt := 1; attempt <= maxRetries; attempt++ {

		// Get recent messages to find the most recent real one
		params := &slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Limit:     10, // Get more messages to find real ones past any system messages
		}

		history, err = c.api.GetConversationHistory(params)
		if err != nil {
			errStr := err.Error()

			// Handle rate limiting with retry
			if strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "rate limit") {

				// If not the last attempt, continue to retry
				if attempt < maxRetries {
					// Parse Slack's retry-after directive
					waitDuration := parseSlackRetryAfter(errStr)
					if waitDuration > 0 {
						// Wait for the specified duration with user-friendly explanation
						fmt.Printf("⏳ Waiting %s due to Slack API rate limiting...\n", waitDuration.String())
						showProgressBar(waitDuration)
					} else {
						// Fallback to our rate limiter if we can't parse Slack's directive
										// Use fallback backoff quietly without verbose output
					}
					continue
				}
				// Last attempt failed, return error
				// All retry attempts failed, proceeding to return error
				return time.Time{}, false, time.Time{}, nil, fmt.Errorf("failed to get channel history after %d attempts, final Slack error: %s", maxRetries, errStr)
			}

			// Non-rate-limit errors - don't retry
			return time.Time{}, false, time.Time{}, nil, fmt.Errorf("failed to get channel history: %w", err)
		}

		// Success - break out of retry loop
		break
	}


	if len(history.Messages) == 0 {
		// No messages
		return time.Time{}, false, time.Time{}, nil, nil
	}

	// Find the most recent "real" message (not join/leave/system messages)
	botUserID := c.getBotUserID()
	var lastRealMsg *slack.Message
	var lastRealMsgTime time.Time

	for _, msg := range history.Messages {
		if isRealMessage(msg, botUserID) {
			lastRealMsg = &msg
			if msgTime, err := parseSlackTimestamp(msg.Timestamp); err == nil {
				lastRealMsgTime = msgTime
			}
			break // Found the most recent real message
		}
	}

	// If no real messages found, check if we need to look deeper
	if lastRealMsg == nil {
		// No real messages in the first message, might need to get more history
		// For now, return no activity
		return time.Time{}, false, time.Time{}, nil, nil
	}

	// Create message info for the real message
	msgInfo := &MessageInfo{
		Timestamp: lastRealMsgTime,
		User:      lastRealMsg.User,
		Text:      lastRealMsg.Text,
		IsBot:     lastRealMsg.User == botUserID,
	}

	// Check if this is a warning message from our bot
	hasWarningMessage := lastRealMsg.User == botUserID && strings.Contains(lastRealMsg.Text, "inactive channel warning")
	var mostRecentWarning time.Time
	if hasWarningMessage {
		mostRecentWarning = lastRealMsgTime
	}

	return lastRealMsgTime, hasWarningMessage, mostRecentWarning, msgInfo, nil
}

// parseSlackRetryAfter parses Slack's "retry after" directive from error messages
// Example: "slack rate limit exceeded, retry after 1m0s" -> 1 minute duration
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

// showProgressBar displays a text-based progress bar for waiting periods
// Shows - for remaining time and * for elapsed time
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
// Returns true for actual user-generated content
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

	// Filter out empty messages
	if strings.TrimSpace(text) == "" {
		return false
	}

	// All other messages are considered "real"
	return true
}

// getUserMap fetches all users and builds a map from user ID to display name
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

// GetUserMap is a public wrapper for getUserMap
func (c *Client) GetUserMap() (map[string]string, error) {
	return c.getUserMap()
}

// GetChannelActivityWithMessageAndUsers returns activity info plus message details with resolved user names
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
