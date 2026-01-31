// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	pinMessagePeer     string
	pinMessageUsername string
)

// pinMessageCmd represents the pin-message command.
var pinMessageCmd = &cobra.Command{
	Use:   "pin-message <message_id>",
	Short: "Pin a Telegram message",
	Long: `Pin a message in a chat.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(1),
	Run:  runPinMessage,
}

func init() {
	rootCmd.AddCommand(pinMessageCmd)
	pinMessageCmd.Flags().StringVarP(&pinMessagePeer, "peer", "p", "", "Peer (e.g., @username)")
	pinMessageCmd.Flags().StringVarP(&pinMessageUsername, "username", "u", "", "Username (without @)")
	pinMessageCmd.MarkFlagsOneRequired("peer", "username")
	pinMessageCmd.MarkFlagsMutuallyExclusive("peer", "username")
}

func runPinMessage(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
	params := map[string]any{
		"messageId": runner.MustParseInt64(args[0]),
	}
	if pinMessagePeer != "" {
		params["peer"] = pinMessagePeer
	} else {
		params["username"] = pinMessageUsername
	}
	result := runner.CallWithParams("pin_message", params)
	runner.PrintResult(result, func(any) {
		fmt.Printf("Message pinned successfully!\n")
	})
}

var (
	unpinMessagePeer     string
	unpinMessageUsername string
)

// unpinMessageCmd represents the unpin-message command.
var unpinMessageCmd = &cobra.Command{
	Use:   "unpin-message <message_id>",
	Short: "Unpin a Telegram message",
	Long: `Unpin a previously pinned message.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(1),
	Run:  runUnpinMessage,
}

func init() {
	rootCmd.AddCommand(unpinMessageCmd)
	unpinMessageCmd.Flags().StringVarP(&unpinMessagePeer, "peer", "p", "", "Peer (e.g., @username)")
	unpinMessageCmd.Flags().StringVarP(&unpinMessageUsername, "username", "u", "", "Username (without @)")
	unpinMessageCmd.MarkFlagsOneRequired("peer", "username")
	unpinMessageCmd.MarkFlagsMutuallyExclusive("peer", "username")
}

func runUnpinMessage(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
	params := map[string]any{
		"messageId": runner.MustParseInt64(args[0]),
	}
	if unpinMessagePeer != "" {
		params["peer"] = unpinMessagePeer
	} else {
		params["username"] = unpinMessageUsername
	}
	result := runner.CallWithParams("unpin_message", params)
	runner.PrintResult(result, func(any) {
		fmt.Printf("Message unpinned successfully!\n")
	})
}
