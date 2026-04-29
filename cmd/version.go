package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Display version information including build time and git commit.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("slack-butler version %s\n", version)
		if buildTime != unknownValue {
			fmt.Printf("Built: %s\n", buildTime)
		}
		if gitCommit != unknownValue {
			fmt.Printf("Commit: %s\n", gitCommit)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
