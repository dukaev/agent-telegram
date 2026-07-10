// Package chat provides commands for managing chats.
package chat

import (
	"agent-telegram/internal/cliutil"
)

// DeletePhotoCmd represents the delete-photo command.
var DeletePhotoCmd = cliutil.NewSimpleCommand(cliutil.SimpleCommandDef{
	Use:   "delete-photo",
	Short: "Delete the photo from a chat or channel",
	Long: `Delete the profile photo from a Telegram chat or channel.

Example:
  agent-telegram chat delete-photo --to @mychannel`,
	Method:  "delete_photo",
	Flags:   []cliutil.Flag{cliutil.ToFlag},
	Success: "Photo deleted successfully",
})
