// Package chat provides commands for managing chats.
package chat

import (
	"agent-telegram/internal/cliutil"
)

// CreateGroupCmd represents the create-group command.
var CreateGroupCmd = cliutil.NewSimpleCommand(cliutil.SimpleCommandDef{
	Use:   "create-group",
	Short: "Create a new group chat",
	Long: `Create a new Telegram group chat.

Example:
  agent-telegram chat create-group --title "My Group" --members @user1 --members @user2`,
	Method: "create_group",
	Flags: []cliutil.Flag{
		cliutil.TitleFlag,
		cliutil.MembersFlag,
	},
	Success: "Group created successfully",
})
