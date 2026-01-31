// Package get provides commands for retrieving information.
package get

import (
	"encoding/json"
	"os"
	"strings"

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
	GroupID: "get",
	Use:     "chats",
	Short:   "List Telegram chats",
	Long:    `List all Telegram chats with optional pagination and filtering.`,
}

// AddChatsCommand adds the chats command to the root command.
func AddChatsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ChatsCmd)

	ChatsCmd.Flags().IntVarP(&ChatsLimit, "limit", "l", 10, "Number of chats to return (max 100)")
	ChatsCmd.Flags().IntVarP(&ChatsOffset, "offset", "o", 0, "Offset for pagination")
	ChatsCmd.Flags().StringVarP(&ChatsSearch, "search", "q", "", "Filter by title or username (case-insensitive)")
	ChatsCmd.Flags().StringVarP(&ChatsType, "type", "t", "", "Filter by type: user, chat, channel, or bot")

	ChatsCmd.Run = func(*cobra.Command, []string) {
		// Validate and sanitize limit/offset
		if ChatsLimit < 1 {
			ChatsLimit = 1
		}
		if ChatsLimit > 100 {
			ChatsLimit = 100
		}
		if ChatsOffset < 0 {
			ChatsOffset = 0
		}

		runner := cliutil.NewRunnerFromCmd(ChatsCmd, true) // Always JSON
		result := runner.CallWithParams("get_chats", map[string]any{
			"limit":  ChatsLimit,
			"offset": ChatsOffset,
		})

		filteredResult := filterChatsResult(result)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(filteredResult)
	}
}

// filterChatsResult filters and transforms the chats result.
//nolint:gocognit // Function requires filtering by multiple criteria
func filterChatsResult(result any) any {
	rMap, ok := result.(map[string]any)
	if !ok {
		return result
	}

	chats, _ := rMap["chats"].([]any)
	var filteredChats []any

	searchLower := strings.ToLower(ChatsSearch)

	for _, chat := range chats {
		chatInfo, ok := chat.(map[string]any)
		if !ok {
			continue
		}

		chatType := cliutil.ExtractString(chatInfo, "type")
		if chatType == "" {
			continue
		}

		// Filter by type if specified
		if ChatsType != "" {
			// Special case for "bot" type - filter users that are bots
			if ChatsType == "bot" {
				isBot, _ := chatInfo["bot"].(bool)
				if chatType != "user" || !isBot {
					continue
				}
			} else if chatType != ChatsType {
				continue
			}
		}

		// Filter by search term if specified
		if ChatsSearch != "" {
			title := cliutil.ExtractString(chatInfo, "title")
			username := cliutil.ExtractString(chatInfo, "peer")
			if !containsSearch(title, username, searchLower) {
				continue
			}
		}

		// Build simplified chat object with only requested fields
		simplified := map[string]any{
			"type": chatType,
		}

		// Add channel_id for channels
		if channelID, ok := chatInfo["channel_id"].(int64); ok && channelID != 0 {
			simplified["channel_id"] = channelID
		}
		if channelID, ok := chatInfo["channel_id"].(float64); ok && channelID != 0 {
			simplified["channel_id"] = int64(channelID)
		}

		// Add chat_id for groups
		if chatID, ok := chatInfo["chat_id"].(int64); ok && chatID != 0 {
			simplified["channel_id"] = chatID
		}
		if chatID, ok := chatInfo["chat_id"].(float64); ok && chatID != 0 {
			simplified["channel_id"] = int64(chatID)
		}

		// Add user_id for users
		if userID, ok := chatInfo["user_id"].(int64); ok && userID != 0 {
			simplified["channel_id"] = userID
		}
		if userID, ok := chatInfo["user_id"].(float64); ok && userID != 0 {
			simplified["channel_id"] = int64(userID)
		}

		// Add peer
		if peer := cliutil.ExtractString(chatInfo, "peer"); peer != "" {
			simplified["peer"] = peer
		}

		// Add title
		if title := cliutil.ExtractString(chatInfo, "title"); title != "" {
			simplified["title"] = title
		}

		// Add username (from peer field or username field)
		if username := cliutil.ExtractString(chatInfo, "username"); username != "" {
			simplified["username"] = username
		}

		filteredChats = append(filteredChats, simplified)
	}

	return map[string]any{
		"chats":  filteredChats,
		"limit":  cliutil.ExtractFloat64(rMap, "limit"),
		"offset": cliutil.ExtractFloat64(rMap, "offset"),
		"count":  len(filteredChats),
		"total":  cliutil.ExtractFloat64(rMap, "count"),
	}
}

// containsSearch checks if the title or username contains the search term.
func containsSearch(title, username, searchLower string) bool {
	if title != "" && strings.Contains(strings.ToLower(title), searchLower) {
		return true
	}
	if username != "" && strings.Contains(strings.ToLower(username), searchLower) {
		return true
	}
	return false
}
