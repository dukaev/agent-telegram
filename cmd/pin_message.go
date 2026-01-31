// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// pinMessageCmd represents the pin-message command.
var pinMessageCmd = &cobra.Command{
	Use:   "pin-message @peer <message_id>",
	Short: "Pin a Telegram message",
	Long: `Pin a message in a chat.

Example: agent-telegram pin-message @user 123456`,
	Args: cobra.ExactArgs(2),
	Run:  runPinMessage,
}

func init() {
	rootCmd.AddCommand(pinMessageCmd)
}

func runPinMessage(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
	result := runner.CallWithParams("pin_message", map[string]any{
		"peer":      args[0],
		"messageId": runner.MustParseInt64(args[1]),
	})
	runner.PrintResult(result, func(any) {
		fmt.Printf("Message pinned successfully!\n")
	})
}

// unpinMessageCmd represents the unpin-message command.
var unpinMessageCmd = &cobra.Command{
	Use:   "unpin-message @peer <message_id>",
	Short: "Unpin a Telegram message",
	Long: `Unpin a previously pinned message.

Example: agent-telegram unpin-message @user 123456`,
	Args: cobra.ExactArgs(2),
	Run:  runUnpinMessage,
}

func init() {
	rootCmd.AddCommand(unpinMessageCmd)
}

func runUnpinMessage(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
	result := runner.CallWithParams("unpin_message", map[string]any{
		"peer":      args[0],
		"messageId": runner.MustParseInt64(args[1]),
	})
	runner.PrintResult(result, func(any) {
		fmt.Printf("Message unpinned successfully!\n")
	})
}
