// Package message provides commands for managing messages.
package message

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	updateMessagePeer     string
	updateMessageUsername string
)

// UpdateCmd represents the update-message command.
var UpdateCmd = &cobra.Command{
	Use:   "update <message_id> <new_text>",
	Short: "Edit a Telegram message",
	Long: `Edit a previously sent message.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(2),
}

// AddUpdateCommand adds the update command to the root command.
func AddUpdateCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(UpdateCmd)

	UpdateCmd.Flags().StringVarP(&updateMessagePeer, "peer", "p", "", "Peer (e.g., @username)")
	UpdateCmd.Flags().StringVarP(&updateMessageUsername, "username", "u", "", "Username (without @)")
	UpdateCmd.MarkFlagsOneRequired("peer", "username")
	UpdateCmd.MarkFlagsMutuallyExclusive("peer", "username")

	UpdateCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(UpdateCmd, false)
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
}
