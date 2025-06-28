package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astrostl/slack-butler/pkg/logger"
	"github.com/astrostl/slack-butler/pkg/slack"

	slackapi "github.com/slack-go/slack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var channelsCmd = &cobra.Command{
	Use:          "channels",
	Short:        "Manage channels in your Slack workspace",
	Long:         `Commands for managing and monitoring channels in your Slack workspace.`,
	SilenceUsage: true, // Don't show usage on errors
}

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect new channels created in a time period",
	Long: `Detect new channels created during a specified time period and optionally announce them to another channel.

When --announce-to is specified, shows a preview of the announcement message.
Use --commit with --announce-to to actually post messages (default is dry run mode).`,
	SilenceUsage: true, // Don't show usage on errors
	RunE:         runDetect,
}

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Manage inactive channel archival with warnings",
	Long: `Detect inactive channels, warn them about upcoming archival, and archive channels that remain inactive after the grace period.

This command operates in a stateless manner by using channel message history to determine activity and warning status.
Warning messages are used as state markers to track the grace period.

The bot will automatically join public channels to send warning messages. Required OAuth scopes:
- channels:read (to list channels) - REQUIRED
- channels:join (to join public channels) - REQUIRED  
- chat:write (to post warnings) - REQUIRED
- channels:manage (to archive channels) - REQUIRED

Use --commit to actually warn and archive channels (default is dry run mode).

NOTE: Archive timing is configured in days with decimal precision for flexible control (e.g., 0.0003 days = ~26 seconds).`,
	SilenceUsage: true, // Don't show usage on errors
	RunE:         runArchive,
}

var highlightCmd = &cobra.Command{
	Use:   "highlight",
	Short: "Highlight random channels to encourage discovery",
	Long: `Randomly select and highlight channels from your workspace to encourage participation and discovery.

This command helps teams discover channels they might not know about by showcasing random active channels.
Useful for promoting engagement and helping members find relevant conversations.

When --announce-to is specified, shows a preview of the highlight message.
Use --commit with --announce-to to actually post messages (default is dry run mode).`,
	SilenceUsage: true, // Don't show usage on errors
	RunE:         runHighlight,
}

var (
	since           string
	announceTo      string
	commit          bool
	warnDays        float64
	archiveDays     float64
	excludeChannels string
	excludePrefixes string
	count           int
)

func init() {
	rootCmd.AddCommand(channelsCmd)
	channelsCmd.AddCommand(detectCmd)
	channelsCmd.AddCommand(archiveCmd)
	channelsCmd.AddCommand(highlightCmd)

	detectCmd.Flags().StringVar(&since, "since", "8", "Number of days to look back (e.g., 1, 7, 30)")
	detectCmd.Flags().StringVar(&announceTo, "announce-to", "", "Channel to announce new channels to (e.g., #general). Required when using --commit")
	detectCmd.Flags().BoolVar(&commit, "commit", false, "Actually post messages (default is dry run mode)")

	archiveCmd.Flags().Float64Var(&warnDays, "warn-days", 30.0, "Number of days of inactivity before warning (supports decimal precision, e.g., 0.0003)")
	archiveCmd.Flags().Float64Var(&archiveDays, "archive-days", 30.0, "Number of days after warning (with no new activity) before archiving (supports decimal precision, e.g., 0.0003)")
	archiveCmd.Flags().BoolVar(&commit, "commit", false, "Actually warn and archive channels (default is dry run mode)")
	archiveCmd.Flags().StringVar(&excludeChannels, "exclude-channels", "", "Comma-separated list of channel names to exclude (with or without # prefix, e.g., 'general,random,#important')")
	archiveCmd.Flags().StringVar(&excludePrefixes, "exclude-prefixes", "", "Comma-separated list of channel prefixes to exclude (with or without # prefix, e.g., 'prod-,#temp-,admin')")

	highlightCmd.Flags().IntVar(&count, "count", 3, "Number of random channels to highlight (e.g., 1, 3, 5)")
	highlightCmd.Flags().StringVar(&announceTo, "announce-to", "", "Channel to announce highlights to (e.g., #general). Required when using --commit")
	highlightCmd.Flags().BoolVar(&commit, "commit", false, "Actually post messages (default is dry run mode)")
}

