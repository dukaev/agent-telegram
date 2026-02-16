package cliutil

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// GetChats fetches and filters chats, outputting as JSON.
func GetChats(cmd *cobra.Command, limit, offset int, search, filterType string) {
	pag := NewPagination(limit, offset, PaginationConfig{
		MaxLimit: MaxLimitStandard,
	})

	runner := NewRunnerFromCmd(cmd, true)
	params := map[string]any{}
	pag.ToParams(params, true)
	result := runner.CallWithParams("get_chats", params)

	filteredResult := filterChatsResult(result, search, filterType)
	//nolint:errchkjson // Output to stdout, error handling not required
	_ = json.NewEncoder(os.Stdout).Encode(filteredResult)
}

// GetChatsWithRunner fetches and filters chats using the provided runner.
func GetChatsWithRunner(runner *Runner, limit, offset int, search, filterType string) {
	pag := NewPagination(limit, offset, PaginationConfig{
		MaxLimit: MaxLimitStandard,
	})

	params := map[string]any{}
	pag.ToParams(params, true)
	result := runner.CallWithParams("get_chats", params)

	filteredResult := filterChatsResult(result, search, filterType)
	runner.PrintResult(filteredResult, func(r any) {
		rMap, ok := r.(map[string]any)
		if !ok {
			return
		}
		chats, _ := rMap["chats"].([]any)
		fmt.Fprintf(os.Stderr, "Chats (%d):\n", len(chats))
		for _, c := range chats {
			chat, ok := c.(map[string]any)
			if !ok {
				continue
			}
			title := ExtractString(chat, "title")
			peer := ExtractString(chat, "peer")
			chatType := ExtractString(chat, "type")
			fmt.Fprintf(os.Stderr, "  - %s [%s] (%s)\n", title, peer, chatType)
		}
	})
}

// filterChatsResult filters and transforms the chats result.
func filterChatsResult(result any, search, filterType string) any {
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

	// Apply filters
	opts := FilterOptions{Search: search, Type: filterType}
	filtered := FilterItems(chatMaps, opts)

	// Simplify items
	simplified := make([]map[string]any, 0, len(filtered.Items))
	for _, item := range filtered.Items {
		simplified = append(simplified, simplifyChatItem(item))
	}

	return map[string]any{
		"chats":  simplified,
		"limit":  ExtractFloat64(rMap, "limit"),
		"offset": ExtractFloat64(rMap, "offset"),
		"count":  len(simplified),
		"total":  float64(filtered.Total),
	}
}

// simplifyChatItem creates a simplified chat object with key fields.
func simplifyChatItem(chatInfo map[string]any) map[string]any {
	simplified := map[string]any{
		"type": ExtractString(chatInfo, "type"),
	}

	// Add ID based on type
	idFields := []string{"channel_id", "chat_id", "user_id"}
	for _, field := range idFields {
		if id := ExtractInt64(chatInfo, field); id != 0 {
			simplified[field] = id
			break
		}
	}

	// Add common fields
	stringFields := []string{"peer", "title", "username"}
	for _, field := range stringFields {
		if v := ExtractString(chatInfo, field); v != "" {
			simplified[field] = v
		}
	}

	return simplified
}
