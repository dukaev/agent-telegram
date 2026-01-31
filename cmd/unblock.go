// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	unblockJSON     bool
	unblockPeer     string
	unblockUsername string
)

// unblockCmd represents the unblock command.
var unblockCmd = &cobra.Command{
	Use:   "unblock",
	Short: "Unblock a Telegram peer",
	Long: `Unblock a Telegram user or chat.

Unblocked peers will be able to send you messages again.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.NoArgs,
	Run:  runUnblock,
}

func init() {
	unblockCmd.Flags().BoolVarP(&unblockJSON, "json", "j", false, "Output as JSON")
	unblockCmd.Flags().StringVarP(&unblockPeer, "peer", "p", "", "Peer (e.g., @username)")
	unblockCmd.Flags().StringVarP(&unblockUsername, "username", "u", "", "Username (without @)")
	unblockCmd.MarkFlagsOneRequired("peer", "username")
	unblockCmd.MarkFlagsMutuallyExclusive("peer", "username")
	rootCmd.AddCommand(unblockCmd)
}

func runUnblock(_ *cobra.Command, _ []string) {
	runner := NewRunnerFromRoot(unblockJSON)
	params := map[string]any{}
	if unblockPeer != "" {
		params["peer"] = unblockPeer
	} else {
		params["username"] = unblockUsername
	}
	result := runner.CallWithParams("unblock", params)
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
