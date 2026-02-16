// Package message provides commands for managing messages.
package message

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	inspectInlineButtonsTo cliutil.Recipient
)

// InspectButtonsCmd represents the inspect-inline-buttons command.
var InspectButtonsCmd = &cobra.Command{
	Use:     "inspect-buttons <message_id>",
	Short:   "Inspect inline buttons in a message",
	Long: `List all inline buttons in a message.

Inline buttons are interactive buttons that appear attached to a specific message,
commonly used by bots for user interaction.

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
	Args: cobra.ExactArgs(1),
}

// AddInspectButtonsCommand adds the inspect-buttons command to the root command.
func AddInspectButtonsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(InspectButtonsCmd)

	InspectButtonsCmd.Flags().VarP(&inspectInlineButtonsTo, "to", "t", "Recipient (@username, username, or chat ID)")
	_ = InspectButtonsCmd.MarkFlagRequired("to")

	InspectButtonsCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(InspectButtonsCmd, false)
		params := map[string]any{
			"messageId": runner.MustParseInt64(args[0]),
		}
		inspectInlineButtonsTo.AddToParams(params)
		result := runner.CallWithParams("inspect_inline_buttons", params)
		runner.PrintResult(result, nil)
	}
}

var (
	pressInlineButtonTo cliutil.Recipient
)

// PressButtonCmd represents the press-inline-button command.
var PressButtonCmd = &cobra.Command{
	Use:     "press-button <message_id> <button_index>",
	Short:   "Press an inline button in a message",
	Long: `Press an inline button by its index.

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
	Args: cobra.ExactArgs(2),
}

// AddPressButtonCommand adds the press-button command to the root command.
func AddPressButtonCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(PressButtonCmd)

	PressButtonCmd.Flags().VarP(&pressInlineButtonTo, "to", "t", "Recipient (@username, username, or chat ID)")
	_ = PressButtonCmd.MarkFlagRequired("to")

	PressButtonCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(PressButtonCmd, false)
		params := map[string]any{
			"messageId":   runner.MustParseInt64(args[0]),
			"buttonIndex": runner.MustParseInt(args[1]),
		}
		pressInlineButtonTo.AddToParams(params)
		result := runner.CallWithParams("press_inline_button", params)
		runner.PrintResult(result, nil)
	}
}
