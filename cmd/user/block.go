// Package user provides commands for managing users.
package user

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	blockJSON    bool
	blockTo      cliutil.Recipient
	blockDisable bool
)

// BlockCmd represents the block command.
var BlockCmd = &cobra.Command{
	Use:   "ban",
	Short: "Block or unblock a Telegram peer",
	Long: `Block or unblock a Telegram user or chat.

Blocked peers will not be able to send you messages or see your online status.

Use --disable to unblock a peer.
Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
	Args: cobra.NoArgs,
}

// AddBlockCommand adds the ban command to the root command.
func AddBlockCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(BlockCmd)

	BlockCmd.Flags().BoolVarP(&blockJSON, "json", "j", false, "Output as JSON")
	BlockCmd.Flags().VarP(&blockTo, "to", "t", "Recipient (@username, username, or chat ID)")
	BlockCmd.Flags().BoolVarP(&blockDisable, "disable", "d", false, "Unblock the peer")
	_ = BlockCmd.MarkFlagRequired("to")

	BlockCmd.Run = func(*cobra.Command, []string) {
		runner := cliutil.NewRunnerFromCmd(BlockCmd, blockJSON)
		params := map[string]any{}
		blockTo.AddToParams(params)

		var method string
		var successMsg string
		if blockDisable {
			method = "unblock"
			successMsg = "Peer unbanned successfully!"
		} else {
			method = "block"
			successMsg = "Peer banned successfully!"
		}

		result := runner.CallWithParams(method, params)
		runner.PrintResult(result, func(result any) {
			r, ok := result.(map[string]any)
			if !ok {
				fmt.Fprintln(os.Stderr, successMsg)
				return
			}
			peer := cliutil.ExtractString(r, "peer")
			fmt.Fprintln(os.Stderr, successMsg)
			fmt.Fprintf(os.Stderr, "  Peer: %s\n", peer)
		})
	}
}
