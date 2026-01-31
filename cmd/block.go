// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	blockJSON bool
)

// blockCmd represents the block command.
var blockCmd = &cobra.Command{
	Use:   "block @peer",
	Short: "Block a Telegram peer",
	Long: `Block a Telegram user or chat.

Blocked peers will not be able to send you messages or see your online status.

Example: agent-telegram block @user`,
	Args: cobra.ExactArgs(1),
	Run:  runBlock,
}

func init() {
	rootCmd.AddCommand(blockCmd)

	blockCmd.Flags().BoolVarP(&blockJSON, "json", "j", false, "Output as JSON")
}

func runBlock(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(blockJSON)
	result := runner.CallWithParams("block", map[string]any{
		"peer": args[0],
	})
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
