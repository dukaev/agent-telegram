// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// inspectInlineButtonsCmd represents the inspect-inline-buttons command.
var inspectInlineButtonsCmd = &cobra.Command{
	Use:   "inspect-inline-buttons @peer <message_id>",
	Short: "Inspect inline buttons in a message",
	Long: `List all inline buttons in a message.

Example: agent-telegram inspect-inline-buttons @user 123456`,
	Args: cobra.ExactArgs(2),
	Run:  runInspectInlineButtons,
}

func init() {
	rootCmd.AddCommand(inspectInlineButtonsCmd)
}

func runInspectInlineButtons(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
	result := runner.CallWithParams("inspect_inline_buttons", map[string]any{
		"peer":      args[0],
		"messageId": runner.MustParseInt64(args[1]),
	})

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

// pressInlineButtonCmd represents the press-inline-button command.
var pressInlineButtonCmd = &cobra.Command{
	Use:   "press-inline-button @peer <message_id> <button_index>",
	Short: "Press an inline button in a message",
	Long: `Press an inline button by its index.

Example: agent-telegram press-inline-button @user 123456 0`,
	Args: cobra.ExactArgs(3),
	Run:  runPressInlineButton,
}

func init() {
	rootCmd.AddCommand(pressInlineButtonCmd)
}

func runPressInlineButton(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(false)
	result := runner.CallWithParams("press_inline_button", map[string]any{
		"peer":        args[0],
		"messageId":   runner.MustParseInt64(args[1]),
		"buttonIndex": runner.MustParseInt(args[2]),
	})
	runner.PrintResult(result, nil)
}
