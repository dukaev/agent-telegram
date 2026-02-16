// Package cmd provides the root command and CLI configuration.
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	// Import subpackages to trigger their init() registration
	_ "agent-telegram/cmd/auth"
	_ "agent-telegram/cmd/chat"
	_ "agent-telegram/cmd/contact"
	_ "agent-telegram/cmd/get"
	_ "agent-telegram/cmd/gift"
	_ "agent-telegram/cmd/message"
	_ "agent-telegram/cmd/open"
	_ "agent-telegram/cmd/search"
	_ "agent-telegram/cmd/send"
	_ "agent-telegram/cmd/sys"
	_ "agent-telegram/cmd/user"
)

var (
	version = "dev"

	// GroupIDAuth is the command group ID for authentication commands.
	GroupIDAuth = "auth"
	// GroupIDMessage is the command group ID for message management commands.
	GroupIDMessage = "message"
	// GroupIDChat is the command group ID for chat commands.
	GroupIDChat = "chat"
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
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDAuth, Title: "Authentication"})
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDMessage, Title: "Manage Messages"})
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDChat, Title: "Chat"})
	RootCmd.AddGroup(&cobra.Group{ID: GroupIDServer, Title: "Server"})

	// Global flags
	RootCmd.PersistentFlags().StringP("socket", "s", "/tmp/agent-telegram.sock", "Path to Unix socket")
	RootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress status messages (data still goes to stdout)")
	RootCmd.PersistentFlags().BoolP("json", "j", false, "Output result as JSON to stdout")
}

