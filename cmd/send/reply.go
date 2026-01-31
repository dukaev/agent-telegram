// Package send provides commands for sending messages and media.
package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	sendReplyFlags SendFlags
)

// ReplyCmd represents the send-reply command.
var ReplyCmd = &cobra.Command{
	Use:   "send-reply <message_id> <text>",
	Short: "Reply to a Telegram message",
	Long: `Send a reply to a specific message in a chat.

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
	Args: cobra.ExactArgs(2),
}

// AddReplyCommand adds the reply command to the root command.
func AddReplyCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ReplyCmd)

	ReplyCmd.Run = func(_ *cobra.Command, args []string) {
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
			cliutil.FormatSuccess(r, "Reply")
		})
	}

	sendReplyFlags.RegisterWithoutCaption(ReplyCmd)
}
