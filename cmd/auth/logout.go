// Package auth provides authentication commands.
package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

// LogoutCmd represents the logout command.
var LogoutCmd = &cobra.Command{
	GroupID: "auth",
	Use:     "logout",
	Short:   "Logout and clear Telegram session",
	Long: `Logout from Telegram by shutting down the running in-memory session.

This will:
  - Ask the running IPC server to logout from Telegram
  - Clear the in-memory session on shutdown
  - You will need to authenticate again to use the app`,
	Run: runLogout,
}

// AddLogoutCommand adds the logout command to the root command.
func AddLogoutCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(LogoutCmd)
}

func runLogout(cmd *cobra.Command, _ []string) {
	socketPath, _ := cmd.Root().PersistentFlags().GetString("socket")
	if socketPath == "" {
		socketPath = "/tmp/agent-telegram.sock"
	}
	if _, err := ipc.NewClient(socketPath).Call("shutdown", nil); err != nil {
		fmt.Println("No active in-memory session found. You are already logged out.")
		return
	}
	fmt.Println("Logout requested. The server will logout from Telegram and stop.")
}
