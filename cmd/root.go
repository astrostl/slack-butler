package cmd

import (
	"fmt"
	"os"
	"slack-buddy-ai/pkg/logger"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version information passed from main
var (
	version   string
	buildTime string
	gitCommit string
)

var rootCmd = &cobra.Command{
	Use:   "slack-buddy",
	Short: "A CLI tool to help manage Slack workspaces",
	Long:  `Slack Buddy is a CLI tool that helps make Slack workspaces more useful and tidy.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle --version flag when no subcommand is specified
		if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
			fmt.Printf("slack-buddy version %s\n", version)
			if buildTime != "unknown" {
				fmt.Printf("Built: %s\n", buildTime)
			}
			if gitCommit != "unknown" {
				fmt.Printf("Commit: %s\n", gitCommit)
			}
			return
		}
		// Show help if no version flag and no subcommand
		cmd.Help()
	},
}

func Execute(ver, build, commit string) {
	// Store version information for use in commands
	version = ver
	buildTime = build
	gitCommit = commit

	if err := rootCmd.Execute(); err != nil {
		logger.WithField("error", err.Error()).Error("Command execution failed")
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	
	// Global flags
	rootCmd.PersistentFlags().String("token", "", "Slack bot token (can also be set via SLACK_TOKEN env var)")
	rootCmd.Flags().BoolP("version", "v", false, "Print version information")
	
	// Bind flags to viper
	if err := viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token")); err != nil {
		logger.WithField("error", err.Error()).Fatal("Failed to bind token flag")
	}
	
}

func initConfig() {
	viper.SetEnvPrefix("SLACK")
	viper.AutomaticEnv()
}