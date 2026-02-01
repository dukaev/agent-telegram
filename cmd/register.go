// Package cmd provides command registration.
package cmd

import (
	"agent-telegram/cmd/auth"
	"agent-telegram/cmd/chat"
	"agent-telegram/cmd/contacts"
	"agent-telegram/cmd/get"
	"agent-telegram/cmd/message"
	"agent-telegram/cmd/search"
	"agent-telegram/cmd/sys"
	"agent-telegram/cmd/user"
)

func init() {
	// Auth commands
	auth.AddLoginCommand(RootCmd)
	auth.AddLogoutCommand(RootCmd)

	// Get commands
	get.AddUserInfoCommand(RootCmd)
	get.AddUpdatesCommand(RootCmd)
	auth.AddOpenCommand(RootCmd)

	// Search commands
	search.AddSearchCommand(RootCmd)

	// Message commands
	message.AddMsgCommand(RootCmd)

	// User commands
	user.AddBlockCommand(RootCmd)

	// Contacts commands
	contacts.AddListContactsCommand(RootCmd)
	contacts.AddAddContactCommand(RootCmd)
	contacts.AddDeleteContactCommand(RootCmd)

	// Chat commands
	chat.AddChatCommand(RootCmd)

	// System commands
	sys.AddStatusCommand(RootCmd)
	sys.AddWatchCommand(RootCmd)
}
