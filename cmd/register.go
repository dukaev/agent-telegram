// Package cmd provides command registration.
package cmd

import (
	"agent-telegram/cmd/auth"
	"agent-telegram/cmd/chat"
	"agent-telegram/cmd/contact"
	"agent-telegram/cmd/folders"
	"agent-telegram/cmd/get"
	"agent-telegram/cmd/gift"
	"agent-telegram/cmd/message"
	"agent-telegram/cmd/open"
	"agent-telegram/cmd/privacy"
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

	// Message commands
	message.AddMsgCommand(RootCmd)

	// User commands
	user.AddUserCommand(RootCmd)

	// Contact commands
	contact.AddContactCommand(RootCmd)

	// Chat commands
	chat.AddChatCommand(RootCmd)

	// Folders commands
	folders.AddFoldersCommand(RootCmd)

	// Gift commands
	gift.AddGiftCommand(RootCmd)

	// Privacy commands
	privacy.AddPrivacyCommand(RootCmd)

	// System commands
	sys.AddStatusCommand(RootCmd)
	sys.AddLLMsTxtCommand(RootCmd)
}
