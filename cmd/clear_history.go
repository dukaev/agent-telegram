// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	clearHistoryJSON bool
	clearHistoryRevoke bool
)

// clearHistoryCmd represents the clear-history command.
var clearHistoryCmd = &cobra.Command{
	Use:   "clear-history @peer",
	Short: "Clear all chat history with a peer",
	Long: `Delete all message history with a Telegram user or chat.

Use --revoke to also delete messages for the other participant (they won't be able to recover them).

Example: agent-telegram clear-history @user --revoke`,
	Args: cobra.ExactArgs(1),
	Run:  runClearHistory,
}

func init() {
	rootCmd.AddCommand(clearHistoryCmd)

	clearHistoryCmd.Flags().BoolVarP(&clearHistoryJSON, "json", "j", false, "Output as JSON")
	clearHistoryCmd.Flags().BoolVar(&clearHistoryRevoke, "revoke", false, "Delete for both parties")
}

func runClearHistory(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(clearHistoryJSON)
	result := runner.CallWithParams("clear_history", map[string]any{
		"peer":   args[0],
		"revoke": clearHistoryRevoke,
	})
	runner.PrintResult(result, func(result any) {
		r, ok := result.(map[string]any)
		if !ok {
			fmt.Printf("History cleared successfully!\n")
			return
		}
		peer := ExtractString(r, "peer")
		revoke, _ := r["revoke"].(bool)
		fmt.Printf("History cleared successfully!\n")
		fmt.Printf("  Peer: @%s\n", peer)
		if revoke {
			fmt.Printf("  Revoke: true (deleted for both parties)\n")
		}
	})
}
