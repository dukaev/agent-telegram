// Package chat provides commands for managing chats.
package chat

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	chatKeyboardTo cliutil.Recipient
)

// KeyboardCmd represents the chat keyboard command.
var KeyboardCmd = &cobra.Command{
	Use:   "keyboard [peer]",
	Short: "Inspect reply keyboard in a chat",
	Long: `Get the current reply keyboard from a chat.

Reply keyboard buttons are the buttons shown at the bottom of the chat interface,
different from inline buttons that appear with messages.

This only works with bots that have sent a reply keyboard.

Peer can be positional or via --to flag.

Examples:
  agent-telegram chat keyboard @bot
  agent-telegram chat keyboard --to @bot`,
	Args: cobra.MaximumNArgs(1),
}

// AddKeyboardCommand adds the keyboard command to the parent command.
func AddKeyboardCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(KeyboardCmd)

	KeyboardCmd.Flags().VarP(&chatKeyboardTo, "to", "t", "Recipient (@username, username, or chat ID)")

	KeyboardCmd.Run = func(_ *cobra.Command, args []string) {
		if len(args) > 0 {
			_ = chatKeyboardTo.Set(args[0])
		}

		if chatKeyboardTo.Peer() == "" {
			fmt.Fprintln(os.Stderr, "Error: peer is required (positional or --to)")
			os.Exit(1)
		}

		runner := cliutil.NewRunnerFromCmd(KeyboardCmd, true) // Always JSON

		params := map[string]any{}
		chatKeyboardTo.AddToParams(params)

		result := runner.CallWithParams("inspect_reply_keyboard", params)
		runner.PrintResult(result, nil)
	}
}
