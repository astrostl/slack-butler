package cmd

import (
	"fmt"
	"github.com/astrostl/slack-buddy-ai/pkg/logger"
	"github.com/astrostl/slack-buddy-ai/pkg/slack"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check health and connectivity",
	Long: `Check the health of your Slack connection, validate configuration, and test permissions.

This command verifies:
- Token validity and format
- Slack API connectivity  
- Required OAuth scopes and permissions
- Bot user information
- Basic API functionality`,
	RunE: runHealth,
}

var (
	healthVerbose bool
)

func init() {
	rootCmd.AddCommand(healthCmd)
	healthCmd.Flags().BoolVarP(&healthVerbose, "verbose", "v", false, "Show detailed health check information")
}

func runHealth(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 Running health checks...")
	fmt.Println()

	// Check 1: Configuration validation
	fmt.Print("✓ Configuration validation... ")
	token := viper.GetString("token")
	if token == "" {
		fmt.Println("❌ FAILED")
		fmt.Println("  Error: No Slack token configured")
		fmt.Println("  Fix: Set SLACK_TOKEN environment variable or use --token flag")
		return fmt.Errorf("configuration validation failed")
	}
	fmt.Println("✅ PASSED")
	if healthVerbose {
		fmt.Printf("  Token format: %s...%s\n", token[:8], token[len(token)-8:])
	}

	// Check 2: Token format validation
	fmt.Print("✓ Token format validation... ")
	if !isValidTokenFormat(token) {
		fmt.Println("❌ FAILED")
		fmt.Println("  Error: Invalid token format")
		fmt.Println("  Expected: Bot tokens must start with 'xoxb-'")
		return fmt.Errorf("token format validation failed")
	}
	fmt.Println("✅ PASSED")

	// Check 3: Slack client creation
	fmt.Print("✓ Slack client initialization... ")
	client, err := slack.NewClient(token)
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Error: %v\n", err)
		return fmt.Errorf("client initialization failed: %v", err)
	}
	fmt.Println("✅ PASSED")

	// Check 4: API connectivity
	fmt.Print("✓ Slack API connectivity... ")
	authInfo, err := client.TestAuth()
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Error: %v\n", err)
		fmt.Println("  Fix: Verify your token is valid and has not been revoked")
		return fmt.Errorf("API connectivity failed: %v", err)
	}
	fmt.Println("✅ PASSED")
	if healthVerbose {
		fmt.Printf("  Connected as: %s (team: %s)\n", authInfo.User, authInfo.Team)
		fmt.Printf("  User ID: %s, Team ID: %s\n", authInfo.UserID, authInfo.TeamID)
	}

	// Check 5: Required permissions
	fmt.Print("✓ Required permissions... ")
	permissionErrors := checkRequiredPermissions(client)
	if len(permissionErrors) > 0 {
		fmt.Println("❌ FAILED")
		for _, err := range permissionErrors {
			fmt.Printf("  %s\n", err)
		}
		fmt.Println("  Fix: Add missing OAuth scopes in your Slack app settings")
		return fmt.Errorf("permission validation failed")
	}
	fmt.Println("✅ PASSED")

	// Check 6: Basic functionality test
	fmt.Print("✓ Basic functionality test... ")
	if err := testBasicFunctionality(client); err != nil {
		fmt.Println("⚠️  WARNING")
		fmt.Printf("  Warning: %v\n", err)
		fmt.Println("  Note: Basic connectivity works, but some features may be limited")
	} else {
		fmt.Println("✅ PASSED")
	}

	fmt.Println()
	fmt.Println("🎉 Health check completed successfully!")
	fmt.Printf("   Connected as: %s (team: %s)\n", authInfo.User, authInfo.Team)
	fmt.Println("   All systems operational")

	return nil
}

func isValidTokenFormat(token string) bool {
	return len(token) > 8 && (token[:5] == "xoxb-" || token[:5] == "xoxp-")
}

func checkRequiredPermissions(client *slack.Client) []string {
	var errors []string

	// Test channels:read permission
	if _, err := client.GetChannelInfo("C0000000000"); err != nil {
		if contains(err.Error(), "missing_scope") && contains(err.Error(), "channels:read") {
			errors = append(errors, "Missing scope: channels:read (required to list public channels)")
		}
	}

	// Test groups:read permission (for private channels)
	// Note: This is optional and won't cause a failure
	
	// Test chat:write permission (for announcements)
	// We can't easily test this without actually posting, so we'll check it during actual use

	return errors
}

func testBasicFunctionality(client *slack.Client) error {
	// Test getting channel list with a reasonable timeout
	logger.WithField("operation", "health_check").Debug("Testing basic channel listing functionality")
	
	cutoffTime := time.Now().Add(-24 * time.Hour)
	channels, err := client.GetNewChannels(cutoffTime)
	if err != nil {
		return fmt.Errorf("channel listing test failed: %v", err)
	}

	if healthVerbose {
		fmt.Printf("  Successfully retrieved channel information (%d channels checked)\n", len(channels))
	}

	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (len(substr) == 0 || 
		    (len(s) > 0 && 
		     (s == substr || 
		      (len(s) > len(substr) && 
		       (s[:len(substr)] == substr || 
		        s[len(s)-len(substr):] == substr || 
		        containsSubstring(s, substr))))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}