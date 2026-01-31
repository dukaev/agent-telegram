// Package cmd provides CLI commands.
//nolint:dupl // Similar to send-file but for different command
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/ipc"
)

var (
	sendVideoJSON    bool
	sendVideoCaption string
)

// sendVideoCmd represents the send-video command.
var sendVideoCmd = &cobra.Command{
	Use:   "send-video @peer <file>",
	Short: "Send a video to a Telegram peer",
	Long: `Send a video file to a Telegram user or chat.

Supported formats: mp4, mov, avi, mkv, webm

Example: agent-telegram send-video @user /path/to/video.mp4`,
	Args: cobra.ExactArgs(2),
	Run:  runSendVideo,
}

func init() {
	rootCmd.AddCommand(sendVideoCmd)

	sendVideoCmd.Flags().BoolVarP(&sendVideoJSON, "json", "j", false, "Output as JSON")
	sendVideoCmd.Flags().StringVar(&sendVideoCaption, "caption", "", "Video caption")
}

func runSendVideo(_ *cobra.Command, args []string) {
	socketPath, _ := rootCmd.Flags().GetString("socket")
	peer := args[0]
	filePath := args[1]

	client := ipc.NewClient(socketPath)
	result, rpcErr := client.Call("send_video", map[string]any{
		"peer":    peer,
		"file":    filePath,
		"caption": sendVideoCaption,
	})
	if rpcErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", rpcErr.Message)
		os.Exit(1)
	}

	if sendVideoJSON {
		printSendVideoJSON(result)
	} else {
		printSendVideoResult(result)
	}
}

// printSendVideoJSON prints the result as JSON.
func printSendVideoJSON(result any) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

// printSendVideoResult prints the result in a human-readable format.
func printSendVideoResult(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid response type\n")
		os.Exit(1)
	}

	id, _ := r["id"].(float64)
	peer, _ := r["peer"].(string)

	fmt.Printf("Video sent successfully!\n")
	fmt.Printf("  Peer: @%s\n", peer)
	fmt.Printf("  ID: %d\n", int64(id))
}
