package cmd

import (
	"fmt"
	"slack-buddy-ai/pkg/logger"
	"slack-buddy-ai/pkg/slack"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var channelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "Manage channels in your Slack workspace",
	Long:  `Commands for managing and monitoring channels in your Slack workspace.`,
}

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect new channels created in a time period",
	Long:  `Detect new channels created during a specified time period and optionally announce them to another channel.`,
	RunE:  runDetect,
}

var (
	since      string
	announceTo string
)

func init() {
	rootCmd.AddCommand(channelsCmd)
	channelsCmd.AddCommand(detectCmd)

	detectCmd.Flags().StringVar(&since, "since", "24h", "Time period to look back (e.g., 24h, 7d, 1w)")
	detectCmd.Flags().StringVar(&announceTo, "announce-to", "", "Channel to announce new channels to (e.g., #general)")
}

func runDetect(cmd *cobra.Command, args []string) error {
	token := viper.GetString("token")
	if token == "" {
		return fmt.Errorf("slack token is required. Set SLACK_TOKEN environment variable or use --token flag")
	}

	duration, err := time.ParseDuration(since)
	if err != nil {
		return fmt.Errorf("invalid time format '%s': %v", since, err)
	}

	cutoffTime := time.Now().Add(-duration)

	client, err := slack.NewClient(token)
	if err != nil {
		return fmt.Errorf("failed to create Slack client: %v", err)
	}

	return runDetectWithClient(client, cutoffTime, announceTo)
}

func runDetectWithClient(client *slack.Client, cutoffTime time.Time, announceChannel string) error {
	newChannels, err := client.GetNewChannels(cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to get new channels: %v", err)
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

	if announceChannel != "" {
		message := client.FormatNewChannelAnnouncement(newChannels, cutoffTime)
		if err := client.PostMessage(announceChannel, message); err != nil {
			logger.WithFields(logger.LogFields{
				"channel": announceChannel,
				"error": err.Error(),
			}).Error("Failed to post announcement")
			return fmt.Errorf("failed to post announcement to %s: %v", announceChannel, err)
		}
		logger.WithField("channel", announceChannel).Info("Announcement posted successfully")
		fmt.Printf("Announcement posted to %s\n", announceChannel)
	}

	return nil
}