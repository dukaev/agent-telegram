// Package message provides commands for managing messages.
package message

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	inspectKeyboardTo cliutil.Recipient
)

// InspectKeyboardCmd represents the inspect-keyboard command.
var InspectKeyboardCmd = &cobra.Command{
	Use:   "inspect-keyboard [peer]",
	Short: "Inspect reply keyboard buttons from a chat",
	Long: `Get the current reply keyboard from a chat.

Reply keyboard buttons are the buttons shown at the bottom of the chat interface,
different from inline buttons that appear with messages.

This only works with bots that have sent a reply keyboard.

Peer can be positional or via --to flag.`,
	Args: cobra.MaximumNArgs(1),
}

// AddInspectKeyboardCommand adds the inspect-keyboard command to the root command.
func AddInspectKeyboardCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(InspectKeyboardCmd)

	InspectKeyboardCmd.Flags().VarP(&inspectKeyboardTo, "to", "t", "Recipient (@username, username, or chat ID)")

	InspectKeyboardCmd.Run = func(_ *cobra.Command, args []string) {
		if len(args) > 0 {
			_ = inspectKeyboardTo.Set(args[0])
		}

		if inspectKeyboardTo.Peer() == "" {
			fmt.Fprintln(os.Stderr, "Error: peer is required (positional or --to)")
			os.Exit(1)
		}

		runner := cliutil.NewRunnerFromCmd(InspectKeyboardCmd, true) // Always JSON

		params := map[string]any{}
		inspectKeyboardTo.AddToParams(params)

		result := runner.CallWithParams("inspect_reply_keyboard", params)
		runner.PrintResult(result, nil)
	}
}
