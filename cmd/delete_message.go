// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteMessageCmd represents the delete-message command.
var deleteMessageCmd = &cobra.Command{
	Use:   "delete-message <message_id>",
	Short: "Delete a Telegram message",
	Long: `Delete a specific message by ID.

Example: agent-telegram delete-message 123456`,
	Args: cobra.ExactArgs(1),
	Run:  runDeleteMessage,
}

func init() {
	rootCmd.AddCommand(deleteMessageCmd)
}

func runDeleteMessage(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
	result := runner.CallWithParams("delete_message", map[string]any{
		"messageId": runner.MustParseInt64(args[0]),
	})
	runner.PrintResult(result, func(any) {
		fmt.Printf("Message deleted successfully!\n")
	})
}
