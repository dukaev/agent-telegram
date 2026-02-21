// Package get provides commands for retrieving information.
package get

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	// ChatsLimit is the number of chats to return.
	ChatsLimit int
	// ChatsOffset is the offset for pagination.
	ChatsOffset int
	// ChatsSearch filters chats by title or username.
	ChatsSearch string
	// ChatsType filters chats by type (user, chat, channel).
	ChatsType string
)

// ChatsCmd represents the chats command.
var ChatsCmd = &cobra.Command{
	Use: "chats",
	Short:   "List Telegram chats",
	Long:    `List all Telegram chats with optional pagination and filtering.`,
}

// AddChatsCommand adds the chats command to the root command.
func AddChatsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ChatsCmd)

	ChatsCmd.Flags().IntVarP(&ChatsLimit, "limit", "l", 10, "Number of chats to return (max 100)")
	ChatsCmd.Flags().IntVarP(&ChatsOffset, "offset", "o", 0, "Offset for pagination")
	ChatsCmd.Flags().StringVarP(&ChatsSearch, "search", "Q", "", "Filter by title or username (case-insensitive)")
	ChatsCmd.Flags().StringVar(&ChatsType, "type", "", "Filter by type: user, chat, channel, or bot")

	ChatsCmd.Run = func(*cobra.Command, []string) {
		cliutil.GetChats(ChatsCmd, ChatsLimit, ChatsOffset, ChatsSearch, ChatsType)
	}
}
