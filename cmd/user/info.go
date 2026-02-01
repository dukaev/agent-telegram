// Package user provides commands for managing user-related operations.
package user

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// InfoCmd represents the user info command.
var InfoCmd = &cobra.Command{
	Use:   "info [@username]",
	Short: "Get information about a Telegram user",
	Long: `Get detailed information about a Telegram user by username.
If no username is provided, returns info about the current user (me).

This returns user ID, username, name, bio, verification status, etc.

Examples:
  agent-telegram user info          # Get current user info
  agent-telegram user info @username  # Get info about a specific user`,
	Args: cobra.MaximumNArgs(1),
}

// AddInfoCommand adds the info command to the parent command.
func AddInfoCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(InfoCmd)

	InfoCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(InfoCmd, true)
		username := ""
		if len(args) > 0 {
			username = args[0]
		}
		cliutil.GetUserInfo(runner, username)
	}
}
