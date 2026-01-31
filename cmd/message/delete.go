// Package message provides commands for managing messages.
package message

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	deleteRevoke bool
)

// DeleteCmd represents the delete command.
var DeleteCmd = &cobra.Command{
	GroupID: "message",
	Use:   "delete <message_id|id1,id2,...|@username>",
	Short: "Delete Telegram message(s) or clear history",
	Long: `Delete messages or clear chat history.

- Single message ID: delete one message
- Comma-separated IDs: delete multiple messages
- Username: clear all history with that user

Use --revoke to also delete messages for the other participant (they won't be able to recover them).

Examples:
  agent-telegram delete 123456
  agent-telegram delete 12345,12346,12347
  agent-telegram delete @username
  agent-telegram delete @username --revoke`,
	Args: cobra.MaximumNArgs(1),
}

// AddDeleteCommand adds the delete command to the root command.
func AddDeleteCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(DeleteCmd)

	DeleteCmd.Flags().BoolVar(&deleteRevoke, "revoke", false, "Delete for both parties (only with username)")

	DeleteCmd.Run = func(_ *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Printf("Error: argument required\n")
			return
		}

		runner := cliutil.NewRunnerFromCmd(DeleteCmd, false)
		arg := args[0]

		// Check if input is a username (starts with @)
		if strings.HasPrefix(arg, "@") {
			// Username - clear history
			params := map[string]any{
				"username": arg,
				"revoke":   deleteRevoke,
			}
			result := runner.CallWithParams("clear_history", params)
			runner.PrintResult(result, func(result any) {
				r, ok := result.(map[string]any)
				if !ok {
					fmt.Printf("History cleared successfully!\n")
					return
				}
				cleared := cliutil.ExtractFloat64(r, "cleared")
				fmt.Printf("History cleared successfully!\n")
				fmt.Printf("  Cleared: %d messages\n", int(cleared))
				if deleteRevoke {
					fmt.Printf("  Revoke: true (deleted for both parties)\n")
				}
			})
			return
		}

		// Check if input contains comma (multiple IDs)
		if strings.Contains(arg, ",") {
			// Multiple IDs - use clear_messages
			idStrs := strings.Split(arg, ",")
			messageIDs := make([]int64, 0, len(idStrs))

			for _, idStr := range idStrs {
				id := runner.MustParseInt64(strings.TrimSpace(idStr))
				messageIDs = append(messageIDs, id)
			}

			result := runner.CallWithParams("clear_messages", map[string]any{
				"messageIds": messageIDs,
			})
			runner.PrintResult(result, func(result any) {
				r, ok := result.(map[string]any)
				if !ok {
					fmt.Printf("Messages deleted successfully!\n")
					return
				}
				cleared := cliutil.ExtractFloat64(r, "cleared")
				fmt.Printf("Messages deleted successfully!\n")
				fmt.Printf("  Deleted: %d messages\n", int(cleared))
			})
		} else {
			// Single ID - use delete_message
			result := runner.CallWithParams("delete_message", map[string]any{
				"messageId": runner.MustParseInt64(arg),
			})
			runner.PrintResult(result, func(any) {
				fmt.Printf("Message deleted successfully!\n")
			})
		}
	}
}
