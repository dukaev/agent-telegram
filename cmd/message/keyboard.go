// Package message provides commands for managing messages.
package message

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	inspectKeyboardTo cliutil.Recipient
)

// InspectKeyboardCmd represents the inspect-keyboard command.
var InspectKeyboardCmd = &cobra.Command{
	GroupID: "message",
	Use:   "inspect-keyboard",
	Short: "Inspect reply keyboard buttons from a chat",
	Long: `Get the current reply keyboard from a chat.

Reply keyboard buttons are the buttons shown at the bottom of the chat interface,
different from inline buttons that appear with messages.

Use --to @username, --to username, or --to <chat_id> to specify the chat.`,
	Args: cobra.NoArgs,
}

// AddInspectKeyboardCommand adds the inspect-keyboard command to the root command.
func AddInspectKeyboardCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(InspectKeyboardCmd)

	InspectKeyboardCmd.Flags().VarP(&inspectKeyboardTo, "to", "t", "Recipient (@username, username, or chat ID)")
	_ = InspectKeyboardCmd.MarkFlagRequired("to")

	InspectKeyboardCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(InspectKeyboardCmd, true) // Always JSON

		params := map[string]any{}
		inspectKeyboardTo.AddToParams(params)

		result := runner.CallWithParams("inspect_reply_keyboard", params)

		// Output as JSON
		json.NewEncoder(os.Stdout).Encode(result)
	}
}
