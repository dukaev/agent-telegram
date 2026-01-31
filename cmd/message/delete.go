// Package message provides commands for managing messages.
package message

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// DeleteCmd represents the delete-message command.
var DeleteCmd = &cobra.Command{
	Use:   "delete <message_id>",
	Short: "Delete a Telegram message",
	Long: `Delete a specific message by ID.

Example: agent-telegram message delete 123456`,
	Args: cobra.ExactArgs(1),
}

// AddDeleteCommand adds the delete command to the root command.
func AddDeleteCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(DeleteCmd)

	DeleteCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(DeleteCmd, false)
		result := runner.CallWithParams("delete_message", map[string]any{
			"messageId": runner.MustParseInt64(args[0]),
		})
		runner.PrintResult(result, func(any) {
			fmt.Printf("Message deleted successfully!\n")
		})
	}
}
