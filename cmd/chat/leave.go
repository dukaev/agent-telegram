// Package chat provides commands for managing chats.
package chat

import (
	"agent-telegram/internal/cliutil"
)

// LeaveCmd represents the leave command.
var LeaveCmd = cliutil.NewSimpleCommand(cliutil.SimpleCommandDef{
	Use:     "leave",
	Short:   "Leave a chat or channel",
	Long:    "Leave a Telegram chat or channel.\n\nExample:\n  agent-telegram chat leave --to @mychannel",
	Method:  "leave",
	Flags:   []cliutil.Flag{cliutil.ToFlag},
	Success: "Left chat successfully",
})
