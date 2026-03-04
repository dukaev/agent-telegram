package message

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	replyCommentTo      cliutil.Recipient
	replyCommentMessage int64
	replyCommentID      int64
)

// ReplyCommentCmd represents the msg reply-comment command.
var ReplyCommentCmd = &cobra.Command{
	Use:   "reply-comment [peer] [text]",
	Short: "Reply to a comment in a channel post discussion",
	Long: `Reply to a specific comment in a channel post's discussion thread.

Examples:
  agent-telegram msg reply-comment @channel -m 36275 -c 36280 "Great point!"
  agent-telegram msg reply-comment --to @channel --message 36275 --comment 36280 "I agree"
  echo "My reply" | agent-telegram msg reply-comment @channel -m 36275 -c 36280`,
	Args: cobra.MaximumNArgs(2),
}

// AddReplyCommentCommand adds the reply-comment command to the parent command.
func AddReplyCommentCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ReplyCommentCmd)

	ReplyCommentCmd.Flags().VarP(&replyCommentTo, "to", "t", "Channel peer")
	ReplyCommentCmd.Flags().Int64VarP(&replyCommentMessage, "message", "m", 0, "Channel post message ID (required)")
	ReplyCommentCmd.Flags().Int64VarP(&replyCommentID, "comment", "c", 0, "Comment ID to reply to (required)")

	ReplyCommentCmd.Run = func(_ *cobra.Command, args []string) {
		var text string
		stdinText := cliutil.ReadStdinIfPiped()

		switch len(args) {
		case 2:
			_ = replyCommentTo.Set(args[0])
			text = args[1]
		case 1:
			if replyCommentTo.Peer() != "" {
				text = args[0]
			} else {
				_ = replyCommentTo.Set(args[0])
				text = stdinText
			}
		case 0:
			text = stdinText
		}

		if replyCommentTo.Peer() == "" {
			fmt.Fprintln(os.Stderr, "Error: peer is required (positional or --to)")
			os.Exit(1)
		}
		if replyCommentMessage == 0 {
			fmt.Fprintln(os.Stderr, "Error: --message is required")
			os.Exit(1)
		}
		if replyCommentID == 0 {
			fmt.Fprintln(os.Stderr, "Error: --comment is required")
			os.Exit(1)
		}
		if text == "" {
			fmt.Fprintln(os.Stderr, "Error: reply text is required (positional arg or stdin)")
			os.Exit(1)
		}

		runner := cliutil.NewRunnerFromCmd(ReplyCommentCmd, true)
		params := map[string]any{
			"peer":      replyCommentTo.Peer(),
			"messageId": replyCommentMessage,
			"commentId": replyCommentID,
			"text":      text,
		}

		result := runner.CallWithParams("reply_to_comment", params)
		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, "reply_to_comment")
		})
	}
}
