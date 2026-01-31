// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
	runner := NewRunnerFromRoot(getUpdatesJSON)
	result := runner.CallWithParams("get_updates", map[string]any{
		"limit": getUpdatesLimit,
	})

	runner.PrintResult(result, printUpdates)
}

// printUpdates prints updates in a human-readable format.
func printUpdates(result any) {
	resultMap, ok := ToMap(result)
	if !ok {
		return
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
	update, ok := ToMap(u)
	if !ok {
		return
	}

	fmt.Printf("  [%v] %v", update["id"], update["type"])
	printMessageText(update["data"])
	fmt.Println()
}

// printMessageText prints the message text if available.
func printMessageText(data any) {
	dataMap, ok := ToMap(data)
	if !ok {
		return
	}

	msg, ok := ToMap(dataMap["message"])
	if !ok {
		return
	}

	text, ok := msg["text"].(string)
	if !ok {
		return
	}

	fmt.Printf(": %s", text)
}
