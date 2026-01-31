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
	sendPhotoJSON bool
)

// sendPhotoCmd represents the send-photo command.
var sendPhotoCmd = &cobra.Command{
	Use:   "send-photo @peer <file>",
	Short: "Send a photo to a Telegram peer",
	Long: `Send a photo file to a Telegram user or chat.

Supported formats: jpg, jpeg, png, gif, webp

Example: agent-telegram send-photo @user /path/to/photo.jpg`,
	Args: cobra.ExactArgs(2),
	Run:  runSendPhoto,
}

func init() {
	rootCmd.AddCommand(sendPhotoCmd)

	sendPhotoCmd.Flags().BoolVarP(&sendPhotoJSON, "json", "j", false, "Output as JSON")
}

func runSendPhoto(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]
	filePath := args[1]

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("send_photo", map[string]any{
		"peer": peer,
		"file": filePath,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	if sendPhotoJSON {
		printSendPhotoJSON(result)
	} else {
		printSendPhotoResult(result)
	}
}

// printSendPhotoJSON prints the result as JSON.
func printSendPhotoJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printSendPhotoResult prints the result in a human-readable format.
func printSendPhotoResult(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	id, _ := r["id"].(float64)
	peer, _ := r["peer"].(string)

	fmt.Printf("Photo sent successfully!\n")
	fmt.Printf("  Peer: @%s\n", peer)
	fmt.Printf("  ID: %d\n", int64(id))
}
