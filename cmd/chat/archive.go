// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// ArchiveCmd represents the archive command.
var ArchiveCmd = cliutil.NewToggleCommand(cliutil.ToggleCommandConfig{
	Use:   "archive",
	Short: "Archive or unarchive a Telegram chat",
	Long: `Archive or unarchive a Telegram chat.

Archived chats are moved to the Archived folder and hidden from the main chat list.
Use --disable to unarchive a previously archived chat.

Use --to @username, --to username, or --to <chat_id> to specify the chat.`,
	EnableMethod:  "archive",
	DisableMethod: "unarchive",
	EnableMsg:     "Chat archived successfully!",
	DisableMsg:    "Chat unarchived successfully!",
})

// AddArchiveCommand adds the archive command to the root command.
func AddArchiveCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ArchiveCmd)
}
