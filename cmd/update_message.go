// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// updateMessageCmd represents the update-message command.
var updateMessageCmd = &cobra.Command{
	Use:   "update-message @peer <message_id> <new_text>",
	Short: "Edit a Telegram message",
	Long: `Edit a previously sent message.

Example: agent-telegram update-message @user 123456 "Updated text"`,
	Args: cobra.ExactArgs(3),
	Run:  runUpdateMessage,
}

func init() {
	rootCmd.AddCommand(updateMessageCmd)
}

func runUpdateMessage(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
	result := runner.CallWithParams("update_message", map[string]any{
		"peer":      args[0],
		"messageId": runner.MustParseInt64(args[1]),
		"text":      args[2],
	})
	runner.PrintResult(result, func(any) {
		fmt.Printf("Message updated successfully!\n")
	})
}
