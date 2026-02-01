// Package message provides commands for managing messages.
package message

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var scheduledPeer string

// ScheduledCmd represents the scheduled command.
var ScheduledCmd = &cobra.Command{
	Use:   "scheduled",
	Short: "List scheduled messages",
	Long: `List all scheduled messages in a chat.

Example:
  agent-telegram msg scheduled --peer @channel`,
	Args: cobra.NoArgs,
}

// AddScheduledCommand adds the scheduled command to the parent command.
func AddScheduledCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ScheduledCmd)

	ScheduledCmd.Flags().StringVarP(&scheduledPeer, "peer", "p", "", "Chat/channel to get scheduled messages from")
	_ = ScheduledCmd.MarkFlagRequired("peer")

	ScheduledCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(ScheduledCmd, true)
		params := map[string]any{
			"peer": scheduledPeer,
		}

		result := runner.CallWithParams("get_scheduled_messages", params)
		//nolint:errchkjson // Output to stdout
		_ = json.NewEncoder(os.Stdout).Encode(result)

		// Print human-readable summary
		if r, ok := result.(map[string]any); ok {
			if count, ok := r["count"].(float64); ok {
				fmt.Fprintf(os.Stderr, "Found %d scheduled message(s)\n", int(count))
			}
		}
	}
}
