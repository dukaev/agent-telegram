// Package cmd provides CLI commands.
//nolint:dupl // Similar to send-file but different command
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var sendVideoFlags SendFlags

func init() {
	sendVideoCmd := sendVideoFlags.NewCommand(CommandConfig{
		Use:   "send-video <file>",
		Short: "Send a video to a Telegram peer",
		Long: `Send a video file to a Telegram user or chat.

Supported formats: mp4, mov, avi, mkv, webm

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
		Args:       cobra.ExactArgs(1),
		Run:        runSendVideo,
		HasCaption: true,
	})
	rootCmd.AddCommand(sendVideoCmd)
}

func runSendVideo(_ *cobra.Command, args []string) {
	runner := sendVideoFlags.NewRunner()
	filePath := args[0]

	params := map[string]any{"file": filePath}
	sendVideoFlags.AddToParams(params)

	result := runner.CallWithParams("send_video", params)

	runner.PrintResult(result, func(r any) {
		rMap, _ := r.(map[string]any)
		id, _ := rMap["id"].(float64)
		peer, _ := rMap["peer"].(string)
		fmt.Printf("Video sent successfully!\n")
		fmt.Printf("  Peer: @%s\n", peer)
		fmt.Printf("  ID: %d\n", int64(id))
	})
}
