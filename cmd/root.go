package cmd

import (
	"fmt"
	"os"

	"github.com/astrostl/slack-butler/pkg/logger"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version information passed from main.
var (
	version   string
	buildTime string
	gitCommit string
)

// Constants for version info.
const unknownValue = "unknown"

var rootCmd = &cobra.Command{
	Use:   "slack-butler",
	Short: "A CLI tool to help manage Slack workspaces",
	Long:  `Slack Butler is a CLI tool that helps make Slack workspaces more useful and tidy.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle --version flag when no subcommand is specified
		versionFlag, err := cmd.Flags().GetBool("version")
		if err == nil && versionFlag {
			fmt.Printf("slack-butler version %s\n", version)
			if buildTime != unknownValue {
				fmt.Printf("Built: %s\n", buildTime)
			}
			if gitCommit != unknownValue {
				fmt.Printf("Commit: %s\n", gitCommit)
			}
			return
		}
		// Show help if no version flag and no subcommand
		if err := cmd.Help(); err != nil {
			// Ignore help display errors
			return
		}
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute(ver, build, commit string) {
	// Store version information for use in commands
	version = ver
	buildTime = build
	gitCommit = commit

	if err := rootCmd.Execute(); err != nil {
		// Cobra already displays the error, no need to log it again
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().String("token", "", "Slack bot token (can also be set via SLACK_TOKEN env var)")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging")
	rootCmd.Flags().BoolP("version", "v", false, "Print version information")

	// Bind flags to viper
	if err := viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token")); err != nil {
		logger.WithField("error", err.Error()).Fatal("Failed to bind token flag")
	}
	if err := viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug")); err != nil {
		logger.WithField("error", err.Error()).Fatal("Failed to bind debug flag")
	}

}

func initConfig() {
	viper.SetEnvPrefix("SLACK")
	viper.AutomaticEnv()

	// Explicitly bind environment variables
	if err := viper.BindEnv("token", "SLACK_TOKEN"); err != nil {
		// BindEnv rarely fails, but handle for completeness
		return
	}
	if err := viper.BindEnv("debug", "SLACK_DEBUG"); err != nil {
		// BindEnv rarely fails, but handle for completeness
		return
	}

	// Set log level based on debug flag
	if viper.GetBool("debug") {
		logger.Log.SetLevel(logrus.DebugLevel)
	} else {
		logger.Log.SetLevel(logrus.InfoLevel)
	}
}
