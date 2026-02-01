// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// CreateChannelCmd represents the create-channel command.
var CreateChannelCmd = cliutil.NewSimpleCommand(cliutil.SimpleCommandDef{
	Use:   "create-channel",
	Short: "Create a new channel or supergroup",
	Long: `Create a new Telegram channel or supergroup with the specified title.

Use --title to set the channel title.
Use --description to set the channel description (optional).
Use --username to set a public username (optional).
Use --megagroup to create a supergroup instead of a channel.

Example:
  agent-telegram create-channel --title "My Channel" --description "My channel description"`,
	Method: "create_channel",
	Flags: []cliutil.Flag{
		cliutil.TitleFlag,
		cliutil.DescriptionFlag,
		cliutil.UsernameFlag,
		cliutil.MegagroupFlag,
	},
	Success: "Channel created successfully",
})

// AddCreateChannelCommand adds the create-channel command to the root command.
func AddCreateChannelCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(CreateChannelCmd)
}
