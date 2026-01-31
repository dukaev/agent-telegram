// Package cmd provides CLI commands.
//nolint:dupl // Similar structure to send-file but for different command
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
	runner := NewRunnerFromRoot(sendVideoJSON)
	peer := args[0]
	filePath := args[1]

	result := runner.CallWithParams("send_video", map[string]any{
		"peer":    peer,
		"file":    filePath,
		"caption": sendVideoCaption,
	})

	runner.PrintResult(result, func(r any) {
		rMap, _ := r.(map[string]any)
		id, _ := rMap["id"].(float64)
		peer, _ := rMap["peer"].(string)
		fmt.Printf("Video sent successfully!\n")
		fmt.Printf("  Peer: @%s\n", peer)
		fmt.Printf("  ID: %d\n", int64(id))
	})
}
