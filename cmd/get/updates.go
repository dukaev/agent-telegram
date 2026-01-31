// Package get provides commands for retrieving information.
package get

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	// GetUpdatesLimit is the number of updates to get.
	GetUpdatesLimit int
	// GetUpdatesJSON enables JSON output.
	GetUpdatesJSON  bool
	// GetUpdatesTo filters updates by recipient.
	GetUpdatesTo cliutil.Recipient
)

// UpdatesCmd represents the get-updates command.
var UpdatesCmd = &cobra.Command{
	GroupID: "get",
	Use:   "updates",
	Short: "Get Telegram updates (pops from store)",
	Long:  `Retrieve Telegram updates from the update store. This removes them from the store.`,
}

// AddUpdatesCommand adds the updates command to the root command.
func AddUpdatesCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(UpdatesCmd)

	UpdatesCmd.Flags().IntVarP(&GetUpdatesLimit, "limit", "l", 10, "Number of updates to get")
	UpdatesCmd.Flags().BoolVarP(&GetUpdatesJSON, "json", "j", false, "Output as JSON")
	UpdatesCmd.Flags().VarP(&GetUpdatesTo, "to", "t", "Recipient (@username, username, or chat ID) to filter updates")
	UpdatesCmd.Run = func(*cobra.Command, []string) {
		runner := cliutil.NewRunnerFromCmd(UpdatesCmd, true) // Always JSON
		params := map[string]any{
			"limit": GetUpdatesLimit,
		}
		GetUpdatesTo.AddToParams(params)
		result := runner.CallWithParams("get_updates", params)
		// Output as JSON
		json.NewEncoder(os.Stdout).Encode(result)
	}
}
