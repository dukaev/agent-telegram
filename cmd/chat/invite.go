// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// InviteCmd represents the invite command.
var InviteCmd = cliutil.NewSimpleCommand(cliutil.SimpleCommandDef{
	Use:    "invite",
	Short:  "Invite users to a chat or channel",
	Long: `Invite users to a Telegram chat or channel.

Example:
  agent-telegram chat invite --to @mychannel --members @user1 --members @user2`,
	Method: "invite",
	Flags: []cliutil.Flag{
		cliutil.ToFlag,
		cliutil.MembersFlag,
	},
	Success: "Members invited successfully",
})

// AddInviteCommand adds the invite command to the root command.
func AddInviteCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(InviteCmd)
}