func runDetect(cmd *cobra.Command, args []string) error {
	token := viper.GetString("token")
	if token == "" {
		return fmt.Errorf("slack token is required. Set SLACK_TOKEN environment variable or use --token flag")
	}

	// announce-to is mandatory when committing changes
	if announceTo == "" && commit {
		return fmt.Errorf("--announce-to is required when using --commit")
	}

	days, err := strconv.ParseFloat(since, 64)
	if err != nil {
		return fmt.Errorf("invalid days format '%s': must be a number (e.g., 1, 7, 30)", since)
	}

	if days < 0 {
		return fmt.Errorf("days must be positive, got %g", days)
	}

	duration := time.Duration(days*24) * time.Hour
	cutoffTime := time.Now().Add(-duration)

	client, err := slack.NewClient(token)
	if err != nil {
		return fmt.Errorf("failed to create Slack client: %w", err)
	}

	// Validate that the announce-to channel exists (if specified)
	if announceTo != "" {
		_, err = client.ResolveChannelNameToID(announceTo)
		if err != nil {
			return fmt.Errorf("announce-to channel '%s' not found: %w", announceTo, err)
		}
	}

	return runDetectWithClient(client, cutoffTime, announceTo, !commit)
}

func displayNewChannels(newChannels []slack.Channel) {
	channelList := make([]string, len(newChannels))
	for i, channel := range newChannels {
		channelList[i] = "#" + channel.Name
	}
	fmt.Printf("New channels found (%d): %s\n\n", len(newChannels), strings.Join(channelList, ", "))
}

func extractChannelNames(channels []slack.Channel) []string {
	channelNames := make([]string, len(channels))
	for i, channel := range channels {
		channelNames[i] = channel.Name
	}
	return channelNames
}

func runDetectWithClient(client *slack.Client, cutoffTime time.Time, announceChannel string, isDryRun bool) error {
	newChannels, allChannels, err := client.GetNewChannelsWithAllChannels(cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to get new channels: %w", err)
	}

	if len(newChannels) == 0 {
		return nil
	}

	displayNewChannels(newChannels)

	if announceChannel != "" {
		return handleAnnouncement(client, newChannels, allChannels, cutoffTime, announceChannel, isDryRun)
	}

	return handleDryRunWithoutChannel(client, newChannels, cutoffTime, isDryRun)
}

func handleAnnouncement(client *slack.Client, newChannels []slack.Channel, allChannels []slackapi.Channel, cutoffTime time.Time, announceChannel string, isDryRun bool) error {
	message := client.FormatNewChannelAnnouncement(newChannels, cutoffTime)
	channelNames := extractChannelNames(newChannels)

	fmt.Printf("Checking for duplicate announcements in %s...\n\n", announceChannel)
	isDuplicate, skippedChannels, err := client.CheckForDuplicateAnnouncementWithDetailsAndChannels(announceChannel, message, channelNames, cutoffTime, allChannels)
	if err != nil {
		logger.WithFields(logger.LogFields{
			"channel": announceChannel,
			"error":   err.Error(),
		}).Warn("Failed to check for duplicate announcements, proceeding with post")
	}

	var finalMessage string
	var channelsToAnnounce []slack.Channel

	if isDuplicate {
		newChannelsToAnnounce := filterSkippedChannels(channelNames, skippedChannels)

		if len(skippedChannels) > 0 {
			fmt.Printf("Channels already announced (skipped): %s\n", strings.Join(skippedChannels, ", "))
		}

		if len(newChannelsToAnnounce) == 0 {
			fmt.Printf("All channels already announced, skipping announcement to %s\n", announceChannel)
			return nil
		}

		channelsToAnnounce = filterChannelsByNames(newChannels, newChannelsToAnnounce)
		finalMessage = client.FormatNewChannelAnnouncement(channelsToAnnounce, cutoffTime)

		if len(newChannelsToAnnounce) > 0 {
			displayAnnouncingChannels(newChannelsToAnnounce, len(skippedChannels))
		}
	} else {
		channelsToAnnounce = newChannels
		finalMessage = message
	}

	return postOrPreviewAnnouncement(client, announceChannel, finalMessage, channelsToAnnounce, cutoffTime, isDryRun)
}

func filterSkippedChannels(channelNames []string, skippedChannels []string) []string {
	var newChannelsToAnnounce []string
	for _, channelName := range channelNames {
		skipped := false
		for _, skippedChannel := range skippedChannels {
			if channelName == skippedChannel {
				skipped = true
				break
			}
		}
		if !skipped {
			newChannelsToAnnounce = append(newChannelsToAnnounce, channelName)
		}
	}
	return newChannelsToAnnounce
}

