// Package send provides commands for sending messages and media.
package send

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	updateMessageID int64
	updateText      string
)

// UpdateCmd represents the send update subcommand.
var UpdateCmd = &cobra.Command{
	Use:   "update <message_id> <new_text>",
	Short: "Edit a previously sent message",
	Long: `Edit a previously sent message.

Use --to @username, --to username, or --to <chat_id> to specify the recipient.

Example:
  send update --to @user 12345 "New text"`,
	Args: cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		runner := sendFlags.NewRunner()
		params := map[string]any{
			"messageId": runner.MustParseInt64(args[0]),
			"text":      args[1],
		}
		sendFlags.To.AddToParams(params)
		result := runner.CallWithParams("update_message", params)
		runner.PrintResult(result, func(any) {
			fmt.Printf("Message updated successfully!\n")
		})
	},
}

func init() {
	SendCmd.AddCommand(UpdateCmd)
	sendFlags.RegisterWithoutCaption(UpdateCmd)
}
