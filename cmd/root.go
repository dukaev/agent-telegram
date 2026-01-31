// Package cmd provides the root command and CLI configuration.
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	// Import subpackages to trigger their init() registration
	_ "agent-telegram/cmd/auth"
	_ "agent-telegram/cmd/chat"
	_ "agent-telegram/cmd/get"
	_ "agent-telegram/cmd/message"
	_ "agent-telegram/cmd/send"
	_ "agent-telegram/cmd/sys"
	_ "agent-telegram/cmd/user"
)

var (
	version = "dev"

	// GroupIDAuth is the command group ID for authentication commands.
	GroupIDAuth = "auth"
	// GroupIDMessaging is the command group ID for messaging commands.
	GroupIDMessaging = "messaging"
	// GroupIDMessage is the command group ID for message management commands.
	GroupIDMessage = "message"
	// GroupIDUser is the command group ID for user commands.
	GroupIDUser = "user"
	// GroupIDChat is the command group ID for chat commands.
	GroupIDChat = "chat"
	// GroupIDGet is the command group ID for get commands.
	GroupIDGet = "get"
	// GroupIDSystem is the command group ID for system commands.
	GroupIDSystem = "system"
	// GroupIDServer is the command group ID for server commands.
	GroupIDServer = "server"
)

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "agent-telegram",
	Short: "Telegram IPC agent CLI",
	Long: `agent-telegram is a CLI tool for interacting with Telegram via IPC server.

It provides commands to:
  - Start an IPC server with Telegram client
  - Interactively login to Telegram
  - Query chats, messages, and user info
  - Send and receive Telegram messages`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add command groups
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDServer, Title: "Server"})
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDAuth, Title: "Authentication"})
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDMessaging, Title: "Messaging"})
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDMessage, Title: "Manage Messages"})
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDUser, Title: "User"})
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDChat, Title: "Chat"})
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDGet, Title: "Get"})
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDSystem, Title: "System"})

	// Global flags
	RootCmd.PersistentFlags().StringP("socket", "s", "/tmp/agent-telegram.sock", "Path to Unix socket")
}

