// Package cmd provides CLI commands.
package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

// updateMessageCmd represents the update-message command.
var updateMessageCmd = &cobra.Command{
	Use:   "update-message @peer <message_id> <new_text>",
	Short: "Edit a Telegram message",
	Long: `Edit a previously sent message.

Example: agent-telegram update-message @user 123456 "Updated text"`,
	Args: cobra.ExactArgs(3),
	Run:  runUpdateMessage,
}

func init() {
	rootCmd.AddCommand(updateMessageCmd)
}

func runUpdateMessage(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]
	messageID, _ := strconv.ParseInt(args[1], 10, 64)
	text := args[2]

	client := ipc.NewClient(socketPath)
	_, rpcErr := client.Call("update_message", map[string]any{
		"peer":      peer,
		"messageId": messageID,
		"text":      text,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	fmt.Printf("Message updated successfully!\n")
}
