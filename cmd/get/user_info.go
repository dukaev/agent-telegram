// Package get provides commands for retrieving information.
package get

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// MyInfoCmd represents the my-info command.
var MyInfoCmd = &cobra.Command{
	GroupID: "auth",
	Use:     "my-info",
	Short:   "Get your profile information",
	Long: `Get detailed information about your Telegram profile.

This returns your user ID, username, name, bio, verification status, etc.`,
	Args: cobra.NoArgs,
}

// AddMyInfoCommand adds the my-info command to the root command.
func AddMyInfoCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(MyInfoCmd)

	MyInfoCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(MyInfoCmd, true)
		cliutil.GetUserInfo(runner, "")
	}
}
