// Package search provides commands for searching Telegram content.
package search

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	globalQuery  string
	globalLimit  int
	globalOffset int
	globalType   string // bots, users, chats, channels, or empty for all

	inChatQuery  string
	inChatPeer   cliutil.Recipient
	inChatLimit  int
	inChatType   string // text, photos, videos, documents, links, audio, voice
	inChatOffset int
)

// SearchCmd represents the search command.
var SearchCmd = &cobra.Command{
	GroupID: "chat",
	Use:     "search",
	Short:   "Search Telegram content",
	Long:    `Search for public chats, channels, bots globally, or search within a specific chat.`,
}

// SearchGlobalCmd represents the search global command.
var SearchGlobalCmd = &cobra.Command{
	Use:     "global <query>",
	Short:   "Search public chats, channels, and bots globally",
	Long: `Search for public chats, channels, and bots globally on Telegram.

Supports filtering by type:
- bots: search only for bots
- users: search only for users
- chats: search only for groups
- channels: search only for channels
- (empty): search everything

Usage: agent-telegram search global "myquery" [--type bots] [--limit 20]`,
	Args: cobra.ExactArgs(1),
}

// SearchInChatCmd represents the search in-chat command.
var SearchInChatCmd = &cobra.Command{
	Use:     "in-chat <query>",
	Short:   "Search for messages within a specific chat",
	Long: `Search for messages within a specific Telegram chat.

Supports filtering by message type:
- text: text messages only
- photos: photo messages only
- videos: video messages only
- documents: document/file messages only
- links: messages containing URLs
- audio: audio messages only
- voice: voice messages only
- (empty): all message types

Usage: agent-telegram search in-chat "myquery" --to @username [--type photos] [--limit 20] [--offset 0]`,
	Args: cobra.ExactArgs(1),
}

// AddSearchCommand adds the search command to the root command.
func AddSearchCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(SearchCmd)
	SearchCmd.AddCommand(SearchGlobalCmd)
	SearchCmd.AddCommand(SearchInChatCmd)

	setupGlobalSearchFlags()
	setupInChatSearchFlags()

	SearchGlobalCmd.Run = runSearchGlobal
	SearchInChatCmd.Run = runSearchInChat
}

// setupGlobalSearchFlags configures global search command flags.
func setupGlobalSearchFlags() {
	SearchGlobalCmd.Flags().IntVarP(&globalLimit, "limit", "l", cliutil.DefaultLimitMedium, "Number of results (max 100)")
	SearchGlobalCmd.Flags().IntVarP(&globalOffset, "offset", "o", 0, "Offset for pagination")
	SearchGlobalCmd.Flags().StringVar(&globalType, "type", "",
		"Filter by type: bots, users, chats, channels, or empty for all")
}

// setupInChatSearchFlags configures in-chat search command flags.
func setupInChatSearchFlags() {
	SearchInChatCmd.Flags().VarP(&inChatPeer, "to", "t", "Recipient (@username, username, or chat ID)")
	SearchInChatCmd.Flags().IntVarP(&inChatLimit, "limit", "l", cliutil.DefaultLimitMedium, "Number of results (max 100)")
	SearchInChatCmd.Flags().StringVar(&inChatType, "type", "",
		"Filter by message type: text, photos, videos, documents, links, audio, voice")
	SearchInChatCmd.Flags().IntVarP(&inChatOffset, "offset", "o", 0, "Offset for pagination (message ID)")
	_ = SearchInChatCmd.MarkFlagRequired("to")
}

// runSearchGlobal executes the global search command.
func runSearchGlobal(_ *cobra.Command, args []string) {
	globalQuery = args[0]

	pag := cliutil.NewPagination(globalLimit, globalOffset, cliutil.PaginationConfig{
		MaxLimit: cliutil.MaxLimitStandard,
	})

	runner := cliutil.NewRunnerFromCmd(SearchGlobalCmd, true)
	runner.SetIDKey("id")
	params := map[string]any{
		"query": globalQuery,
	}
	pag.ToParams(params, true)
	if globalType != "" {
		params["type"] = globalType
	}

	result := runner.CallWithParams("search_global", params)
	runner.PrintResult(result, nil)
}

// runSearchInChat executes the in-chat search command.
func runSearchInChat(_ *cobra.Command, args []string) {
	inChatQuery = args[0]

	pag := cliutil.NewPagination(inChatLimit, inChatOffset, cliutil.PaginationConfig{
		MaxLimit: cliutil.MaxLimitStandard,
	})

	runner := cliutil.NewRunnerFromCmd(SearchInChatCmd, true)
	runner.SetIDKey("id")
	params := map[string]any{
		"peer":  inChatPeer.String(),
		"query": inChatQuery,
	}
	pag.ToParams(params, true)
	if inChatType != "" {
		params["type"] = inChatType
	}

	result := runner.CallWithParams("search_in_chat", params)
	runner.PrintResult(result, nil)
}
