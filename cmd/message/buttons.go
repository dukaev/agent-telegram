// Package message provides commands for managing messages.
package message

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	inspectInlineButtonsPeer     string
	inspectInlineButtonsUsername string
)

// InspectButtonsCmd represents the inspect-inline-buttons command.
var InspectButtonsCmd = &cobra.Command{
	Use:   "inspect-buttons <message_id>",
	Short: "Inspect inline buttons in a message",
	Long: `List all inline buttons in a message.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(1),
}

// AddInspectButtonsCommand adds the inspect-buttons command to the root command.
func AddInspectButtonsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(InspectButtonsCmd)

	InspectButtonsCmd.Flags().StringVarP(&inspectInlineButtonsPeer, "peer", "p", "", "Peer (e.g., @username)")
	InspectButtonsCmd.Flags().StringVarP(&inspectInlineButtonsUsername, "username", "u", "", "Username (without @)")
	InspectButtonsCmd.MarkFlagsOneRequired("peer", "username")
	InspectButtonsCmd.MarkFlagsMutuallyExclusive("peer", "username")

	InspectButtonsCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(InspectButtonsCmd, false)
		params := map[string]any{
			"messageId": runner.MustParseInt64(args[0]),
		}
		if inspectInlineButtonsPeer != "" {
			params["peer"] = inspectInlineButtonsPeer
		} else {
			params["username"] = inspectInlineButtonsUsername
		}
		result := runner.CallWithParams("inspect_inline_buttons", params)

		r, ok := result.(map[string]any)
		if !ok {
			runner.PrintResult(result, nil)
			return
		}
		buttons, _ := r["buttons"].([]any)

		fmt.Printf("Inline buttons (%d):\n", len(buttons))
		for i, btn := range buttons {
			b, _ := btn.(map[string]any)
			text, _ := b["text"].(string)
			data, _ := b["data"].(string)
			fmt.Printf("  [%d] %s", i, text)
			if data != "" {
				fmt.Printf(" -> %s", data)
			}
			fmt.Println()
		}
	}
}

var (
	pressInlineButtonPeer     string
	pressInlineButtonUsername string
)

// PressButtonCmd represents the press-inline-button command.
var PressButtonCmd = &cobra.Command{
	Use:   "press-button <message_id> <button_index>",
	Short: "Press an inline button in a message",
	Long: `Press an inline button by its index.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(2),
}

// AddPressButtonCommand adds the press-button command to the root command.
func AddPressButtonCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(PressButtonCmd)

	PressButtonCmd.Flags().StringVarP(&pressInlineButtonPeer, "peer", "p", "", "Peer (e.g., @username)")
	PressButtonCmd.Flags().StringVarP(&pressInlineButtonUsername, "username", "u", "", "Username (without @)")
	PressButtonCmd.MarkFlagsOneRequired("peer", "username")
	PressButtonCmd.MarkFlagsMutuallyExclusive("peer", "username")

	PressButtonCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(PressButtonCmd, false)
		params := map[string]any{
			"messageId":   runner.MustParseInt64(args[0]),
			"buttonIndex": runner.MustParseInt(args[1]),
		}
		if pressInlineButtonPeer != "" {
			params["peer"] = pressInlineButtonPeer
		} else {
			params["username"] = pressInlineButtonUsername
		}
		result := runner.CallWithParams("press_inline_button", params)
		runner.PrintResult(result, nil)
	}
}
