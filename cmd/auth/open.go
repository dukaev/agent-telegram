// Package auth provides authentication commands.
package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

var (
	// OpenLimit is the number of messages to return.
	OpenLimit  int
	// OpenOffset is the offset for pagination.
	OpenOffset int
	// OpenJSON enables JSON output.
	OpenJSON   bool
)

// OpenCmd represents the open command.
var OpenCmd = &cobra.Command{
	GroupID: "get",
	Use:   "open @username",
	Short: "Open and view messages from a Telegram user/chat",
	Long: `Open and view messages from a Telegram user or chat by username.

Supports pagination with --limit and --offset flags.
Use --json flag for machine-readable output.`,
	Args: cobra.ExactArgs(1),
	Run:  runOpen,
}

// AddOpenCommand adds the open command to the root command.
func AddOpenCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(OpenCmd)

	OpenCmd.Flags().IntVarP(&OpenLimit, "limit", "l", 10, "Number of messages to return (max 100)")
	OpenCmd.Flags().IntVarP(&OpenOffset, "offset", "o", 0, "Offset for pagination (message ID)")
	OpenCmd.Flags().BoolVarP(&OpenJSON, "json", "j", false, "Output as JSON")
}

func runOpen(cmd *cobra.Command, args []string) {
	socketPath, _ := cmd.Flags().GetString("socket")
	username := args[0]

	// Validate and sanitize limit/offset
	if OpenLimit < 1 {
		OpenLimit = 1
	}
	if OpenLimit > 100 {
		OpenLimit = 100
	}
	if OpenOffset < 0 {
		OpenOffset = 0
	}

	client := ipc.NewClient(socketPath)
	result, err := client.Call("get_messages", map[string]any{
		"username": username,
		"limit":    OpenLimit,
		"offset":   OpenOffset,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Always output JSON
	printMessagesJSON(result)
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