func filterChannelsByNames(channels []slack.Channel, names []string) []slack.Channel {
	var filtered []slack.Channel
	for _, channel := range channels {
		for _, name := range names {
			if channel.Name == name {
				filtered = append(filtered, channel)
				break
			}
		}
	}
	return filtered
}

func displayAnnouncingChannels(channelNames []string, skippedCount int) {
	announcingList := make([]string, 0, len(channelNames))
	for _, channelName := range channelNames {
		announcingList = append(announcingList, "#"+channelName)
	}
	fmt.Printf("Announcing channels: %s (skipped %d already announced)\n", strings.Join(announcingList, ", "), skippedCount)
}

func postOrPreviewAnnouncement(client *slack.Client, announceChannel, finalMessage string, channelsToAnnounce []slack.Channel, cutoffTime time.Time, isDryRun bool) error {
	if isDryRun {
		dryRunMessage := client.FormatNewChannelAnnouncementDryRun(channelsToAnnounce, cutoffTime)
		fmt.Printf("\n--- DRY RUN ---\n")
		fmt.Printf("Would announce to channel: %s\n", announceChannel)
		fmt.Printf("Message content:\n%s\n", dryRunMessage)
		fmt.Printf("--- END DRY RUN ---\n")
		fmt.Printf("\nTo actually post this announcement, add --commit to your command\n")
	} else {
		if err := client.PostMessage(announceChannel, finalMessage); err != nil {
			logger.WithFields(logger.LogFields{
				"channel": announceChannel,
				"error":   err.Error(),
			}).Error("Failed to post announcement")
			return fmt.Errorf("failed to post announcement to %s: %w", announceChannel, err)
		}
		fmt.Printf("Announcement posted to %s\n", announceChannel)
	}
	return nil
}

func handleDryRunWithoutChannel(client *slack.Client, newChannels []slack.Channel, cutoffTime time.Time, isDryRun bool) error {
	if isDryRun {
		message := client.FormatNewChannelAnnouncementDryRun(newChannels, cutoffTime)
		fmt.Printf("\n--- DRY RUN ---\n")
		fmt.Printf("Announcement message dry run (use --announce-to to specify target):\n%s\n", message)
		fmt.Printf("--- END DRY RUN ---\n")
		fmt.Printf("\nTo actually post announcements, add --commit to your command\n")
	}
	return nil
}

func runArchive(cmd *cobra.Command, args []string) error {
	token := viper.GetString("token")
	if token == "" {
		return fmt.Errorf("slack token is required. Set SLACK_TOKEN environment variable or use --token flag")
	}

	if warnDays <= 0 {
		return fmt.Errorf("warn-days must be positive, got %g", warnDays)
	}

	if archiveDays <= 0 {
		return fmt.Errorf("archive-days must be positive, got %g", archiveDays)
	}

	// Convert days to seconds for internal use
	warnSeconds := int(warnDays * 24 * 60 * 60)
	archiveSeconds := int(archiveDays * 24 * 60 * 60)

	client, err := slack.NewClient(token)
	if err != nil {
		return fmt.Errorf("failed to create Slack client: %w", err)
	}

	return runArchiveWithClient(client, warnSeconds, archiveSeconds, !commit, excludeChannels, excludePrefixes)
}

func runArchiveWithClient(client *slack.Client, warnSeconds, archiveSeconds int, isDryRun bool, excludeChannels, excludePrefixes string) error {
	isDebug := viper.GetBool("debug")

	if isDebug {
		logger.WithFields(logger.LogFields{
			"warn_seconds":    warnSeconds,
			"archive_seconds": archiveSeconds,
			"dry_run_mode":    isDryRun,
		}).Info("Starting inactive channel analysis")
	}

	fmt.Printf("🔍 Analyzing inactive channels...\n\n")

	// Get user map for name resolution
	userMap, err := getUserMapWithErrorHandling(client, isDebug)
	if err != nil {
		return err
	}

	// Parse and display exclusion lists
	excludeChannelsList, excludePrefixesList := parseExclusionLists(excludeChannels, excludePrefixes)
	displayExclusionInfo(excludeChannelsList, excludePrefixesList)

	// Analyze inactive channels
	toWarn, toArchive, err := getInactiveChannelsWithErrorHandling(client, warnSeconds, archiveSeconds, userMap, excludeChannelsList, excludePrefixesList, isDebug)
	if err != nil {
		return err
	}

	// Report findings
	fmt.Printf("Inactive Channel Analysis Results:\n")
	fmt.Printf("  Channels to warn: %d\n", len(toWarn))
	fmt.Printf("  Channels to archive: %d\n", len(toArchive))
	fmt.Println()

	// Process warnings
	if len(toWarn) > 0 {
		processWarnings(client, toWarn, warnSeconds, archiveSeconds, isDryRun)
	}

	// Process archival
	if len(toArchive) > 0 {
		processArchival(client, toArchive, warnSeconds, archiveSeconds, isDryRun)
	}

	if len(toWarn) == 0 && len(toArchive) == 0 {
		fmt.Printf("No inactive channels found. All channels are active or already processed.\n")
	}

	return nil
}

