// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	blockJSON     bool
	blockPeer     string
	blockUsername string
)

// blockCmd represents the block command.
var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Block a Telegram peer",
	Long: `Block a Telegram user or chat.

Blocked peers will not be able to send you messages or see your online status.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.NoArgs,
	Run:  runBlock,
}

func init() {
	rootCmd.AddCommand(blockCmd)

	blockCmd.Flags().BoolVarP(&blockJSON, "json", "j", false, "Output as JSON")
	blockCmd.Flags().StringVarP(&blockPeer, "peer", "p", "", "Peer (e.g., @username)")
	blockCmd.Flags().StringVarP(&blockUsername, "username", "u", "", "Username (without @)")
	blockCmd.MarkFlagsOneRequired("peer", "username")
	blockCmd.MarkFlagsMutuallyExclusive("peer", "username")
}

func runBlock(_ *cobra.Command, _ []string) {
	runner := NewRunnerFromRoot(blockJSON)
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
		peer := ExtractString(r, "peer")
		fmt.Printf("Peer blocked successfully!\n")
		fmt.Printf("  Peer: @%s\n", peer)
	})
}
