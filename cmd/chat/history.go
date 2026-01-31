// Package chat provides commands for managing chats.
package chat

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	clearHistoryJSON     bool
	clearHistoryRevoke   bool
	clearHistoryPeer     string
	clearHistoryUsername string
)

// HistoryCmd represents the clear-history command.
var HistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Clear all chat history with a peer",
	Long: `Delete all message history with a Telegram user or chat.

Use --revoke to also delete messages for the other participant (they won't be able to recover them).

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.NoArgs,
}

// AddHistoryCommand adds the history command to the root command.
func AddHistoryCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(HistoryCmd)

	HistoryCmd.Flags().BoolVarP(&clearHistoryJSON, "json", "j", false, "Output as JSON")
	HistoryCmd.Flags().BoolVar(&clearHistoryRevoke, "revoke", false, "Delete for both parties")
	HistoryCmd.Flags().StringVarP(&clearHistoryPeer, "peer", "p", "", "Peer (e.g., @username)")
	HistoryCmd.Flags().StringVarP(&clearHistoryUsername, "username", "u", "", "Username (without @)")
	HistoryCmd.MarkFlagsOneRequired("peer", "username")
	HistoryCmd.MarkFlagsMutuallyExclusive("peer", "username")

	HistoryCmd.Run = func(*cobra.Command, []string) {
		runner := cliutil.NewRunnerFromCmd(HistoryCmd, clearHistoryJSON)
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
			peer := cliutil.ExtractString(r, "peer")
			revoke, _ := r["revoke"].(bool)
			fmt.Printf("History cleared successfully!\n")
			fmt.Printf("  Peer: @%s\n", peer)
			if revoke {
				fmt.Printf("  Revoke: true (deleted for both parties)\n")
			}
		})
	}
}
