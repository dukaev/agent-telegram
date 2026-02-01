// Package chat provides commands for managing chats.
package chat

import (
	"encoding/json"
	"os"

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
func filterChatsResult(result any) any {
	rMap, ok := result.(map[string]any)
	if !ok {
		return result
	}

	chats, _ := rMap["chats"].([]any)

	// Convert to []map[string]any for filtering
	chatMaps := make([]map[string]any, 0, len(chats))
	for _, chat := range chats {
		if chatInfo, ok := chat.(map[string]any); ok {
			chatMaps = append(chatMaps, chatInfo)
		}
	}

	// Apply generic filters
	filterOpts := cliutil.FilterOptions{
		Search: listSearch,
		Type:   listType,
	}
	filtered := cliutil.FilterItems(chatMaps, filterOpts)

	// Transform filtered items to simplified format
	simplifiedItems := make([]map[string]any, 0, len(filtered.Items))
	for _, item := range filtered.Items {
		simplified := simplifyChatItem(item)
		simplifiedItems = append(simplifiedItems, simplified)
	}

	return map[string]any{
		"chats":  simplifiedItems,
		"limit":  cliutil.ExtractFloat64(rMap, "limit"),
		"offset": cliutil.ExtractFloat64(rMap, "offset"),
		"count":  len(simplifiedItems),
		"total":  float64(filtered.Total),
	}
}

// simplifyChatItem creates a simplified chat object with key fields.
func simplifyChatItem(chatInfo map[string]any) map[string]any {
	simplified := map[string]any{
		"type": cliutil.ExtractString(chatInfo, "type"),
	}

	// Add ID based on type
	if channelID := extractInt64(chatInfo, "channel_id"); channelID != 0 {
		simplified["channel_id"] = channelID
	}
	if chatID := extractInt64(chatInfo, "chat_id"); chatID != 0 {
		simplified["chat_id"] = chatID
	}
	if userID := extractInt64(chatInfo, "user_id"); userID != 0 {
		simplified["user_id"] = userID
	}

	// Add common fields
	if peer := cliutil.ExtractString(chatInfo, "peer"); peer != "" {
		simplified["peer"] = peer
	}
	if title := cliutil.ExtractString(chatInfo, "title"); title != "" {
		simplified["title"] = title
	}
	if username := cliutil.ExtractString(chatInfo, "username"); username != "" {
		simplified["username"] = username
	}

	return simplified
}

// extractInt64 extracts an int64 value from a map, handling both int64 and float64 types.
func extractInt64(m map[string]any, key string) int64 {
	if v, ok := m[key].(int64); ok {
		return v
	}
	if v, ok := m[key].(float64); ok {
		return int64(v)
	}
	return 0
}
