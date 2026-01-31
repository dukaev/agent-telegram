// Package cmd provides CLI commands.
package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

// deleteMessageCmd represents the delete-message command.
var deleteMessageCmd = &cobra.Command{
	Use:   "delete-message <message_id>",
	Short: "Delete a Telegram message",
	Long: `Delete a specific message by ID.

Example: agent-telegram delete-message 123456`,
	Args: cobra.ExactArgs(1),
	Run:  runDeleteMessage,
}

func init() {
	rootCmd.AddCommand(deleteMessageCmd)
}

func runDeleteMessage(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	messageID, _ := strconv.ParseInt(args[0], 10, 64)

	client := ipc.NewClient(socketPath)
	_, rpcErr := client.Call("delete_message", map[string]any{
		"messageId": messageID,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	fmt.Printf("Message deleted successfully!\n")
}
