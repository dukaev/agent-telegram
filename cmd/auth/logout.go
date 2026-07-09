// Package auth provides authentication commands.
package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
	"agent-telegram/internal/ipc"
	"agent-telegram/internal/paths"
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
		socketPath = paths.DefaultSocketPath
	}
	confirmed, _ := cmd.Flags().GetBool("confirm")
	if !confirmed {
		fmt.Println("Logout requires explicit confirmation: agent-telegram logout --confirm")
		cliutil.Exit(1)
		return
	}
	if _, err := ipc.NewClient(socketPath).CallWithOptions("logout", nil, ipc.CallOptions{Confirm: true}); err != nil {
		fmt.Printf("Logout failed: %s\n", err.Message)
		cliutil.Exit(1)
		return
	}
	fmt.Println("Logout requested. The server will logout from Telegram and stop.")
}
