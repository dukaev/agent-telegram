// Package send provides commands for sending messages and media.
package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	sendVideoFlags SendFlags
)

// VideoCmd represents the send-video command.
var VideoCmd = &cobra.Command{
	Use:   "send-video <file>",
	Short: "Send a video to a Telegram peer",
	Long: `Send a video file to a Telegram user or chat.

Supported formats: mp4, mov, avi, mkv, webm

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
	Args: cobra.ExactArgs(1),
}

// AddVideoCommand adds the video command to the root command.
func AddVideoCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(VideoCmd)

	VideoCmd.Run = func(_ *cobra.Command, args []string) {
		runner := sendVideoFlags.NewRunner()
		filePath := args[0]

		params := map[string]any{"file": filePath}
		sendVideoFlags.AddToParams(params)

		result := runner.CallWithParams("send_video", params)

		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, "Video")
		})
	}

	sendVideoFlags.Register(VideoCmd)
}
