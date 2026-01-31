// Package message provides commands for managing messages.
package message

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	updateMessageTo cliutil.Recipient
)

// UpdateCmd represents the update command.
var UpdateCmd = &cobra.Command{
	GroupID: "message",
	Use:   "update <message_id> <new_text>",
	Short: "Edit a Telegram message",
	Long: `Edit a previously sent message.

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
	Args: cobra.ExactArgs(2),
}

// AddUpdateCommand adds the update command to the root command.
func AddUpdateCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(UpdateCmd)

	UpdateCmd.Flags().VarP(&updateMessageTo, "to", "t", "Recipient (@username, username, or chat ID)")
	_ = UpdateCmd.MarkFlagRequired("to")

	UpdateCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(UpdateCmd, false)
		params := map[string]any{
			"messageId": runner.MustParseInt64(args[0]),
			"text":      args[1],
		}
		updateMessageTo.AddToParams(params)
		result := runner.CallWithParams("update_message", params)
		runner.PrintResult(result, func(any) {
			fmt.Printf("Message updated successfully!\n")
		})
	}
}
