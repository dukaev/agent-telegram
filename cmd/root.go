// Package cmd provides the root command and CLI configuration.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	// Import subpackages to trigger their init() registration
	_ "agent-telegram/cmd/auth"
	_ "agent-telegram/cmd/chat"
	_ "agent-telegram/cmd/contact"
	_ "agent-telegram/cmd/get"
	_ "agent-telegram/cmd/message"
	_ "agent-telegram/cmd/open"
	_ "agent-telegram/cmd/search"
	_ "agent-telegram/cmd/send"
	_ "agent-telegram/cmd/sys"
	_ "agent-telegram/cmd/user"
)

var (
	version   = "dev"
	fullHelp  bool

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

// printFullHelp prints all commands and subcommands recursively.
func printFullHelp() {
	fmt.Println("agent-telegram - Telegram IPC agent CLI")
	fmt.Println()
	fmt.Println("All commands and subcommands:")
	fmt.Println()

	// Group commands by their group
	groups := make(map[string][]*cobra.Command)

	for _, cmd := range RootCmd.Commands() {
		if !cmd.IsAvailableCommand() || cmd.IsAdditionalHelpTopicCommand() {
			continue
		}

		groupID := cmd.GroupID
		if groupID == "" {
			groupID = "Other"
		}
		groups[groupID] = append(groups[groupID], cmd)
	}

	// Define group order
	groupOrder := []string{"auth", "message", "chat", "server", "Other"}

	// Get group titles
	groupTitles := map[string]string{
		"server":  "Server",
		"auth":    "Authentication",
		"message": "Manage Messages",
		"chat":    "Chat",
		"Other":   "Other",
	}

	for _, groupID := range groupOrder {
		cmds, ok := groups[groupID]
		if !ok || len(cmds) == 0 {
			continue
		}

		title := groupTitles[groupID]
		fmt.Printf("%s\n", title)
		fmt.Println(strings.Repeat("-", len(title)))

		for _, cmd := range cmds {
			printCommandTree(cmd, "  ")
		}
		fmt.Println()
	}
}

// printCommandTree prints a command and all its subcommands recursively.
func printCommandTree(cmd *cobra.Command, prefix string) {
	fmt.Printf("%s%s\n", prefix, cmd.Name())

	// Print subcommands
	for _, subcmd := range cmd.Commands() {
		if !subcmd.IsAvailableCommand() || subcmd.IsAdditionalHelpTopicCommand() {
			continue
		}
		printCommandTree(subcmd, prefix+"  ")
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	// Parse flags manually to check for --full before cobra processes help
	if len(os.Args) > 1 && os.Args[1] == "--full" {
		printFullHelp()
		os.Exit(0)
	}

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
	RootCmd.PersistentFlags().BoolVar(&fullHelp, "full", false, "Show all commands and subcommands")
	RootCmd.PersistentFlags().Bool("dry-run", false, "Show what would be executed without actually running")
	RootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress status messages (data still goes to stdout)")
}

