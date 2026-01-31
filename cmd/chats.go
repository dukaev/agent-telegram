// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	chatsLimit  int
	chatsOffset int
	chatsJSON   bool
)

// chatsCmd represents the chats command.
var chatsCmd = &cobra.Command{
	Use:   "chats",
	Short: "List Telegram chats",
	Long:  `List all Telegram chats with optional pagination.`,
	Run: runChats,
}

func init() {
	rootCmd.AddCommand(chatsCmd)

	chatsCmd.Flags().IntVarP(&chatsLimit, "limit", "l", 10, "Number of chats to return (max 100)")
	chatsCmd.Flags().IntVarP(&chatsOffset, "offset", "o", 0, "Offset for pagination")
	chatsCmd.Flags().BoolVarP(&chatsJSON, "json", "j", false, "Output as JSON")
}

func runChats(_ *cobra.Command, _ []string) {
	// Validate and sanitize limit/offset
	if chatsLimit < 1 {
		chatsLimit = 1
	}
	if chatsLimit > 100 {
		chatsLimit = 100
	}
	if chatsOffset < 0 {
		chatsOffset = 0
	}

	runner := NewRunnerFromRoot(chatsJSON)
	result := runner.CallWithParams("get_chats", map[string]any{
		"limit":  chatsLimit,
		"offset": chatsOffset,
	})

	runner.PrintResult(result, func(r any) {
		rMap, ok := ToMap(r)
		if !ok {
			return
		}

		chats, _ := rMap["chats"].([]any)
		limit := ExtractFloat64(rMap, "limit")
		offset := ExtractFloat64(rMap, "offset")
		count := ExtractFloat64(rMap, "count")

		fmt.Printf("Chats (limit: %.0f, offset: %.0f, count: %.0f):\n", limit, offset, count)
		fmt.Println()

		for _, chat := range chats {
			printChat(chat)
		}
	})
}

// printChat prints a single chat.
func printChat(chat any) {
	chatInfo, ok := ToMap(chat)
	if !ok {
		return
	}

	chatType := ExtractString(chatInfo, "type")
	if chatType == "" {
		return
	}

	printChatByType(chatType, chatInfo)
	printUnreadCount(chatInfo)
}

// printChatByType prints chat info based on its type.
func printChatByType(chatType string, info map[string]any) {
	switch chatType {
	case "user":
		printUserChat(info)
	case "chat":
		printGroupChat(info)
	case "channel":
		printChannelChat(info)
	}
}

// printUserChat prints a user chat.
func printUserChat(info map[string]any) {
	name := buildUserName(info)
	fmt.Printf("  %s", name)
	if _, isBot := info["bot"].(bool); isBot {
		fmt.Print(" ")
	}
	fmt.Println()
}

// buildUserName builds a display name from user info.
func buildUserName(info map[string]any) string {
	var name string
	if fn, ok := info["first_name"].(string); ok {
		name = fn
		if ln, ok := info["last_name"].(string); ok && ln != "" {
			name += " " + ln
		}
	}
	if name == "" {
		if un, ok := info["username"].(string); ok {
			name = "@" + un
		}
	}
	return name
}

// printGroupChat prints a group chat.
func printGroupChat(info map[string]any) {
	title, ok := info["title"].(string)
	if !ok {
		return
	}
	fmt.Printf("  %s", title)
	if count, ok := info["participants_count"].(int); ok {
		fmt.Printf(" (%d members)", count)
	}
	fmt.Println()
}

// printChannelChat prints a channel chat.
func printChannelChat(info map[string]any) {
	title, ok := info["title"].(string)
	if !ok {
		return
	}
	fmt.Printf("  %s", title)
	if username, ok := info["username"].(string); ok && username != "" {
		fmt.Printf(" (@%s)", username)
	}
	if megagroup, ok := info["megagroup"].(bool); ok && megagroup {
		fmt.Print(" (group)")
	}
	fmt.Println()
}

// printUnreadCount prints unread count if any.
func printUnreadCount(info map[string]any) {
	if unread, ok := info["unread_count"].(int); ok && unread > 0 {
		fmt.Printf("     %d unread\n", unread)
	}
}
