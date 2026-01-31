// Package cmd provides command registration.
package cmd

import (
	"github.com/spf13/cobra"

	"agent-telegram/cmd/auth"
	"agent-telegram/cmd/get"
	"agent-telegram/cmd/message"
	"agent-telegram/cmd/send"
	"agent-telegram/cmd/user"
)

func init() {
	// Auth commands
	auth.AddLoginCommand(RootCmd)
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDAuth, Title: "Authentication"})

	// Get commands
	get.AddUserInfoCommand(RootCmd)
	get.AddUpdatesCommand(RootCmd)
	get.AddChatsCommand(RootCmd)
	auth.AddOpenCommand(RootCmd)
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDGet, Title: "Information / Query"})

	// Send commands
	send.AddSendCommand(RootCmd)
	send.AddChecklistCommand(RootCmd)
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDMessaging, Title: "Send Messages"})

	// Message commands
	message.AddDeleteCommand(RootCmd)
	message.AddUpdateCommand(RootCmd)
	message.AddForwardCommand(RootCmd)
	message.AddPinCommand(RootCmd)
	message.AddInspectButtonsCommand(RootCmd)
	message.AddPressButtonCommand(RootCmd)
	message.AddInspectKeyboardCommand(RootCmd)
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDMessage, Title: "Manage Messages"})

	// User commands
	user.AddBlockCommand(RootCmd)
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDUser, Title: "User Management"})
}
