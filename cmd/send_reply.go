// Package cmd provides CLI commands.
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	sendReplyJSON bool
)

// sendReplyCmd represents the send-reply command.
var sendReplyCmd = &cobra.Command{
	Use:   "send-reply @peer <message_id> <text>",
	Short: "Reply to a Telegram message",
	Long: `Send a reply to a specific message in a chat.

Example: agent-telegram send-reply @user 123456 "Thanks!"`,
	Args: cobra.ExactArgs(3),
	Run:  runSendReply,
}

func init() {
	sendReplyCmd.Flags().BoolVarP(&sendReplyJSON, "json", "j", false, "Output as JSON")
	rootCmd.AddCommand(sendReplyCmd)
}

func runSendReply(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(sendReplyJSON)
	peer := args[0]
	messageID := runner.MustParseInt64(args[1])
	text := args[2]

	result := runner.CallWithParams("send_reply", map[string]any{
		"peer":      peer,
		"messageId": messageID,
		"text":      text,
	})

	runner.PrintResult(result, func(r any) {
		FormatSuccess(r, "Reply")
	})
}
