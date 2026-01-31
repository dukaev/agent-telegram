// Package get provides commands for retrieving information.
package get

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	// ChatsLimit is the number of chats to return.
	ChatsLimit  int
	// ChatsOffset is the offset for pagination.
	ChatsOffset int
	// ChatsJSON enables JSON output.
	ChatsJSON   bool
)

// ChatsCmd represents the chats command.
var ChatsCmd = &cobra.Command{
	Use:   "chats",
	Short: "List Telegram chats",
	Long:  `List all Telegram chats with optional pagination.`,
}

// AddChatsCommand adds the chats command to the root command.
func AddChatsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ChatsCmd)

	ChatsCmd.Flags().IntVarP(&ChatsLimit, "limit", "l", 10, "Number of chats to return (max 100)")
	ChatsCmd.Flags().IntVarP(&ChatsOffset, "offset", "o", 0, "Offset for pagination")
	ChatsCmd.Flags().BoolVarP(&ChatsJSON, "json", "j", false, "Output as JSON")
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

		runner := cliutil.NewRunnerFromCmd(ChatsCmd, ChatsJSON)
		result := runner.CallWithParams("get_chats", map[string]any{
			"limit":  ChatsLimit,
			"offset": ChatsOffset,
		})

		runner.PrintResult(result, func(r any) {
			rMap, ok := cliutil.ToMap(r)
			if !ok {
				return
			}

			chats, _ := rMap["chats"].([]any)
			limit := cliutil.ExtractFloat64(rMap, "limit")
			offset := cliutil.ExtractFloat64(rMap, "offset")
			count := cliutil.ExtractFloat64(rMap, "count")

			fmt.Printf("Chats (limit: %.0f, offset: %.0f, count: %.0f):\n", limit, offset, count)
			fmt.Println()

			for _, chat := range chats {
				printChat(chat)
			}
		})
	}
}

// printChat prints a single chat.
func printChat(chat any) {
	chatInfo, ok := cliutil.ToMap(chat)
	if !ok {
		return
	}

	chatType := cliutil.ExtractString(chatInfo, "type")
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
