// Package message provides commands for managing messages.
package message

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	pinMessageTo cliutil.Recipient
	pinDisable   bool
)

// PinCmd represents the pin command.
var PinCmd = &cobra.Command{
	GroupID: "message",
	Use:   "pin <message_id>",
	Short: "Pin or unpin a Telegram message",
	Long: `Pin or unpin a message in a chat.

Use --disable to unpin a previously pinned message.
Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
	Args: cobra.ExactArgs(1),
}

// AddPinCommand adds the pin command to the root command.
func AddPinCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(PinCmd)

	PinCmd.Flags().VarP(&pinMessageTo, "to", "t", "Recipient (@username, username, or chat ID)")
	PinCmd.Flags().BoolVarP(&pinDisable, "disable", "d", false, "Unpin the message")
	_ = PinCmd.MarkFlagRequired("to")

	PinCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(PinCmd, false)
		params := map[string]any{
			"messageId": runner.MustParseInt64(args[0]),
		}
		pinMessageTo.AddToParams(params)

		var method string
		var successMsg string
		if pinDisable {
			method = "unpin_message"
			successMsg = "Message unpinned successfully!"
		} else {
			method = "pin_message"
			successMsg = "Message pinned successfully!"
		}

		result := runner.CallWithParams(method, params)
		runner.PrintResult(result, func(any) {
			fmt.Printf("%s\n", successMsg)
		})
	}
}
