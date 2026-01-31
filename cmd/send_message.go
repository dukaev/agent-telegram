// Package cmd provides CLI commands.
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	sendMessageJSON bool
)

// sendMessageCmd represents the send-message command.
var sendMessageCmd = &cobra.Command{
	Use:   "send-message @peer <message>",
	Short: "Send a message to a Telegram peer",
	Long: `Send a message to a Telegram user or chat by username.

Supports sending to users, groups, and channels.`,
	Args: cobra.MinimumNArgs(2),
	Run:  runSendMessage,
}

func init() {
	rootCmd.AddCommand(sendMessageCmd)

	sendMessageCmd.Flags().BoolVarP(&sendMessageJSON, "json", "j", false, "Output as JSON")
}

func runSendMessage(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(sendMessageJSON)
	peer := args[0]
	message := args[1]

	result := runner.CallWithParams("send_message", map[string]any{
		"peer":    peer,
		"message": message,
	})

	runner.PrintResult(result, func(r any) {
		FormatSuccess(r, "Message")
	})
}
