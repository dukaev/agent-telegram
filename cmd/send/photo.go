// Package send provides commands for sending messages and media.
package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	sendPhotoFlags SendFlags
)

// PhotoCmd represents the send-photo command.
var PhotoCmd = &cobra.Command{
	Use:   "send-photo <file>",
	Short: "Send a photo to a Telegram peer",
	Long: `Send a photo file to a Telegram user or chat.

Supported formats: jpg, jpeg, png, gif, webp

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
	Args: cobra.ExactArgs(1),
}

// AddPhotoCommand adds the photo command to the root command.
func AddPhotoCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(PhotoCmd)

	PhotoCmd.Run = func(_ *cobra.Command, args []string) {
		runner := sendPhotoFlags.NewRunner()
		filePath := args[0]

		params := map[string]any{"file": filePath}
		sendPhotoFlags.AddToParams(params)

		result := runner.CallWithParams("send_photo", params)

		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, "Photo")
		})
	}

	sendPhotoFlags.Register(PhotoCmd)
}
