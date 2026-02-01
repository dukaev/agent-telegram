// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"

	"agent-telegram/cmd/mute"
)

// MuteCmd represents the mute command.
var MuteCmd = mute.NewMuteCommand()

// AddMuteCommand adds the mute command to the root command.
func AddMuteCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(MuteCmd)
}
