// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	clearHistoryJSON     bool
	clearHistoryRevoke   bool
	clearHistoryPeer     string
	clearHistoryUsername string
)

// clearHistoryCmd represents the clear-history command.
var clearHistoryCmd = &cobra.Command{
	Use:   "clear-history",
	Short: "Clear all chat history with a peer",
	Long: `Delete all message history with a Telegram user or chat.

Use --revoke to also delete messages for the other participant (they won't be able to recover them).

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.NoArgs,
	Run:  runClearHistory,
}

func init() {
	rootCmd.AddCommand(clearHistoryCmd)

	clearHistoryCmd.Flags().BoolVarP(&clearHistoryJSON, "json", "j", false, "Output as JSON")
	clearHistoryCmd.Flags().BoolVar(&clearHistoryRevoke, "revoke", false, "Delete for both parties")
	clearHistoryCmd.Flags().StringVarP(&clearHistoryPeer, "peer", "p", "", "Peer (e.g., @username)")
	clearHistoryCmd.Flags().StringVarP(&clearHistoryUsername, "username", "u", "", "Username (without @)")
	clearHistoryCmd.MarkFlagsOneRequired("peer", "username")
	clearHistoryCmd.MarkFlagsMutuallyExclusive("peer", "username")
}

func runClearHistory(_ *cobra.Command, _ []string) {
	runner := NewRunnerFromRoot(clearHistoryJSON)
	params := map[string]any{
		"revoke": clearHistoryRevoke,
	}
	if clearHistoryPeer != "" {
		params["peer"] = clearHistoryPeer
	} else {
		params["username"] = clearHistoryUsername
	}
	result := runner.CallWithParams("clear_history", params)
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
