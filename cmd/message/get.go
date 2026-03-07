package message

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	getTo cliutil.Recipient
)

// GetCmd represents the msg get command.
var GetCmd = &cobra.Command{
	Use:   "get <message_id>",
	Short: "Get a single message by ID",
	Long: `Get a single message by ID from a Telegram chat.

Returns the message with text, reactions, reply info, and metadata.

Examples:
  agent-telegram msg get 12345 --to @channel
  agent-telegram msg get 12345 -t @username`,
	Args: cobra.ExactArgs(1),
}

// AddGetCommand adds the get command to the parent command.
func AddGetCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(GetCmd)

	GetCmd.Flags().VarP(&getTo, "to", "t", "Chat/channel to get the message from (required)")

	cliutil.RegisterMethod(GetCmd, "get_message")

	GetCmd.Run = func(_ *cobra.Command, args []string) {
		msgID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: invalid message ID")
			os.Exit(1)
		}

		if getTo.Peer() == "" {
			fmt.Fprintln(os.Stderr, "Error: --to is required")
			os.Exit(1)
		}

		runner := cliutil.NewRunnerFromCmd(GetCmd, true)
		params := map[string]any{
			"peer":      getTo.Peer(),
			"messageId": msgID,
		}

		result := runner.CallWithParams("get_message", params)
		runner.PrintResult(result, nil)
	}
}
