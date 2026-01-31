// Package message provides commands for managing messages.
package message

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	forwardFromPeer cliutil.Recipient
	forwardToPeer   cliutil.Recipient
)

// ForwardCmd represents the forward command.
var ForwardCmd = &cobra.Command{
	GroupID: "message",
	Use:   "forward <message_id>",
	Short: "Forward a Telegram message to another peer",
	Long: `Forward a message from one peer to another.

Use --from-peer @username, --from-peer username, or --from-peer <chat_id> to specify the source.
Use --to-peer @username, --to-peer username, or --to-peer <chat_id> to specify the destination.`,
	Args: cobra.ExactArgs(1),
}

// AddForwardCommand adds the forward command to the root command.
func AddForwardCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ForwardCmd)

	ForwardCmd.Flags().VarP(&forwardFromPeer, "from-peer", "f", "Source peer (@username, username, or chat ID)")
	ForwardCmd.Flags().VarP(&forwardToPeer, "to-peer", "t", "Destination peer (@username, username, or chat ID)")
	_ = ForwardCmd.MarkFlagRequired("from-peer")
	_ = ForwardCmd.MarkFlagRequired("to-peer")

	ForwardCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(ForwardCmd, true) // Always JSON
		params := map[string]any{
			"messageId": runner.MustParseInt64(args[0]),
		}
		forwardFromPeer.AddToParams(params)
		params["toPeer"] = forwardToPeer.Peer()
		result := runner.CallWithParams("forward_message", params)

		// Output as JSON
		json.NewEncoder(os.Stdout).Encode(result)
	}
}
