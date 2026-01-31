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
	sendReplyJSON bool
)

// sendReplyCmd represents the send-reply command.
var sendReplyCmd = &cobra.Command{
	Use:   "send-reply @peer <message_id> <text>",
	Short: "Reply to a Telegram message",
	Long: `Send a reply to a specific message in a chat.

Example: agent-telegram send-reply @user 123456 "Thanks!"`,
	Args: cobra.ExactArgs(3),
	Run:  runSendReply,
}

func init() {
	sendReplyCmd.Flags().BoolVarP(&sendReplyJSON, "json", "j", false, "Output as JSON")
	rootCmd.AddCommand(sendReplyCmd)
}

func runSendReply(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]
	var messageID int64
	_, err := fmt.Sscanf(args[1], "%d", &messageID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid message ID: %v\n", err)
		os.Exit(1)
	}
	text := args[2]

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("send_reply", map[string]any{
		"peer":      peer,
		"messageId": messageID,
		"text":      text,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	if sendReplyJSON {
		printSendReplyJSON(result)
	} else {
		printSendReplyResult(result)
	}
}

func printSendReplyJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func printSendReplyResult(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	id, _ := r["id"].(float64)
	peer, _ := r["peer"].(string)

	fmt.Printf("Reply sent successfully!\n")
	fmt.Printf("  Peer: @%s\n", peer)
	fmt.Printf("  ID: %d\n", int64(id))
}
