// Package cmd provides CLI commands.
package cmd

import (
	"github.com/spf13/cobra"
)

var sendReplyFlags SendFlags

func init() {
	sendReplyCmd := sendReplyFlags.NewCommand(CommandConfig{
		Use:   "send-reply <message_id> <text>",
		Short: "Reply to a Telegram message",
		Long: `Send a reply to a specific message in a chat.

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
		Args: cobra.ExactArgs(2),
		Run:  runSendReply,
	})
	rootCmd.AddCommand(sendReplyCmd)
}

func runSendReply(_ *cobra.Command, args []string) {
	runner := sendReplyFlags.NewRunner()
	messageID := runner.MustParseInt64(args[0])
	text := args[1]

	params := map[string]any{
		"messageId": messageID,
		"text":      text,
	}
	sendReplyFlags.AddToParams(params)

	result := runner.CallWithParams("send_reply", params)

	runner.PrintResult(result, func(r any) {
		FormatSuccess(r, "Reply")
	})
}
