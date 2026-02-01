// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// JoinChatCmd represents the join command.
var JoinChatCmd = cliutil.NewSimpleCommand(cliutil.SimpleCommandDef{
	Use:    "join",
	Short:  "Join a chat using an invite link",
	Long: `Join a Telegram chat or channel using an invite link.

Example:
  agent-telegram chat join --link https://t.me/+abc123`,
	Method: "join_chat",
	Flags: []cliutil.Flag{
		{Name: "inviteLink", Short: "l", Usage: "Invite link to join", Required: true},
	},
	Success: "Joined chat successfully",
})

// AddJoinChatCommand adds the join command to the root command.
func AddJoinChatCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(JoinChatCmd)
}
