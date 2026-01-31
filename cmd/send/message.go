// Package send provides commands for sending messages and media.
package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	sendMessageFlags SendFlags
)

// MessageCmd represents the send-message command.
var MessageCmd = &cobra.Command{
	Use:   "send-message <message>",
	Short: "Send a message to a Telegram peer",
	Long: `Send a message to a Telegram user or chat.

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
	Args: cobra.ExactArgs(1),
}

// AddMessageCommand adds the message command to the root command.
func AddMessageCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(MessageCmd)

	MessageCmd.Run = func(_ *cobra.Command, args []string) {
		runner := sendMessageFlags.NewRunner()
		message := args[0]

		params := map[string]any{"message": message}
		sendMessageFlags.AddToParams(params)

		result := runner.CallWithParams("send_message", params)

		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, "Message")
		})
	}

	sendMessageFlags.RegisterWithoutCaption(MessageCmd)
}
