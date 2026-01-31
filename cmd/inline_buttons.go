// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	inspectInlineButtonsPeer     string
	inspectInlineButtonsUsername string
)

// inspectInlineButtonsCmd represents the inspect-inline-buttons command.
var inspectInlineButtonsCmd = &cobra.Command{
	Use:   "inspect-inline-buttons <message_id>",
	Short: "Inspect inline buttons in a message",
	Long: `List all inline buttons in a message.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(1),
	Run:  runInspectInlineButtons,
}

func init() {
	rootCmd.AddCommand(inspectInlineButtonsCmd)
	inspectInlineButtonsCmd.Flags().StringVarP(&inspectInlineButtonsPeer, "peer", "p", "", "Peer (e.g., @username)")
	inspectInlineButtonsCmd.Flags().StringVarP(&inspectInlineButtonsUsername, "username", "u", "", "Username (without @)")
	inspectInlineButtonsCmd.MarkFlagsOneRequired("peer", "username")
	inspectInlineButtonsCmd.MarkFlagsMutuallyExclusive("peer", "username")
}

func runInspectInlineButtons(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
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

var (
	pressInlineButtonPeer     string
	pressInlineButtonUsername string
)

// pressInlineButtonCmd represents the press-inline-button command.
var pressInlineButtonCmd = &cobra.Command{
	Use:   "press-inline-button <message_id> <button_index>",
	Short: "Press an inline button in a message",
	Long: `Press an inline button by its index.

Use --peer @username or --username to specify the recipient.`,
	Args: cobra.ExactArgs(2),
	Run:  runPressInlineButton,
}

func init() {
	rootCmd.AddCommand(pressInlineButtonCmd)
	pressInlineButtonCmd.Flags().StringVarP(&pressInlineButtonPeer, "peer", "p", "", "Peer (e.g., @username)")
	pressInlineButtonCmd.Flags().StringVarP(&pressInlineButtonUsername, "username", "u", "", "Username (without @)")
	pressInlineButtonCmd.MarkFlagsOneRequired("peer", "username")
	pressInlineButtonCmd.MarkFlagsMutuallyExclusive("peer", "username")
}

func runPressInlineButton(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
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
