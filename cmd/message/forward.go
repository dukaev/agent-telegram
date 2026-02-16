// Package message provides commands for managing messages.
package message

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// ForwardCmd represents the forward command.
var ForwardCmd = &cobra.Command{
	Use:     "forward <message_id>",
	Short:   "Forward a Telegram message to another user or chat",
	Long: `Forward a message from one peer to another.

Use --from @username, --from username, or --from <chat_id> to specify the source.
Use --to @username, --to username, or --to <chat_id> to specify the destination.`,
	Args: cobra.ExactArgs(1),
}

// AddForwardCommand adds the forward command to the root command.
func AddForwardCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ForwardCmd)

	var fromPeer, toPeer cliutil.Recipient
	ForwardCmd.Flags().VarP(&fromPeer, "from", "f", "Source peer (@username, username, or chat ID)")
	ForwardCmd.Flags().VarP(&toPeer, "to", "t", "Destination peer (@username, username, or chat ID)")
	_ = ForwardCmd.MarkFlagRequired("from")
	_ = ForwardCmd.MarkFlagRequired("to")

	ForwardCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(ForwardCmd, true) // Always JSON
		params := map[string]any{
			"messageId": runner.MustParseInt64(args[0]),
		}
		params["fromPeer"] = fromPeer.Peer()
		params["toPeer"] = toPeer.Peer()
		result := runner.CallWithParams("forward_message", params)
		runner.PrintResult(result, nil)
	}
}
