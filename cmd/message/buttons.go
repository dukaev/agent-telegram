// Package message provides commands for managing messages.
package message

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	inspectInlineButtonsPeer cliutil.Recipient
)

// InspectButtonsCmd represents the inspect-inline-buttons command.
var InspectButtonsCmd = &cobra.Command{
	GroupID: "message",
	Use:   "inspect-buttons <message_id>",
	Short: "Inspect inline buttons in a message",
	Long: `List all inline buttons in a message.

Use --peer @username, --peer username, or --peer <chat_id> to specify the recipient.`,
	Args: cobra.ExactArgs(1),
}

// AddInspectButtonsCommand adds the inspect-buttons command to the root command.
func AddInspectButtonsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(InspectButtonsCmd)

	InspectButtonsCmd.Flags().VarP(&inspectInlineButtonsPeer, "peer", "p", "Peer (@username, username, or chat ID)")
	_ = InspectButtonsCmd.MarkFlagRequired("peer")

	InspectButtonsCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(InspectButtonsCmd, false)
		params := map[string]any{
			"messageId": runner.MustParseInt64(args[0]),
		}
		inspectInlineButtonsPeer.AddToParams(params)
		result := runner.CallWithParams("inspect_inline_buttons", params)

		// Output as JSON
		json.NewEncoder(os.Stdout).Encode(result)
	}
}

var (
	pressInlineButtonPeer cliutil.Recipient
)

// PressButtonCmd represents the press-inline-button command.
var PressButtonCmd = &cobra.Command{
	GroupID: "message",
	Use:   "press-button <message_id> <button_index>",
	Short: "Press an inline button in a message",
	Long: `Press an inline button by its index.

Use --peer @username, --peer username, or --peer <chat_id> to specify the recipient.`,
	Args: cobra.ExactArgs(2),
}

// AddPressButtonCommand adds the press-button command to the root command.
func AddPressButtonCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(PressButtonCmd)

	PressButtonCmd.Flags().VarP(&pressInlineButtonPeer, "peer", "p", "Peer (@username, username, or chat ID)")
	_ = PressButtonCmd.MarkFlagRequired("peer")

	PressButtonCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(PressButtonCmd, false)
		params := map[string]any{
			"messageId":   runner.MustParseInt64(args[0]),
			"buttonIndex": runner.MustParseInt(args[1]),
		}
		pressInlineButtonPeer.AddToParams(params)
		result := runner.CallWithParams("press_inline_button", params)
		runner.PrintResult(result, nil)
	}
}
