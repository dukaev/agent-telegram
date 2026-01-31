// Package user provides commands for managing users.
//nolint:dupl // Similar to block, only text/messages differ
package user

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	// unblockJSON enables JSON output.
	unblockJSON     bool
	// unblockPeer is the peer to unblock (e.g., @username).
	unblockPeer     string
	// unblockUsername is the username to unblock (without @).
	unblockUsername string
)

// UnblockCmd represents the unblock command.
var UnblockCmd = &cobra.Command{
	Use:   "unblock",
	Short: "Unblock a Telegram peer",
	Long: `Unblock a Telegram user or chat.

Unblocked peers will be able to send you messages again.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.NoArgs,
}

// AddUnblockCommand adds the unblock command to the root command.
func AddUnblockCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(UnblockCmd)

	UnblockCmd.Flags().BoolVarP(&unblockJSON, "json", "j", false, "Output as JSON")
	UnblockCmd.Flags().StringVarP(&unblockPeer, "peer", "p", "", "Peer (e.g., @username)")
	UnblockCmd.Flags().StringVarP(&unblockUsername, "username", "u", "", "Username (without @)")
	UnblockCmd.MarkFlagsOneRequired("peer", "username")
	UnblockCmd.MarkFlagsMutuallyExclusive("peer", "username")

	UnblockCmd.Run = func(*cobra.Command, []string) {
		runner := cliutil.NewRunnerFromCmd(UnblockCmd, unblockJSON)
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
			peer := cliutil.ExtractString(r, "peer")
			fmt.Printf("Peer unblocked successfully!\n")
			fmt.Printf("  Peer: @%s\n", peer)
		})
	}
}
