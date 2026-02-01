// Package auth provides authentication commands.
package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// LogoutCmd represents the logout command.
var LogoutCmd = &cobra.Command{
	GroupID: "auth",
	Use:     "logout",
	Short:   "Logout and clear Telegram session",
	Long: `Logout from Telegram by removing the saved session.

This will:
  - Delete the saved session file
  - Remove authentication data
  - You will need to login again to use the app`,
	Run: runLogout,
}

// AddLogoutCommand adds the logout command to the root command.
func AddLogoutCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(LogoutCmd)
}

func runLogout(_ *cobra.Command, _ []string) {
	// Get the session path
	sessionPath := getSessionPath()

	// Check if session exists
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		fmt.Println("No active session found. You are already logged out.")
		return
	}

	// Delete the session file
	if err := os.Remove(sessionPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to delete session: %v\n", err)
		os.Exit(1)
	}

	// Also try to remove the user directory if it's empty
	userDir := filepath.Dir(sessionPath)
	if err := os.Remove(userDir); err != nil {
		// Directory might not be empty, that's okay
	}

	fmt.Println("Logged out successfully!")
	fmt.Printf("Session removed from: %s\n", sessionPath)
}

// getSessionPath returns the default session path.
func getSessionPath() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".agent-telegram", "user_1", "session.json")
}
