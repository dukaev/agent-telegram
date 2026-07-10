// Package chat provides commands for managing chats.
package chat

import (
	"agent-telegram/internal/cliutil"
)

// EditTitleCmd represents the edit-title command.
var EditTitleCmd = cliutil.NewSimpleCommand(cliutil.SimpleCommandDef{
	Use:   "edit-title",
	Short: "Edit the title of a chat or channel",
	Long: `Edit the title of a Telegram chat or channel.

Example:
  agent-telegram chat edit-title --to @mychannel --title "New Title"`,
	Method:  "edit_title",
	Flags:   []cliutil.Flag{cliutil.ToFlag, cliutil.TitleFlag},
	Success: "Title updated successfully",
})
