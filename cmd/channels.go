package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astrostl/slack-buddy-ai/pkg/logger"
	"github.com/astrostl/slack-buddy-ai/pkg/slack"

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
	Long: `Detect new channels created during a specified time period and announce them to another channel.

The --announce-to flag is required to specify the target channel.
Use --commit to actually post messages (default is dry run mode).`,
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

NOTE: Currently using SECONDS for testing (will be changed back to days later).`,
	SilenceUsage: true, // Don't show usage on errors
	RunE:         runArchive,
}

var (
	since           string
	announceTo      string
	commit          bool
	warnSeconds     int
	archiveSeconds  int
	excludeChannels string
	excludePrefixes string
)

func init() {
	rootCmd.AddCommand(channelsCmd)
	channelsCmd.AddCommand(detectCmd)
	channelsCmd.AddCommand(archiveCmd)

	detectCmd.Flags().StringVar(&since, "since", "8", "Number of days to look back (e.g., 1, 7, 30)")
	detectCmd.Flags().StringVar(&announceTo, "announce-to", "", "Channel to announce new channels to (e.g., #general) [REQUIRED]")
	detectCmd.Flags().BoolVar(&commit, "commit", false, "Actually post messages (default is dry run mode)")

	archiveCmd.Flags().IntVar(&warnSeconds, "warn-seconds", 300, "Number of seconds of inactivity before warning (default: 300 = 5 minutes)")
	archiveCmd.Flags().IntVar(&archiveSeconds, "archive-seconds", 60, "Number of seconds after warning (with no new activity) before archiving (default: 60 = 1 minute)")
	archiveCmd.Flags().BoolVar(&commit, "commit", false, "Actually warn and archive channels (default is dry run mode)")
	archiveCmd.Flags().StringVar(&excludeChannels, "exclude-channels", "", "Comma-separated list of channel names to exclude (with or without # prefix, e.g., 'general,random,#important')")
	archiveCmd.Flags().StringVar(&excludePrefixes, "exclude-prefixes", "", "Comma-separated list of channel prefixes to exclude (with or without # prefix, e.g., 'prod-,#temp-,admin')")
}

func runDetect(cmd *cobra.Command, args []string) error {
	token := viper.GetString("token")
	if token == "" {
		return fmt.Errorf("slack token is required. Set SLACK_TOKEN environment variable or use --token flag")
	}

	// announce-to is mandatory
	if announceTo == "" {
		return fmt.Errorf("--announce-to is required")
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

	// Validate that the announce-to channel exists
	_, err = client.ResolveChannelNameToID(announceTo)
	if err != nil {
		return fmt.Errorf("announce-to channel '%s' not found: %w", announceTo, err)
	}

	return runDetectWithClient(client, cutoffTime, announceTo, !commit)
}

func runDetectWithClient(client *slack.Client, cutoffTime time.Time, announceChannel string, isDryRun bool) error {
	newChannels, allChannels, err := client.GetNewChannelsWithAllChannels(cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to get new channels: %w", err)
	}

	if len(newChannels) == 0 {
		return nil
	}

	// Show simple summary list on one line
	channelList := make([]string, len(newChannels))
	for i, channel := range newChannels {
		channelList[i] = "#" + channel.Name
	}
	fmt.Printf("New channels found (%d): %s\n\n", len(newChannels), strings.Join(channelList, ", "))

	if announceChannel != "" {
		message := client.FormatNewChannelAnnouncement(newChannels, cutoffTime)

		// Extract channel names for duplicate checking
		channelNames := make([]string, len(newChannels))
		for i, channel := range newChannels {
			channelNames[i] = channel.Name
		}

		// Check for duplicate announcements and track skipped channels (regardless of dry run mode)
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
			// Separate channels into new and skipped
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

			if len(skippedChannels) > 0 {
				fmt.Printf("Channels already announced (skipped): %s\n", strings.Join(skippedChannels, ", "))
			}

			// If all channels were already announced, skip entirely
			if len(newChannelsToAnnounce) == 0 {
				fmt.Printf("All channels already announced, skipping announcement to %s\n", announceChannel)
				return nil
			}

			// Update the message to only include new channels
			var newChannelsToFormat []slack.Channel
			for _, channel := range newChannels {
				for _, newChannelName := range newChannelsToAnnounce {
					if channel.Name == newChannelName {
						newChannelsToFormat = append(newChannelsToFormat, channel)
						break
					}
				}
			}
			channelsToAnnounce = newChannelsToFormat
			finalMessage = client.FormatNewChannelAnnouncement(newChannelsToFormat, cutoffTime)

			// Show what's being announced
			if len(newChannelsToAnnounce) > 0 {
				var announcingList []string
				for _, channelName := range newChannelsToAnnounce {
					announcingList = append(announcingList, "#"+channelName)
				}
				fmt.Printf("Announcing channels: %s (skipped %d already announced)\n", strings.Join(announcingList, ", "), len(skippedChannels))
			}
		} else {
			channelsToAnnounce = newChannels
			finalMessage = message
		}

		if isDryRun {
			// For dry run, create a pretty version with readable names
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
	} else if isDryRun {
		// Show what announcement message would look like even without a target channel
		message := client.FormatNewChannelAnnouncement(newChannels, cutoffTime)
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

	if warnSeconds <= 0 {
		return fmt.Errorf("warn-seconds must be positive, got %d", warnSeconds)
	}

	if archiveSeconds <= 0 {
		return fmt.Errorf("archive-seconds must be positive, got %d", archiveSeconds)
	}

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
	if isDebug {
		fmt.Printf("📞 API Call 1: Getting user list for name resolution...\n")
	}
	userMap, err := client.GetUserMap()
	if err != nil {
		if strings.Contains(err.Error(), "rate_limited") || strings.Contains(err.Error(), "rate limit") {
			fmt.Printf("⚠️  Slack API rate limit exceeded on user list.\n")
			fmt.Printf("   The system should have done backoff.\n")
			return fmt.Errorf("rate limited by Slack API")
		}
		if strings.Contains(err.Error(), "missing_scope") || strings.Contains(err.Error(), "users:read") {
			fmt.Printf("❌ Missing required OAuth scope 'users:read'\n")
			fmt.Printf("   This scope is needed to resolve user names for message authors.\n")
			fmt.Printf("   Add 'users:read' scope in your Slack app settings at https://api.slack.com/apps\n")
			return fmt.Errorf("missing required OAuth scope 'users:read'")
		}
		return fmt.Errorf("failed to get users: %w", err)
	}
	if isDebug {
		fmt.Printf("✅ Got %d users from API\n\n", len(userMap))
	}

	// Parse exclusion lists
	var excludeChannelsList []string
	var excludePrefixesList []string

	if excludeChannels != "" {
		for _, channel := range strings.Split(excludeChannels, ",") {
			channel = strings.TrimSpace(channel)
			// Remove # prefix if present
			if strings.HasPrefix(channel, "#") {
				channel = channel[1:]
			}
			if channel != "" {
				excludeChannelsList = append(excludeChannelsList, channel)
			}
		}
	}

	if excludePrefixes != "" {
		for _, prefix := range strings.Split(excludePrefixes, ",") {
			prefix = strings.TrimSpace(prefix)
			// Remove # prefix if present
			if strings.HasPrefix(prefix, "#") {
				prefix = prefix[1:]
			}
			if prefix != "" {
				excludePrefixesList = append(excludePrefixesList, prefix)
			}
		}
	}

	// Show exclusion info if any are specified
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

	toWarn, toArchive, err := client.GetInactiveChannelsWithDetailsAndExclusions(warnSeconds, archiveSeconds, userMap, excludeChannelsList, excludePrefixesList, isDebug)
	if err != nil {
		// Check if this is a rate limit error and provide helpful guidance
		if strings.Contains(err.Error(), "rate_limited") || strings.Contains(err.Error(), "rate limit") {
			fmt.Printf("⚠️  Slack API rate limit exceeded.\n")
			fmt.Printf("   The analysis was stopped to respect API limits.\n")
			fmt.Printf("   Please wait a few minutes before running the command again.\n")
			fmt.Printf("   \n")
			fmt.Printf("   Tip: Consider running with longer time periods (e.g. --warn-seconds=3600) to reduce API calls.\n")
			return fmt.Errorf("rate limited by Slack API")
		}
		return fmt.Errorf("failed to analyze inactive channels: %w", err)
	}

	// Report findings
	fmt.Printf("Inactive Channel Analysis Results:\n")
	fmt.Printf("  Channels to warn: %d\n", len(toWarn))
	fmt.Printf("  Channels to archive: %d\n", len(toArchive))
	fmt.Println()

	// Process warnings
	if len(toWarn) > 0 {
		fmt.Printf("Channels to warn about inactivity:\n")
		for _, channel := range toWarn {
			fmt.Printf("  #%s (inactive since: %s, members: %d)\n",
				channel.Name,
				channel.LastActivity.Format("2006-01-02 15:04:05"),
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

	// Process archival
	if len(toArchive) > 0 {
		fmt.Printf("Channels to archive (grace period expired):\n")
		for _, channel := range toArchive {
			fmt.Printf("  #%s (inactive since: %s, members: %d)\n",
				channel.Name,
				channel.LastActivity.Format("2006-01-02 15:04:05"),
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

	if len(toWarn) == 0 && len(toArchive) == 0 {
		fmt.Printf("No inactive channels found. All channels are active or already processed.\n")
	}

	return nil
}
