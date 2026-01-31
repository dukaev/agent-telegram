// Package cmd provides CLI commands.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

var (
	sendMessageJSON bool
)

// sendMessageCmd represents the send-message command.
var sendMessageCmd = &cobra.Command{
	Use:   "send-message @peer <message>",
	Short: "Send a message to a Telegram peer",
	Long: `Send a message to a Telegram user or chat by username.

Supports sending to users, groups, and channels.`,
	Args: cobra.MinimumNArgs(2),
	Run:  runSendMessage,
}

func init() {
	rootCmd.AddCommand(sendMessageCmd)

	sendMessageCmd.Flags().BoolVarP(&sendMessageJSON, "json", "j", false, "Output as JSON")
}

func runSendMessage(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]
	message := args[1]

	client := ipc.NewClient(socketPath)
	result, err := client.Call("send_message", map[string]any{
		"peer":    peer,
		"message": message,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if sendMessageJSON {
		printSendMessageJSON(result)
	} else {
		printSendMessageResult(result)
	}
}

// printSendMessageJSON prints the result as JSON.
func printSendMessageJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printSendMessageResult prints the result in a human-readable format.
func printSendMessageResult(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	id, _ := r["id"].(float64)
	peer, _ := r["peer"].(string)

	fmt.Printf("Message sent successfully!\n")
	fmt.Printf("  Peer: @%s\n", peer)
	fmt.Printf("  ID: %d\n", int64(id))
}
