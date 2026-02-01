// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// DeletePhotoCmd represents the delete-photo command.
var DeletePhotoCmd = cliutil.NewSimpleCommand(cliutil.SimpleCommandDef{
	Use:     "delete-photo",
	Short:   "Delete the photo from a chat or channel",
	Long: `Delete the profile photo from a Telegram chat or channel.

Example:
  agent-telegram chat delete-photo --peer @mychannel`,
	Method:  "delete_photo",
	Flags:   []cliutil.Flag{cliutil.PeerFlag},
	Success: "Photo deleted successfully",
})

// AddDeletePhotoCommand adds the delete-photo command to the root command.
func AddDeletePhotoCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(DeletePhotoCmd)
}
