// Package chat provides commands for managing chats.
package chat

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	// ChatsLimit is the number of chats to return.
	listLimit int
	// ChatsOffset is the offset for pagination.
	listOffset int
	// ChatsSearch filters chats by title or username.
	listSearch string
	// ChatsType filters chats by type (user, chat, channel).
	listType string
)

// ListCmd represents the chat list command.
var ListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List Telegram chats",
	Long:    `List all Telegram chats with optional pagination and filtering.`,
	Example: `  agent-telegram chat list
  agent-telegram chat list --limit 50
  agent-telegram chat list --search mychannel
  agent-telegram chat list --type channel`,
}

// AddListCommand adds the list command to the parent command.
func AddListCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ListCmd)

	ListCmd.Flags().IntVarP(&listLimit, "limit", "l", 10, "Number of chats to return (max 100)")
	ListCmd.Flags().IntVarP(&listOffset, "offset", "o", 0, "Offset for pagination")
	ListCmd.Flags().StringVarP(&listSearch, "search", "q", "", "Filter by title or username (case-insensitive)")
	ListCmd.Flags().StringVarP(&listType, "type", "t", "", "Filter by type: user, chat, channel, or bot")

	ListCmd.Run = func(*cobra.Command, []string) {
		// Validate and sanitize limit/offset
		if listLimit < 1 {
			listLimit = 1
		}
		if listLimit > 100 {
			listLimit = 100
		}
		if listOffset < 0 {
			listOffset = 0
		}

		runner := cliutil.NewRunnerFromCmd(ListCmd, true) // Always JSON
		result := runner.CallWithParams("get_chats", map[string]any{
			"limit":  listLimit,
			"offset": listOffset,
		})

		filteredResult := filterChatsResult(result)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(filteredResult)
	}
}

// filterChatsResult filters and transforms the chats result.
//nolint:gocognit,funlen // Function requires filtering by multiple criteria
func filterChatsResult(result any) any {
	rMap, ok := result.(map[string]any)
	if !ok {
		return result
	}

	chats, _ := rMap["chats"].([]any)
	var filteredChats []any

	searchLower := strings.ToLower(listSearch)

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
		if listType != "" {
			// Special case for "bot" type - filter users that are bots
			if listType == "bot" {
				isBot, _ := chatInfo["bot"].(bool)
				if chatType != "user" || !isBot {
					continue
				}
			} else if chatType != listType {
				continue
			}
		}

		// Filter by search term if specified
		if listSearch != "" {
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
