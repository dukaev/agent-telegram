// Package auth provides authentication commands.
package auth

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"agent-telegram/cli/ui"
	"agent-telegram/internal/auth"
	"agent-telegram/internal/config"
	"agent-telegram/internal/ipc"
)

// Default Telegram API credentials.
const (
	defaultAppID   = "35699202"
	defaultAppHash = "7e97f16795114cf3046d1aebf9de886d"
)

var (
	// LoginAppID is the Telegram API App ID.
	LoginAppID string
	// LoginAppHash is the Telegram API App Hash.
	LoginAppHash string
	// LoginPhone is the phone number for login.
	LoginPhone string
	// LoginMock enables mock mode for UI testing.
	LoginMock bool
)

// LoginCmd represents the login command.
var LoginCmd = &cobra.Command{
	GroupID: "auth",
	Use:     "login",
	Short:   "Interactive login with Telegram authentication",
	Long: `Interactively login to Telegram using authentication code.

This will guide you through the login process:
  1. Enter your phone number
  2. Enter the verification code sent to Telegram
  3. Enter 2FA password if enabled

Telegram API credentials (optional):
  Default credentials are used if not provided.

  To use your own credentials (https://my.telegram.org/apps):
    - Set TELEGRAM_APP_ID and TELEGRAM_APP_HASH environment variables
    - Or create a .env file in the current directory
    - Or pass --app-id and --app-hash flags`,
	Run: runLogin,
}

// AddLoginCommand adds the login command to the root command.
func AddLoginCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(LoginCmd)

	LoginCmd.Flags().StringVar(&LoginAppID, "app-id", os.Getenv("TELEGRAM_APP_ID"), "Telegram API App ID")
	LoginCmd.Flags().StringVar(&LoginAppHash, "app-hash", os.Getenv("TELEGRAM_APP_HASH"), "Telegram API App Hash")
	LoginCmd.Flags().StringVar(&LoginPhone, "phone", os.Getenv("AGENT_TELEGRAM_PHONE"), "Phone number")
	LoginCmd.Flags().BoolVar(&LoginMock, "mock", false, "Mock mode for UI testing (no real API calls)")
}

func runLogin(_ *cobra.Command, _ []string) {
	// Load .env file
	_ = godotenv.Load()

	// Re-read env vars after loading .env (flags are evaluated before .env is loaded)
	if LoginAppID == "" {
		LoginAppID = os.Getenv("TELEGRAM_APP_ID")
	}
	if LoginAppHash == "" {
		LoginAppHash = os.Getenv("TELEGRAM_APP_HASH")
	}

	// Use default credentials if not provided
	if LoginAppID == "" {
		LoginAppID = defaultAppID
	}
	if LoginAppHash == "" {
		LoginAppHash = defaultAppHash
	}

	// Parse app ID
	appID, err := strconv.Atoi(LoginAppID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid app-id: %v\n", err)
		os.Exit(1)
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

	// Save credentials to config.json
	if err := config.SaveConfig(appID, LoginAppHash); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to save config: %v\n", err)
	}

	fmt.Printf("\nSession saved to: %s\n", sessionPath)

	// Try to reload session if server is running
	if reloadServerSession() {
		fmt.Println("Server session reloaded successfully")
	} else {
		fmt.Println("You can now use the serve command to start the server")
	}
}

// reloadServerSession attempts to reload the session on a running server.
// Returns true if server was running and session was reloaded.
func reloadServerSession() bool {
	client := ipc.NewClient("/tmp/agent-telegram.sock")

	// Check if server is running
	_, err := client.Call("status", nil)
	if err != nil {
		// Server not running
		return false
	}

	// Server is running, request session reload
	_, err = client.Call("reload_session", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to reload server session: %v\n", err)
		fmt.Fprintln(os.Stderr, "Please restart the server manually: agent-telegram stop && agent-telegram serve")
		return false
	}

	// Wait a bit for the reload to complete
	time.Sleep(2 * time.Second)

	// Verify the reload worked
	_, err = client.Call("status", nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: Server may still be reloading...")
	}

	return true
}

// createAuthService creates an auth service based on the provided parameters.
// If appID and appHash are provided, it uses NewServiceWithConfig.
// Otherwise, it uses NewService which loads from environment.
func createAuthService(ctx context.Context, appID int, appHash, phone string) (*auth.Service, error) {
	// Use provided credentials if available
	if appID != 0 && appHash != "" {
		return auth.NewServiceWithConfig(ctx, appID, appHash, phone)
	}
	// Use default service (loads from AGENT_TELEGRAM_* env vars)
	return auth.NewService(ctx)
}
