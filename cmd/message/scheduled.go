// Package message provides commands for managing messages.
package message

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var scheduledTo cliutil.Recipient

// ScheduledCmd represents the scheduled command.
var ScheduledCmd = &cobra.Command{
	Use:   "scheduled",
	Short: "List scheduled messages",
	Long: `List all scheduled messages in a chat.

Example:
  agent-telegram msg scheduled --to @channel`,
	Args: cobra.NoArgs,
}

// AddScheduledCommand adds the scheduled command to the parent command.
func AddScheduledCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ScheduledCmd)

	ScheduledCmd.Flags().VarP(&scheduledTo, "to", "t", "Chat/channel to get scheduled messages from")
	_ = ScheduledCmd.MarkFlagRequired("to")

	ScheduledCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(ScheduledCmd, true)
		runner.SetIDKey("id")
		params := map[string]any{}
		scheduledTo.AddToParams(params)

		result := runner.CallWithParams("get_scheduled_messages", params)
		runner.PrintResult(result, func(r any) {
			if m, ok := r.(map[string]any); ok {
				if count, ok := m["count"].(float64); ok {
					fmt.Fprintf(os.Stderr, "Found %d scheduled message(s)\n", int(count))
				}
			}
		})
	}
}
