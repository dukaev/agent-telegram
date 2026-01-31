// Package get provides commands for retrieving information.
package get

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	// GetUpdatesLimit is the number of updates to get.
	GetUpdatesLimit int
	// GetUpdatesJSON enables JSON output.
	GetUpdatesJSON  bool
)

// UpdatesCmd represents the get-updates command.
var UpdatesCmd = &cobra.Command{
	Use:   "updates",
	Short: "Get Telegram updates (pops from store)",
	Long:  `Retrieve Telegram updates from the update store. This removes them from the store.`,
}

// AddUpdatesCommand adds the updates command to the root command.
func AddUpdatesCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(UpdatesCmd)

	UpdatesCmd.Flags().IntVarP(&GetUpdatesLimit, "limit", "l", 10, "Number of updates to get")
	UpdatesCmd.Flags().BoolVarP(&GetUpdatesJSON, "json", "j", false, "Output as JSON")
	UpdatesCmd.Run = func(*cobra.Command, []string) {
		runner := cliutil.NewRunnerFromCmd(UpdatesCmd, GetUpdatesJSON)
		result := runner.CallWithParams("get_updates", map[string]any{
			"limit": GetUpdatesLimit,
		})
		runner.PrintResult(result, printUpdates)
	}
}

// printUpdates prints updates in a human-readable format.
func printUpdates(result any) {
	resultMap, ok := cliutil.ToMap(result)
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
	update, ok := cliutil.ToMap(u)
	if !ok {
		return
	}

	fmt.Printf("  [%v] %v", update["id"], update["type"])
	printMessageText(update["data"])
	fmt.Println()
}

// printMessageText prints the message text if available.
func printMessageText(data any) {
	dataMap, ok := cliutil.ToMap(data)
	if !ok {
		return
	}

	msg, ok := cliutil.ToMap(dataMap["message"])
	if !ok {
		return
	}

	text, ok := msg["text"].(string)
	if !ok {
		return
	}

	fmt.Printf(": %s", text)
}
