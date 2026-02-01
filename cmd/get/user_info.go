// Package get provides commands for retrieving information.
package get

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// UserInfoCmd represents the info command.
var UserInfoCmd = &cobra.Command{
	GroupID: "get",
	Use:     "info [@username]",
	Short:   "Get information about a Telegram user",
	Long: `Get detailed information about a Telegram user by username.
If no username is provided, returns info about the current user (me).

This returns user ID, username, name, bio, verification status, etc.

Examples:
  agent-telegram info          # Get current user info
  agent-telegram info @username  # Get info about a specific user`,
	Args: cobra.MaximumNArgs(1),
}

// MyInfoCmd represents the my-info command (alias for info without args).
var MyInfoCmd = &cobra.Command{
	GroupID: "auth",
	Use:     "my-info",
	Short:   "Get your profile information",
	Long: `Get detailed information about your Telegram profile.

This returns your user ID, username, name, bio, verification status, etc.`,
	Args: cobra.NoArgs,
}

// AddUserInfoCommand adds the user-info command to the root command.
func AddUserInfoCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(UserInfoCmd)
	rootCmd.AddCommand(MyInfoCmd)

	UserInfoCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(UserInfoCmd, true)
		username := ""
		if len(args) > 0 {
			username = args[0]
		}
		cliutil.GetUserInfo(runner, username)
	}

	MyInfoCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(MyInfoCmd, true)
		cliutil.GetUserInfo(runner, "")
	}
}

// AddMyInfoCommand adds only the my-info command to the root command.
func AddMyInfoCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(MyInfoCmd)

	MyInfoCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(MyInfoCmd, true)
		cliutil.GetUserInfo(runner, "")
	}
}
