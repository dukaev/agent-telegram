// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	updateMessagePeer     string
	updateMessageUsername string
)

// updateMessageCmd represents the update-message command.
var updateMessageCmd = &cobra.Command{
	Use:   "update-message <message_id> <new_text>",
	Short: "Edit a Telegram message",
	Long: `Edit a previously sent message.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(2),
	Run:  runUpdateMessage,
}

func init() {
	rootCmd.AddCommand(updateMessageCmd)
	updateMessageCmd.Flags().StringVarP(&updateMessagePeer, "peer", "p", "", "Peer (e.g., @username)")
	updateMessageCmd.Flags().StringVarP(&updateMessageUsername, "username", "u", "", "Username (without @)")
	updateMessageCmd.MarkFlagsOneRequired("peer", "username")
	updateMessageCmd.MarkFlagsMutuallyExclusive("peer", "username")
}

func runUpdateMessage(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
	params := map[string]any{
		"messageId": runner.MustParseInt64(args[0]),
		"text":      args[1],
	}
	if updateMessagePeer != "" {
		params["peer"] = updateMessagePeer
	} else {
		params["username"] = updateMessageUsername
	}
	result := runner.CallWithParams("update_message", params)
	runner.PrintResult(result, func(any) {
		fmt.Printf("Message updated successfully!\n")
	})
}
