// Package chat provides commands for managing chats.
package chat

import (
	"agent-telegram/internal/cliutil"
)

// SubscribeCmd represents the subscribe command.
var SubscribeCmd = cliutil.NewSimpleCommand(cliutil.SimpleCommandDef{
	Use:   "subscribe",
	Short: "Subscribe to a public channel",
	Long: `Subscribe to a public Telegram channel by username.

Example:
  agent-telegram chat subscribe --channel @telegram`,
	Method:  "subscribe_channel",
	Flags:   []cliutil.Flag{cliutil.ChannelFlag},
	Success: "Subscribed to channel successfully",
})
