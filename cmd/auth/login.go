// Package auth provides authentication commands.
package auth

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
	// LoginAppID is the Telegram API App ID.
	LoginAppID   string
	// LoginAppHash is the Telegram API App Hash.
	LoginAppHash string
	// LoginPhone is the phone number for login.
	LoginPhone string
	// LoginMock enables mock mode for UI testing.
	LoginMock    bool
)

// LoginCmd represents the login command.
var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Interactive login with Telegram authentication",
	Long: `Interactively login to Telegram using authentication code.

This will guide you through the login process:
  1. Enter your phone number
  2. Enter the verification code sent to Telegram
  3. Enter 2FA password if enabled`,
	Run: runLogin,
}

// AddLoginCommand adds the login command to the root command.
func AddLoginCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(LoginCmd)

	LoginCmd.Flags().StringVar(&LoginAppID, "app-id", os.Getenv("AGENT_TELEGRAM_APP_ID"), "Telegram API App ID")
	LoginCmd.Flags().StringVar(&LoginAppHash, "app-hash", os.Getenv("AGENT_TELEGRAM_APP_HASH"), "Telegram API App Hash")
	LoginCmd.Flags().StringVar(&LoginPhone, "phone", os.Getenv("AGENT_TELEGRAM_PHONE"), "Phone number")
	LoginCmd.Flags().BoolVar(&LoginMock, "mock", false, "Mock mode for UI testing (no real API calls)")
}

func runLogin(_ *cobra.Command, _ []string) {
	// Parse app ID (optional, uses hardcoded default if not provided)
	var appID int
	var err error
	if LoginAppID != "" {
		appID, err = strconv.Atoi(LoginAppID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid app-id: %v\n", err)
			os.Exit(1)
		}
	}

	// Create context
	ctx := context.Background()

	// Create auth service (skip in mock mode)
	var authService *auth.Service
	if !LoginMock {
		authService, err = createAuthService(ctx, appID, LoginAppHash, LoginPhone)
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
