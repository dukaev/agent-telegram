// Package search provides commands for searching Telegram content.
package search

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	globalQuery  string
	globalLimit  int
	globalType   string // bots, users, chats, channels, or empty for all

	inChatQuery  string
	inChatPeer   cliutil.Recipient
	inChatLimit  int
	inChatType   string // text, photos, videos, documents, links, audio, voice
	inChatOffset int
)

// SearchCmd represents the search command.
var SearchCmd = &cobra.Command{
	GroupID: "search",
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

	// Global search flags
	SearchGlobalCmd.Flags().IntVarP(&globalLimit, "limit", "l", 20, "Number of results (max 100)")
	SearchGlobalCmd.Flags().StringVarP(&globalType, "type", "t", "", "Filter by type: bots, users, chats, channels, or empty for all")

	// In-chat search flags
	SearchInChatCmd.Flags().VarP(&inChatPeer, "to", "t", "Recipient (@username, username, or chat ID)")
	SearchInChatCmd.Flags().IntVarP(&inChatLimit, "limit", "l", 20, "Number of results (max 100)")
	SearchInChatCmd.Flags().StringVarP(&inChatType, "type", "T", "", "Filter by message type: text, photos, videos, documents, links, audio, voice")
	SearchInChatCmd.Flags().IntVarP(&inChatOffset, "offset", "o", 0, "Offset for pagination (message ID)")
	_ = SearchInChatCmd.MarkFlagRequired("to")

	// Global search command handler
	SearchGlobalCmd.Run = func(_ *cobra.Command, args []string) {
		globalQuery = args[0]

		// Validate and sanitize limit
		if globalLimit < 1 {
			globalLimit = 1
		}
		if globalLimit > 100 {
			globalLimit = 100
		}

		runner := cliutil.NewRunnerFromCmd(SearchGlobalCmd, true) // Always JSON
		params := map[string]any{
			"query": globalQuery,
			"limit": globalLimit,
		}
		if globalType != "" {
			params["type"] = globalType
		}

		result := runner.CallWithParams("search_global", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)
	}

	// In-chat search command handler
	SearchInChatCmd.Run = func(_ *cobra.Command, args []string) {
		inChatQuery = args[0]

		// Validate and sanitize limit
		if inChatLimit < 1 {
			inChatLimit = 1
		}
		if inChatLimit > 100 {
			inChatLimit = 100
		}

		if inChatOffset < 0 {
			inChatOffset = 0
		}

		runner := cliutil.NewRunnerFromCmd(SearchInChatCmd, true) // Always JSON
		params := map[string]any{
			"peer":   inChatPeer.String(),
			"query":  inChatQuery,
			"limit":  inChatLimit,
			"offset": inChatOffset,
		}
		if inChatType != "" {
			params["type"] = inChatType
		}

		result := runner.CallWithParams("search_in_chat", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)
	}
}
