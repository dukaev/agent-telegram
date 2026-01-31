// Package message provides commands for managing messages.
package message

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	pinMessagePeer     string
	pinMessageUsername string
)

// PinCmd represents the pin-message command.
var PinCmd = &cobra.Command{
	Use:   "pin <message_id>",
	Short: "Pin a Telegram message",
	Long: `Pin a message in a chat.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(1),
}

// AddPinCommand adds the pin command to the root command.
func AddPinCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(PinCmd)

	PinCmd.Flags().StringVarP(&pinMessagePeer, "peer", "p", "", "Peer (e.g., @username)")
	PinCmd.Flags().StringVarP(&pinMessageUsername, "username", "u", "", "Username (without @)")
	PinCmd.MarkFlagsOneRequired("peer", "username")
	PinCmd.MarkFlagsMutuallyExclusive("peer", "username")

	PinCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(PinCmd, false)
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
}

var (
	unpinMessagePeer     string
	unpinMessageUsername string
)

// UnpinCmd represents the unpin-message command.
var UnpinCmd = &cobra.Command{
	Use:   "unpin <message_id>",
	Short: "Unpin a Telegram message",
	Long: `Unpin a previously pinned message.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(1),
}

// AddUnpinCommand adds the unpin command to the root command.
func AddUnpinCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(UnpinCmd)

	UnpinCmd.Flags().StringVarP(&unpinMessagePeer, "peer", "p", "", "Peer (e.g., @username)")
	UnpinCmd.Flags().StringVarP(&unpinMessageUsername, "username", "u", "", "Username (without @)")
	UnpinCmd.MarkFlagsOneRequired("peer", "username")
	UnpinCmd.MarkFlagsMutuallyExclusive("peer", "username")

	UnpinCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(UnpinCmd, false)
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
}
