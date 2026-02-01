// Package cmd provides command registration.
package cmd

import (
	"agent-telegram/cmd/auth"
	"agent-telegram/cmd/chat"
	"agent-telegram/cmd/contact"
	"agent-telegram/cmd/get"
	"agent-telegram/cmd/info"
	"agent-telegram/cmd/message"
	"agent-telegram/cmd/open"
	"agent-telegram/cmd/search"
	"agent-telegram/cmd/sys"
	"agent-telegram/cmd/user"
)

func init() {
	// Auth commands
	auth.AddLoginCommand(RootCmd)
	auth.AddLogoutCommand(RootCmd)
	get.AddMyInfoCommand(RootCmd)

	// Get commands
	get.AddUpdatesCommand(RootCmd)

	// Search commands
	search.AddSearchCommand(RootCmd)

	// Open command
	open.AddOpenCommand(RootCmd)

	// Info command
	info.AddInfoCommand(RootCmd)

	// Message commands
	message.AddMsgCommand(RootCmd)

	// User commands
	user.AddUserCommand(RootCmd)

	// Contact commands
	contact.AddContactCommand(RootCmd)

	// Chat commands
	chat.AddChatCommand(RootCmd)

	// System commands
	sys.AddStatusCommand(RootCmd)
	sys.AddWatchCommand(RootCmd)
}
