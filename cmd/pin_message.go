// Package cmd provides CLI commands.
package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

// pinMessageCmd represents the pin-message command.
var pinMessageCmd = &cobra.Command{
	Use:   "pin-message @peer <message_id>",
	Short: "Pin a Telegram message",
	Long: `Pin a message in a chat.

Example: agent-telegram pin-message @user 123456`,
	Args: cobra.ExactArgs(2),
	Run:  runPinMessage,
}

func init() {
	rootCmd.AddCommand(pinMessageCmd)
}

func runPinMessage(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]
	messageID, _ := strconv.ParseInt(args[1], 10, 64)

	client := ipc.NewClient(socketPath)
	_, rpcErr := client.Call("pin_message", map[string]any{
		"peer":      peer,
		"messageId": messageID,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	fmt.Printf("Message pinned successfully!\n")
}

// unpinMessageCmd represents the unpin-message command.
var unpinMessageCmd = &cobra.Command{
	Use:   "unpin-message @peer <message_id>",
	Short: "Unpin a Telegram message",
	Long: `Unpin a previously pinned message.

Example: agent-telegram unpin-message @user 123456`,
	Args: cobra.ExactArgs(2),
	Run:  runUnpinMessage,
}

func init() {
	rootCmd.AddCommand(unpinMessageCmd)
}

func runUnpinMessage(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]
	messageID, _ := strconv.ParseInt(args[1], 10, 64)

	client := ipc.NewClient(socketPath)
	_, rpcErr := client.Call("unpin_message", map[string]any{
		"peer":      peer,
		"messageId": messageID,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	fmt.Printf("Message unpinned successfully!\n")
}
