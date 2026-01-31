// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	unblockJSON bool
)

// unblockCmd represents the unblock command.
var unblockCmd = &cobra.Command{
	Use:   "unblock @peer",
	Short: "Unblock a Telegram peer",
	Long: `Unblock a Telegram user or chat.

Unblocked peers will be able to send you messages again.

Example: agent-telegram unblock @user`,
	Args: cobra.ExactArgs(1),
	Run:  runUnblock,
}

func init() {
	unblockCmd.Flags().BoolVarP(&unblockJSON, "json", "j", false, "Output as JSON")
	rootCmd.AddCommand(unblockCmd)
}

func runUnblock(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(unblockJSON)
	result := runner.CallWithParams("unblock", map[string]any{
		"peer": args[0],
	})
	runner.PrintResult(result, func(result any) {
		r, ok := result.(map[string]any)
		if !ok {
			fmt.Printf("Peer unblocked successfully!\n")
			return
		}
		peer := ExtractString(r, "peer")
		fmt.Printf("Peer unblocked successfully!\n")
		fmt.Printf("  Peer: @%s\n", peer)
	})
}
