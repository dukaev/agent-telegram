// Package chat provides commands for managing chats.
package chat

import (
	"agent-telegram/internal/cliutil"
)

// PinChatCmd represents the pin command.
var PinChatCmd = cliutil.NewToggleCommand(cliutil.ToggleCommandConfig{
	Use:   "pin",
	Short: "Pin or unpin a chat in the dialog list",
	Long: `Pin or unpin a chat in your Telegram dialog list.

Pinned chats stay at the top of your chat list.

Use --disable to unpin a previously pinned chat.
Use --to @username, --to username, or --to <chat_id> to specify the chat.`,
	SingleMethod: "pin_chat",
	EnableMsg:    "Chat pinned successfully!",
	DisableMsg:   "Chat unpinned successfully!",
})
