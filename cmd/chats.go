// Package cmd provides the root command and CLI configuration.
package cmd

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	chatsLimit  int
	chatsOffset int
	chatsSearch string
	chatsType   string
)

// ChatsCmd represents the chats command (alias for chat list).
var ChatsCmd = &cobra.Command{
	GroupID: GroupIDChat,
	Use:     "chats",
	Aliases: []string{"dialogs"},
	Short:   "List Telegram chats",
	Long:    `List all Telegram chats with optional pagination and filtering. Alias for 'chat list'.`,
	Example: `  agent-telegram chats
  agent-telegram chats --limit 50
  agent-telegram chats --search mychannel
  agent-telegram chats --type channel`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(cmd, true)
		runner.SetIDKey("peer")
		cliutil.GetChatsWithRunner(runner, chatsLimit, chatsOffset, chatsSearch, chatsType)
	},
}

func init() {
	ChatsCmd.Flags().IntVarP(&chatsLimit, "limit", "l", 10, "Number of chats to return (max 100)")
	ChatsCmd.Flags().IntVarP(&chatsOffset, "offset", "o", 0, "Offset for pagination")
	ChatsCmd.Flags().StringVarP(&chatsSearch, "search", "Q", "", "Filter by title or username (case-insensitive)")
	ChatsCmd.Flags().StringVarP(&chatsType, "type", "T", "", "Filter by type: user, chat, channel, or bot")

	RootCmd.AddCommand(ChatsCmd)
}
