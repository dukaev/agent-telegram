// Package cmd provides CLI commands.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

var (
	openLimit  int
	openOffset int
	openJSON   bool
)

// openCmd represents the open command.
var openCmd = &cobra.Command{
	Use:   "open @username",
	Short: "Open and view messages from a Telegram user/chat",
	Long: `Open and view messages from a Telegram user or chat by username.

Supports pagination with --limit and --offset flags.
Use --json flag for machine-readable output.`,
	Args: cobra.ExactArgs(1),
	Run:  runOpen,
}

func init() {
	rootCmd.AddCommand(openCmd)

	openCmd.Flags().IntVarP(&openLimit, "limit", "l", 10, "Number of messages to return (max 100)")
	openCmd.Flags().IntVarP(&openOffset, "offset", "o", 0, "Offset for pagination (message ID)")
	openCmd.Flags().BoolVarP(&openJSON, "json", "j", false, "Output as JSON")
}

func runOpen(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	username := args[0]

	// Validate and sanitize limit/offset
	if openLimit < 1 {
		openLimit = 1
	}
	if openLimit > 100 {
		openLimit = 100
	}
	if openOffset < 0 {
		openOffset = 0
	}

	client := ipc.NewClient(socketPath)
	result, err := client.Call("get_messages", map[string]any{
		"username": username,
		"limit":    openLimit,
		"offset":   openOffset,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if openJSON {
		printMessagesJSON(result)
	} else {
		printMessages(result)
	}
}

// printMessagesJSON prints the result as JSON.
func printMessagesJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printMessages prints messages in a human-readable format.
func printMessages(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	messages, _ := r["messages"].([]any)
	limit := r["limit"].(float64)
	offset := r["offset"].(float64)
	count := r["count"].(float64)
	username := r["username"].(string)

	fmt.Printf("Messages from @%s (limit: %.0f, offset: %.0f, count: %.0f):\n", username, limit, offset, count)
	fmt.Println()

	if len(messages) == 0 {
		fmt.Println("  No messages found.")
		return
	}

	for _, msg := range messages {
		printMessage(msg)
	}
}

// printMessage prints a single message.
func printMessage(msg any) {
	msgInfo, ok := msg.(map[string]any)
	if !ok {
		return
	}

	id, _ := msgInfo["id"].(float64)
	date, _ := msgInfo["date"].(float64)
	text, _ := msgInfo["text"].(string)
	out, _ := msgInfo["out"].(bool)
	fromName, hasFrom := msgInfo["from_name"].(string)

	// Format date
	tm := time.Unix(int64(date), 0)
	dateStr := tm.Format("2006-01-02 15:04:05")

	// Print direction arrow
	dir := "←"
	if out {
		dir = "→"
	}

	// Print header
	fmt.Printf("[%d] %s ", int64(id), dateStr)

	switch {
	case hasFrom && fromName != "":
		fmt.Printf("%s %s: ", dir, fromName)
	case out:
		fmt.Printf("%s You: ", dir)
	default:
		fmt.Printf("%s: ", dir)
	}

	// Print text
	if text != "" {
		fmt.Println(text)
	} else {
		fmt.Println("(empty or media message)")
	}
}
