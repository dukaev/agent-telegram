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

// PinMessageCmd represents the pin-message command.
var PinMessageCmd = &cobra.Command{
	Use:   "pin-message <message_id>",
	Short: "Pin or unpin a Telegram message",
	Long: `Pin or unpin a message in a chat.

Use --disable to unpin a previously pinned message.
Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
	Args: cobra.ExactArgs(1),
}

// AddPinMessageCommand adds the pin-message command to the root command.
func AddPinMessageCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(PinMessageCmd)

	PinMessageCmd.Flags().VarP(&pinMessageTo, "to", "t", "Recipient (@username, username, or chat ID)")
	PinMessageCmd.Flags().BoolVarP(&pinDisable, "disable", "d", false, "Unpin the message")
	_ = PinMessageCmd.MarkFlagRequired("to")

	PinMessageCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(PinMessageCmd, false)
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
