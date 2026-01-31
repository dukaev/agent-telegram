// Package cmd provides CLI commands.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

var (
	clearMessagesJSON bool
)

// clearMessagesCmd represents the clear-messages command.
var clearMessagesCmd = &cobra.Command{
	Use:   "clear-messages <id1,id2,...>",
	Short: "Clear specific Telegram messages",
	Long: `Delete specific messages by their IDs.

Message IDs can be obtained using the 'open' command.

Example: agent-telegram clear-messages 12345,12346,12347`,
	Args: cobra.ExactArgs(1),
	Run:  runClearMessages,
}

func init() {
	rootCmd.AddCommand(clearMessagesCmd)

	clearMessagesCmd.Flags().BoolVarP(&clearMessagesJSON, "json", "j", false, "Output as JSON")
}

func runClearMessages(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")

	idStrs := strings.Split(args[0], ",")
	messageIDs := make([]int64, 0, len(idStrs))

	for _, idStr := range idStrs {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid message ID '%s': %v\n", idStr, err)
			os.Exit(1)
		}
		messageIDs = append(messageIDs, id)
	}

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("clear_messages", map[string]any{
		"messageIds": messageIDs,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	if clearMessagesJSON {
		printClearMessagesJSON(result)
	} else {
		printClearMessagesResult(result)
	}
}

// printClearMessagesJSON prints the result as JSON.
func printClearMessagesJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printClearMessagesResult prints the result in a human-readable format.
func printClearMessagesResult(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	cleared, _ := r["cleared"].(float64)

	fmt.Printf("Messages cleared successfully!\n")
	fmt.Printf("  Cleared: %d messages\n", int(cleared))
}
