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
	getUpdatesLimit int
	getUpdatesJSON  bool
)

// getUpdatesCmd represents the get-updates command.
var getUpdatesCmd = &cobra.Command{
	Use:   "get-updates",
	Short: "Get Telegram updates (pops from store)",
	Long:  `Retrieve Telegram updates from the update store. This removes them from the store.`,
	Run: runGetUpdates,
}

func init() {
	rootCmd.AddCommand(getUpdatesCmd)

	getUpdatesCmd.Flags().IntVarP(&getUpdatesLimit, "limit", "l", 10, "Number of updates to get")
	getUpdatesCmd.Flags().BoolVarP(&getUpdatesJSON, "json", "j", false, "Output as JSON")
}

func runGetUpdates(_ *cobra.Command, _ []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")

	client := ipc.NewClient(socketPath)
	result, err := client.Call("get_updates", map[string]any{
		"limit": getUpdatesLimit,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if getUpdatesJSON {
		printJSON(result)
		return
	}

	printUpdates(result)
}

// printJSON prints the result as JSON.
func printJSON(result any) {
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}

// printUpdates prints updates in a human-readable format.
func printUpdates(result any) {
	resultMap, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response\n")
		os.Exit(1)
	}

	updates, ok := resultMap["updates"].([]any)
	if !ok || len(updates) == 0 {
		fmt.Println("No updates")
		return
	}

	fmt.Printf("Updates (%d):\n", len(updates))
	for _, u := range updates {
		printUpdate(u)
	}
}

// printUpdate prints a single update.
func printUpdate(u any) {
	update, ok := u.(map[string]any)
	if !ok {
		return
	}

	fmt.Printf("  [%v] %v", update["id"], update["type"])
	printMessageText(update["data"])
	fmt.Println()
}

// printMessageText prints the message text if available.
func printMessageText(data any) {
	dataMap, ok := data.(map[string]any)
	if !ok {
		return
	}

	msg, ok := dataMap["message"].(map[string]any)
	if !ok {
		return
	}

	text, ok := msg["text"].(string)
	if !ok {
		return
	}

	fmt.Printf(": %s", text)
}
