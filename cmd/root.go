package cmd

import (
	"os"
	"slack-buddy-ai/pkg/logger"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "slack-buddy",
	Short: "A CLI tool to help manage Slack workspaces",
	Long:  `Slack Buddy is a CLI tool that helps make Slack workspaces more useful and tidy.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.WithField("error", err.Error()).Error("Command execution failed")
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().String("token", "", "Slack bot token (can also be set via SLACK_TOKEN env var)")
	if err := viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token")); err != nil {
		logger.WithField("error", err.Error()).Fatal("Failed to bind token flag")
	}
}

func initConfig() {
	viper.SetEnvPrefix("SLACK")
	viper.AutomaticEnv()
}