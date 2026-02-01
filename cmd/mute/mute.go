// Package mute provides commands for muting Telegram chats.
package mute

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// MuteConfig is the toggle command configuration for mute commands.
var MuteConfig = cliutil.ToggleCommandConfig{
	Use:   "mute",
	Short: "Mute or unmute a Telegram chat",
	Long: `Mute or unmute a Telegram chat to control notifications.

Muted chats will not send you notifications for new messages.
Use --disable to unmute a previously muted chat.

Use --to @username, --to username, or --to <chat_id> to specify the chat.`,
	EnableMethod:  "mute",
	DisableMethod: "unmute",
	EnableMsg:     "Chat muted successfully!",
	DisableMsg:    "Chat unmuted successfully!",
}

// MuteCmd represents the mute command.
var MuteCmd = cliutil.NewToggleCommand(MuteConfig)

func init() {
	MuteCmd.GroupID = "user"
}

// AddMuteCommand adds the mute command to the root command.
func AddMuteCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(MuteCmd)
}

// NewMuteCommand creates a new mute command (useful for subcommands).
func NewMuteCommand() *cobra.Command {
	return cliutil.NewToggleCommand(MuteConfig)
}
