// Package cmd provides CLI commands.
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"agent-telegram/cli/ui"
	"agent-telegram/internal/auth"
)

var (
	loginAppID   string
	loginAppHash string
	loginPhone   string
	loginMock    bool
)

// loginCmd represents the login command.
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Interactive login with Telegram authentication",
	Long: `Interactively login to Telegram using authentication code.

This will guide you through the login process:
  1. Enter your phone number
  2. Enter the verification code sent to Telegram
  3. Enter 2FA password if enabled`,
	Run: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVar(&loginAppID, "app-id", os.Getenv("AGENT_TELEGRAM_APP_ID"), "Telegram API App ID")
	loginCmd.Flags().StringVar(&loginAppHash, "app-hash", os.Getenv("AGENT_TELEGRAM_APP_HASH"), "Telegram API App Hash")
	loginCmd.Flags().StringVar(&loginPhone, "phone", os.Getenv("AGENT_TELEGRAM_PHONE"), "Phone number")
	loginCmd.Flags().BoolVar(&loginMock, "mock", false, "Mock mode for UI testing (no real API calls)")
}

func runLogin(_ *cobra.Command, _ []string) {
	// Parse app ID (optional, uses hardcoded default if not provided)
	var appID int
	var err error
	if loginAppID != "" {
		appID, err = strconv.Atoi(loginAppID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid app-id: %v\n", err)
			os.Exit(1)
		}
	}

	// Create context
	ctx := context.Background()

	// Create auth service (skip in mock mode)
	var authService *auth.Service
	if !loginMock {
		authService, err = createAuthService(ctx, appID, loginAppHash, loginPhone)
		if err != nil {
			log.Fatalf("Failed to create auth service: %v", err)
		}
	}

	// Run interactive login UI with Telegram authentication
	sessionPath, err := ui.RunLoginUIWithAuth(ctx, authService)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	fmt.Printf("\nSession saved to: %s\n", sessionPath)
	fmt.Println("You can now use the serve command to start the server")
}

// createAuthService creates an auth service based on the provided parameters.
// If phone is provided with appID and appHash, it uses NewServiceWithConfig.
// Otherwise, it uses NewService.
func createAuthService(ctx context.Context, appID int, appHash, phone string) (*auth.Service, error) {
	// Use configured values if phone is provided with valid app credentials
	if phone != "" && appID != 0 && appHash != "" {
		return auth.NewServiceWithConfig(ctx, appID, appHash, phone)
	}
	// Use default service (will prompt for credentials)
	return auth.NewService(ctx)
}
