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
	sendFileJSON bool
)

// sendFileCmd represents the send-file command.
var sendFileCmd = &cobra.Command{
	Use:   "send-file @peer <file>",
	Short: "Send a file to a Telegram peer",
	Long: `Send any file to a Telegram user or chat.

Example: agent-telegram send-file @user /path/to/file.pdf`,
	Args: cobra.ExactArgs(2),
	Run:  runSendFile,
}

func init() {
	rootCmd.AddCommand(sendFileCmd)

	sendFileCmd.Flags().BoolVarP(&sendFileJSON, "json", "j", false, "Output as JSON")
	sendFileCmd.Flags().StringVar(&sendFileCaption, "caption", "", "File caption")
}

func runSendFile(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]
	filePath := args[1]

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("send_file", map[string]any{
		"peer":    peer,
		"file":    filePath,
		"caption": sendFileCaption,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	if sendFileJSON {
		printSendFileJSON(result)
	} else {
		printSendFileResult(result)
	}
}

// printSendFileJSON prints the result as JSON.
func printSendFileJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printSendFileResult prints the result in a human-readable format.
func printSendFileResult(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	id, _ := r["id"].(float64)
	peer, _ := r["peer"].(string)

	fmt.Printf("File sent successfully!\n")
	fmt.Printf("  Peer: @%s\n", peer)
	fmt.Printf("  ID: %d\n", int64(id))
}
