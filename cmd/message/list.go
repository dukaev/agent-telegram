// Package message provides commands for managing messages.
package message

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	listTo     cliutil.Recipient
	listLimit  int
	listOffset int
)

// ListCmd represents the msg list command.
var ListCmd = &cobra.Command{
	Use:   "list [peer]",
	Short: "Get messages from a chat",
	Long: `Get messages from a Telegram chat by username, user ID, or special peer.

Examples:
  agent-telegram msg list @username
  agent-telegram msg list 272650856
  agent-telegram msg list me
  agent-telegram msg list --to @username --limit 20`,
	Args: cobra.MaximumNArgs(1),
}

// AddListCommand adds the list command to the parent command.
func AddListCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ListCmd)

	ListCmd.Flags().VarP(&listTo, "to", "t", "Chat/user to get messages from")
	ListCmd.Flags().IntVar(&listLimit, "limit", 10, "Number of messages to fetch")
	ListCmd.Flags().IntVar(&listOffset, "offset", 0, "Offset message ID")

	ListCmd.Run = func(_ *cobra.Command, args []string) {
		if len(args) > 0 {
			_ = listTo.Set(args[0])
		}

		if listTo.Peer() == "" {
			fmt.Fprintln(os.Stderr, "Error: peer is required (positional or --to)")
			os.Exit(1)
		}

		runner := cliutil.NewRunnerFromCmd(ListCmd, true)
		runner.SetIDKey("id")
		params := map[string]any{
			"username": listTo.Peer(),
			"limit":    listLimit,
		}
		if listOffset > 0 {
			params["offset"] = listOffset
		}

		result := runner.CallWithParams("get_messages", params)
		runner.PrintResult(result, nil)
	}
}