// getUserMapWithErrorHandling gets user map with proper error handling and logging.
func getUserMapWithErrorHandling(client *slack.Client, isDebug bool) (map[string]string, error) {
	if isDebug {
		fmt.Printf("📞 API Call 1: Getting user list for name resolution...\n")
	}
	userMap, err := client.GetUserMap()
	if err != nil {
		if strings.Contains(err.Error(), "rate_limited") || strings.Contains(err.Error(), "rate limit") {
			fmt.Printf("⚠️  Slack API rate limit exceeded on user list.\n")
			fmt.Printf("   The system should have done backoff.\n")
			return nil, fmt.Errorf("rate limited by Slack API")
		}
		if strings.Contains(err.Error(), "missing_scope") || strings.Contains(err.Error(), "users:read") {
			fmt.Printf("❌ Missing required OAuth scope 'users:read'\n")
			fmt.Printf("   This scope is needed to resolve user names for message authors.\n")
			fmt.Printf("   Add 'users:read' scope in your Slack app settings at https://api.slack.com/apps\n")
			return nil, fmt.Errorf("missing required OAuth scope 'users:read'")
		}
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	if isDebug {
		fmt.Printf("✅ Got %d users from API\n\n", len(userMap))
	}
	return userMap, nil
}

// parseExclusionLists parses comma-separated exclusion lists and removes # prefixes.
func parseExclusionLists(excludeChannels, excludePrefixes string) ([]string, []string) {
	var excludeChannelsList []string
	var excludePrefixesList []string

	if excludeChannels != "" {
		for _, channel := range strings.Split(excludeChannels, ",") {
			channel = strings.TrimSpace(channel)
			// Remove # prefix if present
			channel = strings.TrimPrefix(channel, "#")
			if channel != "" {
				excludeChannelsList = append(excludeChannelsList, channel)
			}
		}
	}

	if excludePrefixes != "" {
		for _, prefix := range strings.Split(excludePrefixes, ",") {
			prefix = strings.TrimSpace(prefix)
			// Remove # prefix if present
			prefix = strings.TrimPrefix(prefix, "#")
			if prefix != "" {
				excludePrefixesList = append(excludePrefixesList, prefix)
			}
		}
	}

	return excludeChannelsList, excludePrefixesList
}

// displayExclusionInfo shows configured exclusions to the user.
func displayExclusionInfo(excludeChannelsList, excludePrefixesList []string) {
	if len(excludeChannelsList) > 0 || len(excludePrefixesList) > 0 {
		fmt.Printf("📋 Channel exclusions configured:\n")
		if len(excludeChannelsList) > 0 {
			fmt.Printf("   Excluded channels: %s\n", strings.Join(excludeChannelsList, ", "))
		}
		if len(excludePrefixesList) > 0 {
			fmt.Printf("   Excluded prefixes: %s\n", strings.Join(excludePrefixesList, ", "))
		}
		fmt.Println()
	}
}

// getInactiveChannelsWithErrorHandling analyzes inactive channels with proper error handling.
func getInactiveChannelsWithErrorHandling(client *slack.Client, warnSeconds, archiveSeconds int, userMap map[string]string, excludeChannelsList, excludePrefixesList []string, isDebug bool) ([]slack.Channel, []slack.Channel, error) {
	toWarn, toArchive, err := client.GetInactiveChannelsWithDetailsAndExclusions(warnSeconds, archiveSeconds, userMap, excludeChannelsList, excludePrefixesList, isDebug)
	if err != nil {
		// Check if this is a rate limit error and provide helpful guidance
		if strings.Contains(err.Error(), "rate_limited") || strings.Contains(err.Error(), "rate limit") {
			fmt.Printf("⚠️  Slack API rate limit exceeded.\n")
			fmt.Printf("   The analysis was stopped to respect API limits.\n")
			fmt.Printf("   Please wait a few minutes before running the command again.\n")
			fmt.Printf("   \n")
			fmt.Printf("   Tip: Consider running with longer time periods (e.g. --warn-days=30) to reduce API calls.\n")
			return nil, nil, fmt.Errorf("rate limited by Slack API")
		}
		return nil, nil, fmt.Errorf("failed to analyze inactive channels: %w", err)
	}
	return toWarn, toArchive, nil
}

// displayChannelDetails shows channel information with last message details.
func displayChannelDetails(channels []slack.Channel, title string) {
	fmt.Printf("%s:\n", title)
	for _, channel := range channels {
		// Calculate days of inactivity
		daysSinceActive := int(time.Since(channel.LastActivity).Hours() / 24)
		daysText := "days"
		if daysSinceActive == 1 {
			daysText = "day"
		}

		fmt.Printf("  #%s (inactive since: %s, %d %s ago, members: %d)\n",
			channel.Name,
			channel.LastActivity.Format("2006-01-02 15:04:05"),
			daysSinceActive,
			daysText,
			channel.MemberCount)

		// Show last message details if available
		if channel.LastMessage != nil {
			messageText := channel.LastMessage.Text
			if len(messageText) > 60 {
				messageText = messageText[:57] + "..."
			}
			messageText = strings.ReplaceAll(messageText, "\n", " ")

			authorName := channel.LastMessage.UserName
			if authorName == "" {
				authorName = channel.LastMessage.User
			}

			botIndicator := ""
			if channel.LastMessage.IsBot {
				botIndicator = " (bot)"
			}

			fmt.Printf("    └─ Last message by: %s%s | \"%s\"\n", authorName, botIndicator, messageText)
		}
	}
	fmt.Println()
}

// processWarnings handles warning channels in both dry-run and real modes.
func processWarnings(client *slack.Client, toWarn []slack.Channel, warnSeconds, archiveSeconds int, isDryRun bool) {
	displayChannelDetails(toWarn, "Channels to warn about inactivity")

	if isDryRun {
		fmt.Printf("--- DRY RUN ---\n")
		fmt.Printf("Would warn %d channels about upcoming archival\n", len(toWarn))
		if len(toWarn) > 0 {
			fmt.Printf("Example warning message for #%s:\n", toWarn[0].Name)
			exampleMessage := client.FormatInactiveChannelWarning(toWarn[0], warnSeconds, archiveSeconds)
			fmt.Printf("%s\n", exampleMessage)
		}
		fmt.Printf("--- END DRY RUN ---\n\n")
	} else {
		fmt.Printf("Sending warnings to %d channels (joining channels as needed)...\n", len(toWarn))
		warningsSent := 0
		for _, channel := range toWarn {
			if err := client.WarnInactiveChannel(channel, warnSeconds, archiveSeconds); err != nil {
				logger.WithFields(logger.LogFields{
					"channel": channel.Name,
					"error":   err.Error(),
				}).Error("Failed to send warning")
				fmt.Printf("  Failed to warn #%s: %s\n", channel.Name, err.Error())
			} else {
				warningsSent++
				logger.WithField("channel", channel.Name).Info("Warning sent successfully")
				fmt.Printf("  ✓ Warned #%s\n", channel.Name)
			}
		}
		fmt.Printf("Warnings sent: %d/%d\n\n", warningsSent, len(toWarn))
	}
}

// processArchival handles archiving channels in both dry-run and real modes.
func processArchival(client *slack.Client, toArchive []slack.Channel, warnSeconds, archiveSeconds int, isDryRun bool) {
	displayChannelDetails(toArchive, "Channels to archive (grace period expired)")

	if isDryRun {
		fmt.Printf("--- DRY RUN ---\n")
		fmt.Printf("Would archive %d channels\n", len(toArchive))
		if len(toArchive) > 0 {
			fmt.Printf("Example archival message for #%s:\n", toArchive[0].Name)
			exampleArchivalMessage := client.FormatChannelArchivalMessage(toArchive[0], warnSeconds, archiveSeconds)
			fmt.Printf("%s\n", exampleArchivalMessage)
		}
		fmt.Printf("--- END DRY RUN ---\n\n")
	} else {
		fmt.Printf("Archiving %d channels...\n", len(toArchive))
		archived := 0
		for _, channel := range toArchive {
			if err := client.ArchiveChannelWithThresholds(channel, warnSeconds, archiveSeconds); err != nil {
				logger.WithFields(logger.LogFields{
					"channel": channel.Name,
					"error":   err.Error(),
				}).Error("Failed to archive channel")
				fmt.Printf("  Failed to archive #%s: %s\n", channel.Name, err.Error())
			} else {
				archived++
				logger.WithField("channel", channel.Name).Info("Channel archived successfully")
				fmt.Printf("  ✓ Archived #%s\n", channel.Name)
			}
		}
		fmt.Printf("Channels archived: %d/%d\n\n", archived, len(toArchive))
	}
}

func runHighlight(cmd *cobra.Command, args []string) error {
	token := viper.GetString("token")
	if token == "" {
		return fmt.Errorf("slack token is required. Set SLACK_TOKEN environment variable or use --token flag")
	}

	// announce-to is mandatory when committing changes
	if announceTo == "" && commit {
		return fmt.Errorf("--announce-to is required when using --commit")
	}

	if count <= 0 {
		return fmt.Errorf("count must be positive, got %d", count)
	}

	client, err := slack.NewClient(token)
	if err != nil {
		return fmt.Errorf("failed to create Slack client: %w", err)
	}

	// Validate that the announce-to channel exists (if specified)
	if announceTo != "" {
		_, err = client.ResolveChannelNameToID(announceTo)
		if err != nil {
			return fmt.Errorf("announce-to channel '%s' not found: %w", announceTo, err)
		}
	}

	return runHighlightWithClient(client, count, announceTo, !commit)
}

func runHighlightWithClient(client *slack.Client, highlightCount int, announceChannel string, isDryRun bool) error {
	randomChannels, err := client.GetRandomChannels(highlightCount)
	if err != nil {
		return fmt.Errorf("failed to get random channels: %w", err)
	}

	if len(randomChannels) == 0 {
		fmt.Printf("No channels found to highlight.\n")
		return nil
	}

	fmt.Printf("Random channels to highlight (%d): ", len(randomChannels))
	channelNames := make([]string, len(randomChannels))
	for i, channel := range randomChannels {
		channelNames[i] = "#" + channel.Name
	}
	fmt.Printf("%s\n\n", strings.Join(channelNames, ", "))

	if announceChannel != "" {
		return handleHighlightAnnouncement(client, randomChannels, announceChannel, isDryRun)
	}

	return handleHighlightDryRunWithoutChannel(client, randomChannels, isDryRun)
}

func handleHighlightAnnouncement(client *slack.Client, channels []slack.Channel, announceChannel string, isDryRun bool) error {
	message := client.FormatChannelHighlightAnnouncement(channels)

	if isDryRun {
		dryRunMessage := client.FormatChannelHighlightAnnouncementDryRun(channels)
		fmt.Printf("--- DRY RUN ---\n")
		fmt.Printf("Would announce to channel: %s\n", announceChannel)
		fmt.Printf("Message content:\n%s\n", dryRunMessage)
		fmt.Printf("--- END DRY RUN ---\n")
		fmt.Printf("\nTo actually post this highlight, add --commit to your command\n")
	} else {
		if err := client.PostMessage(announceChannel, message); err != nil {
			logger.WithFields(logger.LogFields{
				"channel": announceChannel,
				"error":   err.Error(),
			}).Error("Failed to post highlight")
			return fmt.Errorf("failed to post highlight to %s: %w", announceChannel, err)
		}
		fmt.Printf("Channel highlight posted to %s\n", announceChannel)
	}
	return nil
}

func handleHighlightDryRunWithoutChannel(client *slack.Client, channels []slack.Channel, isDryRun bool) error {
	if isDryRun {
		message := client.FormatChannelHighlightAnnouncementDryRun(channels)
		fmt.Printf("--- DRY RUN ---\n")
		fmt.Printf("Channel highlight message dry run (use --announce-to to specify target):\n%s\n", message)
		fmt.Printf("--- END DRY RUN ---\n")
		fmt.Printf("\nTo actually post highlights, add --commit to your command\n")
	}
	return nil
}
