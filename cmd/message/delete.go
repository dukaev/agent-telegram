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
	deleteTo     cliutil.Recipient
)

// DeleteCmd represents the delete command.
var DeleteCmd = &cobra.Command{
	GroupID: "message",
	Use:   "delete <message_id|id1,id2,...>",
	Short: "Delete Telegram message(s)",
	Long: `Delete messages from a chat.

- Single message ID: delete one message
- Comma-separated IDs: delete multiple messages

Use --to @username, --to username, or --to <chat_id> to specify the chat.
Use --revoke to also delete messages for the other participant (they won't be able to recover them).

To clear all history with a peer, use the history command with --revoke flag.

Examples:
  agent-telegram delete --to @user 123456
  agent-telegram delete --to @user 12345,12346,12347
  agent-telegram delete --to @user 12345 --revoke`,
	Args: cobra.MaximumNArgs(1),
}

// AddDeleteCommand adds the delete command to the root command.
func AddDeleteCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(DeleteCmd)

	DeleteCmd.Flags().BoolVar(&deleteRevoke, "revoke", false, "Delete for both parties")
	DeleteCmd.Flags().VarP(&deleteTo, "to", "t", "Recipient (@username, username, or chat ID)")
	_ = DeleteCmd.MarkFlagRequired("to")

	DeleteCmd.Run = func(_ *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Printf("Error: message_id or id1,id2,... required\n")
			return
		}

		runner := cliutil.NewRunnerFromCmd(DeleteCmd, false)
		arg := args[0]

		// Check if input contains comma (multiple IDs)
		if strings.Contains(arg, ",") {
			// Multiple IDs - use clear_messages
			idStrs := strings.Split(arg, ",")
			messageIDs := make([]int64, 0, len(idStrs))

			for _, idStr := range idStrs {
				id := runner.MustParseInt64(strings.TrimSpace(idStr))
				messageIDs = append(messageIDs, id)
			}

			params := map[string]any{
				"messageIds": messageIDs,
			}
			deleteTo.AddToParams(params)

			result := runner.CallWithParams("clear_messages", params)
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
			params := map[string]any{
				"messageId": runner.MustParseInt64(arg),
				"revoke":    deleteRevoke,
			}
			deleteTo.AddToParams(params)

			result := runner.CallWithParams("delete_message", params)
			runner.PrintResult(result, func(any) {
				fmt.Printf("Message deleted successfully!\n")
			})
		}
	}
}
