package cmd

import (
	"fmt"
	"strconv"
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
	Long: `Detect new channels created during a specified time period and optionally announce them to another channel.

Use --dry-run to preview what would be announced without actually posting messages.`,
	SilenceUsage: true, // Don't show usage on errors
	RunE:         runDetect,
}

var (
	since      string
	announceTo string
	dryRun     bool
)

func init() {
	rootCmd.AddCommand(channelsCmd)
	channelsCmd.AddCommand(detectCmd)

	detectCmd.Flags().StringVar(&since, "since", "1", "Number of days to look back (e.g., 1, 7, 30)")
	detectCmd.Flags().StringVar(&announceTo, "announce-to", "", "Channel to announce new channels to (e.g., #general)")
	detectCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be announced without actually posting messages")
}

func runDetect(cmd *cobra.Command, args []string) error {
	token := viper.GetString("token")
	if token == "" {
		return fmt.Errorf("slack token is required. Set SLACK_TOKEN environment variable or use --token flag")
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

	return runDetectWithClient(client, cutoffTime, announceTo, dryRun)
}

func runDetectWithClient(client *slack.Client, cutoffTime time.Time, announceChannel string, isDryRun bool) error {
	newChannels, err := client.GetNewChannels(cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to get new channels: %w", err)
	}

	if len(newChannels) == 0 {
		logger.WithFields(logger.LogFields{
			"since": cutoffTime.Format("2006-01-02 15:04:05"),
		}).Info("No new channels found")
		return nil
	}

	logger.WithFields(logger.LogFields{
		"count": len(newChannels),
		"since": cutoffTime.Format("2006-01-02 15:04:05"),
	}).Info("Found new channels")

	for _, channel := range newChannels {
		logger.WithFields(logger.LogFields{
			"channel": channel.Name,
			"created": channel.Created.Format("2006-01-02 15:04:05"),
		}).Info("New channel detected")
		fmt.Printf("  #%s (created: %s)\n", channel.Name, channel.Created.Format("2006-01-02 15:04:05"))
	}

	// Add summary list at the end for easy copying
	if len(newChannels) > 0 {
		fmt.Println()
		fmt.Printf("New channels found (%d):\n", len(newChannels))
		for _, channel := range newChannels {
			fmt.Printf("  #%s\n", channel.Name)
		}
	}

	if announceChannel != "" {
		message := client.FormatNewChannelAnnouncement(newChannels, cutoffTime)

		if isDryRun {
			fmt.Printf("\n--- DRY RUN ---\n")
			fmt.Printf("Would announce to channel: %s\n", announceChannel)
			fmt.Printf("Message content:\n%s\n", message)
			fmt.Printf("--- END DRY RUN ---\n")
			logger.WithField("channel", announceChannel).Info("Dry run: announcement preview shown")
		} else {
			if err := client.PostMessage(announceChannel, message); err != nil {
				logger.WithFields(logger.LogFields{
					"channel": announceChannel,
					"error":   err.Error(),
				}).Error("Failed to post announcement")
				return fmt.Errorf("failed to post announcement to %s: %w", announceChannel, err)
			}
			logger.WithField("channel", announceChannel).Info("Announcement posted successfully")
			fmt.Printf("Announcement posted to %s\n", announceChannel)
		}
	} else if isDryRun {
		// Show what announcement message would look like even without a target channel
		message := client.FormatNewChannelAnnouncement(newChannels, cutoffTime)
		fmt.Printf("\n--- DRY RUN ---\n")
		fmt.Printf("Announcement message preview (use --announce-to to specify target):\n%s\n", message)
		fmt.Printf("--- END DRY RUN ---\n")
		logger.Info("Dry run: message preview shown without target channel")
	}

	return nil
}
