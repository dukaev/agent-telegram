// Package cmd provides CLI commands.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
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
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]
	messageID, _ := strconv.ParseInt(args[1], 10, 64)

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("inspect_inline_buttons", map[string]any{
		"peer":      peer,
		"messageId": messageID,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
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
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]
	messageID, _ := strconv.ParseInt(args[1], 10, 64)
	buttonIndex, _ := strconv.Atoi(args[2])

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("press_inline_button", map[string]any{
		"peer":        peer,
		"messageId":   messageID,
		"buttonIndex": buttonIndex,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}
