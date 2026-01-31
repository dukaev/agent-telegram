// Package cmd provides command registration.
package cmd

import (
	"agent-telegram/cmd/auth"
	"agent-telegram/cmd/chat"
	"agent-telegram/cmd/get"
	"agent-telegram/cmd/message"
	"agent-telegram/cmd/send"
	"agent-telegram/cmd/sys"
	"agent-telegram/cmd/user"
)

func init() {
	// Auth commands
	auth.AddLoginCommand(RootCmd)

	// Get commands
	get.AddUserInfoCommand(RootCmd)
	get.AddUpdatesCommand(RootCmd)
	get.AddChatsCommand(RootCmd)
	auth.AddOpenCommand(RootCmd)

	// Send commands
	send.AddSendCommand(RootCmd)

	// Message commands
	message.AddDeleteCommand(RootCmd)
	message.AddForwardCommand(RootCmd)
	message.AddPinCommand(RootCmd)
	message.AddInspectButtonsCommand(RootCmd)
	message.AddPressButtonCommand(RootCmd)
	message.AddInspectKeyboardCommand(RootCmd)
	message.AddReactionCommand(RootCmd)

	// User commands
	user.AddBlockCommand(RootCmd)

	// Chat commands
	chat.AddPinChatCommand(RootCmd)
	chat.AddMuteCommand(RootCmd)
	chat.AddArchiveCommand(RootCmd)

	// System commands
	sys.AddStatusCommand(RootCmd)
	sys.AddWatchCommand(RootCmd)
}
