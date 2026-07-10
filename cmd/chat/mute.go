// Package chat provides commands for managing chats.
package chat

import (
	"agent-telegram/cmd/mute"
)

// MuteCmd represents the mute command.
var MuteCmd = mute.NewMuteCommand()
