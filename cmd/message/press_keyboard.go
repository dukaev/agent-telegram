// Package message provides commands for managing messages.
package message

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
	"agent-telegram/cmd/send"
)

var (
	pressKeyboardTo      cliutil.Recipient
	pressKeyboardWait    bool
	pressKeyboardTimeout time.Duration
)

// PressKeyboardCmd represents the press-keyboard command.
var PressKeyboardCmd = &cobra.Command{
	Use:   "press-keyboard [peer] <button_text_or_index>",
	Short: "Press a reply keyboard button",
	Long: `Press a reply keyboard button by its text or index.

Reply keyboard buttons are the buttons shown at the bottom of the chat interface.
This command inspects the keyboard, finds the button, and sends its text as a message.

Peer can be positional or via --to flag.

Examples:
  agent-telegram msg press-keyboard @bot "Menu"
  agent-telegram msg press-keyboard @bot 0
  agent-telegram msg press-keyboard --to @bot "Menu" --wait-reply`,
	Args: cobra.RangeArgs(1, 2),
}

// AddPressKeyboardCommand adds the press-keyboard command to the parent command.
func AddPressKeyboardCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(PressKeyboardCmd)

	PressKeyboardCmd.Flags().VarP(&pressKeyboardTo, "to", "t", "Recipient (@username, username, or chat ID)")
	PressKeyboardCmd.Flags().BoolVarP(&pressKeyboardWait, "wait-reply", "w", false, "Wait for a reply after pressing")
	PressKeyboardCmd.Flags().DurationVar(&pressKeyboardTimeout, "timeout", 10*time.Second, "Timeout for --wait-reply")

	PressKeyboardCmd.Run = func(_ *cobra.Command, args []string) {
		var buttonQuery string

		switch len(args) {
		case 2:
			_ = pressKeyboardTo.Set(args[0])
			buttonQuery = args[1]
		case 1:
			if pressKeyboardTo.Peer() != "" {
				buttonQuery = args[0]
			} else {
				fmt.Fprintln(os.Stderr, "Error: peer and button_text_or_index are required")
				fmt.Fprintln(os.Stderr, "Usage: press-keyboard [peer] <button_text_or_index>")
				os.Exit(1)
			}
		}

		if pressKeyboardTo.Peer() == "" {
			fmt.Fprintln(os.Stderr, "Error: peer is required (positional or --to)")
			os.Exit(1)
		}

		runner := cliutil.NewRunnerFromCmd(PressKeyboardCmd, false)

		// Step 1: Inspect keyboard
		kbParams := map[string]any{}
		pressKeyboardTo.AddToParams(kbParams)
		kbResult := runner.Call("inspect_reply_keyboard", kbParams)

		// Step 2: Find button text
		buttonText := findButtonText(kbResult, buttonQuery)

		// Step 3: Send the button text as a message
		sendParams := map[string]any{
			"message": buttonText,
		}
		pressKeyboardTo.AddToParams(sendParams)
		result := runner.Call("send_message", sendParams)

		// Step 4: Optionally wait for reply
		if pressKeyboardWait {
			send.HandleWaitReply(runner, pressKeyboardTo.Peer(), result, pressKeyboardTimeout)
			return
		}

		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, "press_keyboard")
		})
	}
}

// findButtonText finds a button by text match or index from the keyboard result.
func findButtonText(result any, query string) string {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintln(os.Stderr, "Error: unexpected keyboard result format")
		os.Exit(1)
	}

	found, _ := r["found"].(bool)
	if !found {
		fmt.Fprintln(os.Stderr, "Error: no reply keyboard found in this chat")
		os.Exit(1)
	}

	kb, ok := r["keyboard"].(map[string]any)
	if !ok {
		fmt.Fprintln(os.Stderr, "Error: unexpected keyboard format")
		os.Exit(1)
	}

	rows, ok := kb["rows"].([]any)
	if !ok || len(rows) == 0 {
		fmt.Fprintln(os.Stderr, "Error: keyboard has no buttons")
		os.Exit(1)
	}

	// Collect all buttons flat
	var buttons []string
	for _, row := range rows {
		rowItems, ok := row.([]any)
		if !ok {
			continue
		}
		for _, btn := range rowItems {
			btnMap, ok := btn.(map[string]any)
			if !ok {
				continue
			}
			if text, ok := btnMap["text"].(string); ok {
				buttons = append(buttons, text)
			}
		}
	}

	if len(buttons) == 0 {
		fmt.Fprintln(os.Stderr, "Error: keyboard has no text buttons")
		os.Exit(1)
	}

	// Try numeric index first
	if idx, err := strconv.Atoi(query); err == nil {
		if idx < 0 || idx >= len(buttons) {
			fmt.Fprintf(os.Stderr, "Error: button index %d out of range (0-%d)\n", idx, len(buttons)-1)
			os.Exit(1)
		}
		return buttons[idx]
	}

	// Try exact text match
	for _, text := range buttons {
		if text == query {
			return text
		}
	}

	// Try case-insensitive substring match
	queryLower := strings.ToLower(query)
	for _, text := range buttons {
		if strings.Contains(strings.ToLower(text), queryLower) {
			return text
		}
	}

	fmt.Fprintf(os.Stderr, "Error: button %q not found\n", query)
	fmt.Fprintln(os.Stderr, "Available buttons:")
	for i, text := range buttons {
		fmt.Fprintf(os.Stderr, "  [%d] %s\n", i, text)
	}
	os.Exit(1)
	return "" // unreachable
}
