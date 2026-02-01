// Package open provides commands for opening chats and joining via invite links.
package open

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	// OpenLimit is the number of messages to return.
	openLimit int
	// OpenOffset is the offset for pagination.
	openOffset int
)

// OpenCmd represents the open command.
var OpenCmd = &cobra.Command{
	GroupID: "user",
	Use:     "open [@username|invite-link]",
	Short:   "Open a chat or join via invite link",
	Long: `Open and view messages from a Telegram user/chat, or join via invite link.

If the argument is a Telegram invite link, it will join the chat.
Otherwise, it will open and view messages from the user/chat.

Supports various invite link formats:
  - https://t.me/+hash
  - https://t.me/joinchat/hash
  - tg://join?invite=hash
  - +hash
  - hash

Supports pagination with --limit and --offset flags.

Examples:
  agent-telegram open @username
  agent-telegram open https://t.me/+abc123`,
	Args: cobra.ExactArgs(1),
}

// AddOpenCommand adds the open command to the root command.
func AddOpenCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(OpenCmd)

	OpenCmd.Flags().IntVarP(&openLimit, "limit", "l", 10, "Number of messages to return (max 100)")
	OpenCmd.Flags().IntVarP(&openOffset, "offset", "o", 0, "Offset for pagination")

	OpenCmd.Run = func(_ *cobra.Command, args []string) {
		arg := args[0]

		// Check if argument is an invite link
		if isInviteLink(arg) {
			runJoin(arg)
		} else {
			runOpen(arg)
		}
	}
}

// isInviteLink checks if the argument appears to be an invite link.
func isInviteLink(arg string) bool {
	arg = strings.TrimSpace(arg)

	// Check for common invite link patterns
	prefixes := []string{
		"https://t.me/+",
		"https://t.me/joinchat/",
		"tg://join?invite=",
		"+",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(arg, prefix) {
			return true
		}
	}

	// If it's a hash-like string (no special chars and reasonable length)
	// and starts with a digit or letter, it might be a hash
	if len(arg) >= 10 && len(arg) <= 50 {
		// Check if it contains only URL-safe characters
		for _, c := range arg {
			if !isURLSafe(c) {
				return false
			}
		}
		return true
	}

	return false
}

// isURLSafe checks if a character is URL-safe (no special chars except -, _, ~)
func isURLSafe(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_' || c == '~'
}

// runJoin executes the join chat logic.
func runJoin(inviteLink string) {
	runner := cliutil.NewRunnerFromCmd(OpenCmd, false)
	params := map[string]any{
		"inviteLink": inviteLink,
	}

	result := runner.CallWithParams("join_chat", params)
	runner.PrintResult(result, func(result any) {
		r, ok := result.(map[string]any)
		if !ok {
			fmt.Println("Joined chat successfully!")
			return
		}
		if chatID, ok := r["chatId"].(float64); ok {
			fmt.Printf("Joined chat successfully! Chat ID: %d\n", int64(chatID))
		} else {
			fmt.Println("Joined chat successfully!")
		}
		if title, ok := r["title"].(string); ok && title != "" {
			fmt.Printf("Title: %s\n", title)
		}
	})
}

// runOpen executes the open chat logic (view messages).
func runOpen(username string) {
	// Validate and sanitize limit/offset
	limit := openLimit
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}
	offset := openOffset
	if offset < 0 {
		offset = 0
	}

	runner := cliutil.NewRunnerFromCmd(OpenCmd, true) // Always JSON
	result := runner.CallWithParams("get_messages", map[string]any{
		"username": username,
		"limit":    limit,
		"offset":   offset,
	})

	// Output as JSON
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}
