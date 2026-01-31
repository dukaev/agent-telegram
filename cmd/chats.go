// Package cmd provides CLI commands.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
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
	socketPath, _ := rootCmd.Flags().GetString("socket")

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

	client := ipc.NewClient(socketPath)
	result, err := client.Call("get_chats", map[string]int{
		"limit":  chatsLimit,
		"offset": chatsOffset,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if chatsJSON {
		printChatsJSON(result)
	} else {
		printChats(result)
	}
}

// printChatsJSON prints the result as JSON.
func printChatsJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printChats prints chats in a human-readable format.
func printChats(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	chats, _ := r["chats"].([]any)
	limit := r["limit"].(float64)
	offset := r["offset"].(float64)
	count := r["count"].(float64)

	fmt.Printf("Chats (limit: %.0f, offset: %.0f, count: %.0f):\n", limit, offset, count)
	fmt.Println()

	for _, chat := range chats {
		printChat(chat)
	}
}

// printChat prints a single chat.
func printChat(chat any) {
	chatInfo, ok := chat.(map[string]any)
	if !ok {
		return
	}

	chatType, hasType := chatInfo["type"].(string)
	if !hasType {
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
