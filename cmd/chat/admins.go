// Package chat provides commands for managing chats.
package chat

import (
	"agent-telegram/internal/cliutil"
)

// AdminsCmd represents the admins command.
var AdminsCmd = cliutil.NewListCommand(cliutil.ListCommandConfig{
	Use:   "admins",
	Short: "List admins in a chat or channel",
	Long: `List all administrators in a Telegram chat or channel.

Use --to @username or --to username to specify the chat/channel.
Use --limit to set the maximum number of admins to return (max 200).

Example:
  agent-telegram chat admins --to @mychannel --limit 20`,
	Method:    "get_admins",
	PrintFunc: cliutil.PrintAdmins,
	MaxLimit:  200,
	HasOffset: true,
})
