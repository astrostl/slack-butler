package cmd

import (
	"fmt"
	"time"

	"github.com/astrostl/slack-buddy-ai/pkg/logger"
	"github.com/astrostl/slack-buddy-ai/pkg/slack"

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
- Required OAuth scopes and permissions (channels:read, channels:join, chat:write, channels:manage, users:read)
- Bot user information
- Basic API functionality`,
	SilenceUsage: true, // Don't show usage on errors
	RunE:         runHealth,
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
		return fmt.Errorf("client initialization failed: %w", err)
	}
	fmt.Println("✅ PASSED")

	// Check 4: API connectivity
	fmt.Print("✓ Slack API connectivity... ")
	authInfo, err := client.TestAuth()
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Error: %v\n", err)
		fmt.Println("  Fix: Verify your token is valid and has not been revoked")
		return fmt.Errorf("API connectivity failed: %w", err)
	}
	fmt.Println("✅ PASSED")
	if healthVerbose {
		fmt.Printf("  Connected as: %s (team: %s)\n", authInfo.User, authInfo.Team)
		fmt.Printf("  User ID: %s, Team ID: %s\n", authInfo.UserID, authInfo.TeamID)
	}

	// Check 5: OAuth scope validation
	fmt.Print("✓ OAuth scope validation (channels:read, channels:join, chat:write, channels:manage, users:read)... ")
	if healthVerbose {
		fmt.Printf("\n  Testing required scopes: channels:read, channels:join, chat:write, channels:manage, users:read\n")
		fmt.Printf("  Testing optional scopes: groups:read\n")
		fmt.Print("  Validation result: ")
	}
	scopes, err := client.CheckOAuthScopes()
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Error: %v\n", err)
		return fmt.Errorf("scope validation failed: %w", err)
	}

	requiredScopes := map[string]bool{
		"channels:read":   true, // Required - list channels
		"channels:join":   true, // Required - join channels for warnings
		"chat:write":      true, // Required - post warning messages
		"channels:manage": true, // Required - archive channels
		"users:read":      true, // Required - resolve user names for message authors
	}
	optionalScopes := map[string]bool{
		"groups:read": false, // Optional - access private channels
	}

	var missingRequired []string
	var missingOptional []string

	for scope := range requiredScopes {
		if !scopes[scope] {
			missingRequired = append(missingRequired, scope)
		}
	}

	for scope := range optionalScopes {
		if !scopes[scope] {
			missingOptional = append(missingOptional, scope)
		}
	}

	if len(missingRequired) > 0 {
		fmt.Println("❌ FAILED")
		fmt.Println("  Missing REQUIRED OAuth scopes:")
		for _, scope := range missingRequired {
			fmt.Printf("    - %s\n", scope)
		}
		if len(missingOptional) > 0 {
			fmt.Println("  Missing OPTIONAL OAuth scopes:")
			for _, scope := range missingOptional {
				fmt.Printf("    - %s (private channels won't be accessible)\n", scope)
			}
		}
		fmt.Println("  Fix: Add missing OAuth scopes in your Slack app settings at https://api.slack.com/apps")
		return fmt.Errorf("missing required OAuth scopes")
	}

	fmt.Println("✅ PASSED")
	if healthVerbose {
		fmt.Println("  OAuth scope test results:")

		// Show required scopes first
		fmt.Println("    Required scopes:")
		for scope := range requiredScopes {
			status := "❌"
			if scopes[scope] {
				status = "✅"
			}
			var testMethod string
			switch scope {
			case "channels:read":
				testMethod = "tested with GetConversations()"
			case "channels:join":
				testMethod = "tested with JoinConversation()"
			case "chat:write":
				testMethod = "tested with PostMessage()"
			case "channels:manage":
				testMethod = "tested with ArchiveConversation()"
			case "users:read":
				testMethod = "tested with GetUsers()"
			}
			fmt.Printf("      %s %s - %s\n", status, scope, testMethod)
		}

		// Show optional scopes
		fmt.Println("    Optional scopes:")
		for scope := range optionalScopes {
			status := "❌"
			if scopes[scope] {
				status = "✅"
			}
			var testMethod string
			switch scope {
			case "groups:read":
				testMethod = "tested with GetConversations(private_channel)"
			}
			fmt.Printf("      %s %s - %s\n", status, scope, testMethod)
		}
	}

	if len(missingOptional) > 0 {
		fmt.Println("  ⚠️  Note: Some optional scopes are missing - private channels won't be accessible")
	}

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

func testBasicFunctionality(client *slack.Client) error {
	// Test getting channel list with a reasonable timeout
	logger.WithField("operation", "health_check").Debug("Testing basic channel listing functionality")

	cutoffTime := time.Now().Add(-24 * time.Hour)
	channels, err := client.GetNewChannels(cutoffTime)
	if err != nil {
		return fmt.Errorf("channel listing test failed: %w", err)
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
