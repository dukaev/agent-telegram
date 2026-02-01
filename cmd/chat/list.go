// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	listLimit  int
	listOffset int
	listSearch string
	listType   string
)

// ListCmd represents the chat list command.
var ListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List Telegram chats",
	Long:    `List all Telegram chats with optional pagination and filtering.

Shortcut: 'chats' command is an alias for 'chat list'.`,
	Example: `  agent-telegram chat list
  agent-telegram chats
  agent-telegram chat list --limit 50
  agent-telegram chat list --search mychannel
  agent-telegram chat list --type channel`,
}

// AddListCommand adds the list command to the parent command.
func AddListCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ListCmd)

	ListCmd.Flags().IntVarP(&listLimit, "limit", "l", 10, "Number of chats to return (max 100)")
	ListCmd.Flags().IntVarP(&listOffset, "offset", "o", 0, "Offset for pagination")
	ListCmd.Flags().StringVarP(&listSearch, "search", "Q", "", "Filter by title or username (case-insensitive)")
	ListCmd.Flags().StringVarP(&listType, "type", "T", "", "Filter by type: user, chat, channel, or bot")

	ListCmd.Run = func(*cobra.Command, []string) {
		cliutil.GetChats(ListCmd, listLimit, listOffset, listSearch, listType)
	}
}
