// Package cmd provides CLI commands.
package cmd

import (
	"github.com/spf13/cobra"
)

var sendMessageFlags SendFlags

func init() {
	sendMessageCmd := sendMessageFlags.NewCommand(CommandConfig{
		Use:   "send-message <message>",
		Short: "Send a message to a Telegram peer",
		Long: `Send a message to a Telegram user or chat.

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
		Args: cobra.ExactArgs(1),
		Run:  runSendMessage,
	})
	rootCmd.AddCommand(sendMessageCmd)
}

func runSendMessage(_ *cobra.Command, args []string) {
	runner := sendMessageFlags.NewRunner()
	message := args[0]

	params := map[string]any{"message": message}
	sendMessageFlags.AddToParams(params)

	result := runner.CallWithParams("send_message", params)

	runner.PrintResult(result, func(r any) {
		FormatSuccess(r, "Message")
	})
}
