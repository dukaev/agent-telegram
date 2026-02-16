// Package message provides commands for managing messages.
package message

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	clearTo      cliutil.Recipient
	clearHistory bool
	clearRevoke  bool
)

// ClearCmd represents the msg clear command.
var ClearCmd = &cobra.Command{
	Use:   "clear [message_ids]",
	Short: "Clear messages or history",
	Long: `Clear specific messages or entire chat history.

By default, clears specific messages by IDs.
Use --history to clear all chat history for a peer.
Use --revoke with --history to also delete for the other participant.

Examples:
  agent-telegram msg clear --to @user 12345,12346,12347
  agent-telegram msg clear --to @user --history
  agent-telegram msg clear --to @user --history --revoke`,
	Args: cobra.MaximumNArgs(1),
}

// AddClearCommand adds the clear command to the parent command.
func AddClearCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ClearCmd)

	ClearCmd.Flags().VarP(&clearTo, "to", "t", "Chat/user to clear messages from")
	ClearCmd.Flags().BoolVar(&clearHistory, "history", false, "Clear entire chat history")
	ClearCmd.Flags().BoolVar(&clearRevoke, "revoke", false, "Also delete for the other participant (with --history)")
	_ = ClearCmd.MarkFlagRequired("to")

	ClearCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(ClearCmd, false)

		if clearHistory {
			// Clear entire history
			params := map[string]any{
				"revoke": clearRevoke,
			}
			clearTo.AddToParams(params)

			result := runner.CallWithParams("clear_history", params)
			runner.PrintResult(result, func(any) {
				cliutil.PrintSuccessSummary(result, "Chat history cleared")
			})
			return
		}

		// Clear specific messages by IDs
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Error: message IDs required (e.g., 123 or 123,456,789), or use --history")
			os.Exit(1)
		}

		idStrs := strings.Split(args[0], ",")
		messageIDs := make([]int64, 0, len(idStrs))
		for _, idStr := range idStrs {
			id := runner.MustParseInt64(strings.TrimSpace(idStr))
			messageIDs = append(messageIDs, id)
		}

		params := map[string]any{
			"messageIds": messageIDs,
		}
		clearTo.AddToParams(params)

		result := runner.CallWithParams("clear_messages", params)
		runner.PrintResult(result, func(r any) {
			m, ok := r.(map[string]any)
			if !ok {
				fmt.Fprintln(os.Stderr, "Messages cleared successfully!")
				return
			}
			cleared := cliutil.ExtractFloat64(m, "cleared")
			fmt.Fprintf(os.Stderr, "Cleared %d message(s)\n", int(cleared))
		})
	}
}
