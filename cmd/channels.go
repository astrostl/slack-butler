package cmd

import (
	"fmt"
	"slack-buddy/pkg/slack"
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

	newChannels, err := client.GetNewChannels(cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to get new channels: %v", err)
	}

	if len(newChannels) == 0 {
		fmt.Printf("No new channels found since %s\n", cutoffTime.Format("2006-01-02 15:04:05"))
		return nil
	}

	fmt.Printf("Found %d new channel(s) since %s:\n", len(newChannels), cutoffTime.Format("2006-01-02 15:04:05"))
	for _, channel := range newChannels {
		fmt.Printf("  #%s (created: %s)\n", channel.Name, channel.Created.Format("2006-01-02 15:04:05"))
	}

	if announceTo != "" {
		message := client.FormatNewChannelAnnouncement(newChannels, cutoffTime)
		if err := client.PostMessage(announceTo, message); err != nil {
			return fmt.Errorf("failed to post announcement to %s: %v", announceTo, err)
		}
		fmt.Printf("Announcement posted to %s\n", announceTo)
	}

	return nil
}