// Package user provides commands for managing users.
//nolint:dupl // Similar to unblock, only text/messages differ
package user

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	blockJSON     bool
	blockPeer     string
	blockUsername string
)

// BlockCmd represents the block command.
var BlockCmd = &cobra.Command{
	Use:   "block",
	Short: "Block a Telegram peer",
	Long: `Block a Telegram user or chat.

Blocked peers will not be able to send you messages or see your online status.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.NoArgs,
}

// AddBlockCommand adds the block command to the root command.
func AddBlockCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(BlockCmd)

	BlockCmd.Flags().BoolVarP(&blockJSON, "json", "j", false, "Output as JSON")
	BlockCmd.Flags().StringVarP(&blockPeer, "peer", "p", "", "Peer (e.g., @username)")
	BlockCmd.Flags().StringVarP(&blockUsername, "username", "u", "", "Username (without @)")
	BlockCmd.MarkFlagsOneRequired("peer", "username")
	BlockCmd.MarkFlagsMutuallyExclusive("peer", "username")

	BlockCmd.Run = func(*cobra.Command, []string) {
		runner := cliutil.NewRunnerFromCmd(BlockCmd, blockJSON)
		params := map[string]any{}
		if blockPeer != "" {
			params["peer"] = blockPeer
		} else {
			params["username"] = blockUsername
		}
		result := runner.CallWithParams("block", params)
		runner.PrintResult(result, func(result any) {
			r, ok := result.(map[string]any)
			if !ok {
				fmt.Printf("Peer blocked successfully!\n")
				return
			}
			peer := cliutil.ExtractString(r, "peer")
			fmt.Printf("Peer blocked successfully!\n")
			fmt.Printf("  Peer: @%s\n", peer)
		})
	}
}
