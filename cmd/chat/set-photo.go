// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// SetPhotoCmd represents the set-photo command.
var SetPhotoCmd = cliutil.NewSimpleCommand(cliutil.SimpleCommandDef{
	Use:     "set-photo",
	Short:   "Set the photo for a chat or channel",
	Long: `Set the profile photo for a Telegram chat or channel.

Example:
  agent-telegram chat set-photo --to @mychannel --file photo.jpg`,
	Method:  "set_photo",
	Flags:   []cliutil.Flag{cliutil.ToFlag, cliutil.FileFlag},
	Success: "Photo set successfully",
})

// AddSetPhotoCommand adds the set-photo command to the root command.
func AddSetPhotoCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(SetPhotoCmd)
}
