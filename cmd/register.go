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
	auth.AddOpenCommand(RootCmd)

	// Get commands
	get.AddMeCommand(RootCmd)
	get.AddUserInfoCommand(RootCmd)
	get.AddUpdatesCommand(RootCmd)
	get.AddChatsCommand(RootCmd)

	// Send commands
	send.AddMessageCommand(RootCmd)
	send.AddReplyCommand(RootCmd)
	send.AddPhotoCommand(RootCmd)
	send.AddVideoCommand(RootCmd)
	send.AddFileCommand(RootCmd)
	send.AddContactCommand(RootCmd)
	send.AddLocationCommand(RootCmd)
	send.AddPollCommand(RootCmd)
	send.AddChecklistCommand(RootCmd)

	// Message commands
	message.AddDeleteCommand(RootCmd)
	message.AddUpdateCommand(RootCmd)
	message.AddPinCommand(RootCmd)
	message.AddUnpinCommand(RootCmd)
	message.AddInspectButtonsCommand(RootCmd)
	message.AddPressButtonCommand(RootCmd)
	message.AddAddReactionCommand(RootCmd)
	message.AddRemoveReactionCommand(RootCmd)
	message.AddListReactionsCommand(RootCmd)

	// User commands
	user.AddBlockCommand(RootCmd)
	user.AddUnblockCommand(RootCmd)
	user.AddProfileCommand(RootCmd)
	user.AddAvatarCommand(RootCmd)

	// Chat commands
	chat.AddClearCommand(RootCmd)
	chat.AddHistoryCommand(RootCmd)

	// Sys commands
	sys.AddPingCommand(RootCmd)
	sys.AddEchoCommand(RootCmd)
	sys.AddStatusCommand(RootCmd)
}
