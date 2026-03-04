package message

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	repliesTo      cliutil.Recipient
	repliesMessage int64
	repliesLimit   int
	repliesOffset  int
)

// RepliesCmd represents the msg replies command.
var RepliesCmd = &cobra.Command{
	Use:   "replies [peer]",
	Short: "Get replies (comments) to a channel post",
	Long: `Get replies/comments to a channel post from a Telegram channel.

Examples:
  agent-telegram msg replies @channel --message 12345
  agent-telegram msg replies --to @channel -m 12345 --limit 20`,
	Args: cobra.MaximumNArgs(1),
}

// AddRepliesCommand adds the replies command to the parent command.
func AddRepliesCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(RepliesCmd)

	RepliesCmd.Flags().VarP(&repliesTo, "to", "t", "Channel to get replies from")
	RepliesCmd.Flags().Int64VarP(&repliesMessage, "message", "m", 0, "Message ID to get replies for (required)")
	RepliesCmd.Flags().IntVar(&repliesLimit, "limit", 50, "Number of replies to fetch")
	RepliesCmd.Flags().IntVar(&repliesOffset, "offset", 0, "Offset message ID for pagination")

	RepliesCmd.Run = func(_ *cobra.Command, args []string) {
		if len(args) > 0 {
			_ = repliesTo.Set(args[0])
		}

		if repliesTo.Peer() == "" {
			fmt.Fprintln(os.Stderr, "Error: peer is required (positional or --to)")
			os.Exit(1)
		}
		if repliesMessage == 0 {
			fmt.Fprintln(os.Stderr, "Error: --message is required")
			os.Exit(1)
		}

		runner := cliutil.NewRunnerFromCmd(RepliesCmd, true)
		runner.SetIDKey("id")
		params := map[string]any{
			"peer":      repliesTo.Peer(),
			"messageId": repliesMessage,
			"limit":     repliesLimit,
		}
		if repliesOffset > 0 {
			params["offsetId"] = repliesOffset
		}

		result := runner.CallWithParams("get_replies", params)
		runner.PrintResult(result, nil)
	}
}
