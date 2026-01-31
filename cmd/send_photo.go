// Package cmd provides CLI commands.
package cmd

import (
	"github.com/spf13/cobra"
)

var sendPhotoFlags SendFlags

func init() {
	sendPhotoCmd := sendPhotoFlags.NewCommand(CommandConfig{
		Use:   "send-photo <file>",
		Short: "Send a photo to a Telegram peer",
		Long: `Send a photo file to a Telegram user or chat.

Supported formats: jpg, jpeg, png, gif, webp

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
		Args:       cobra.ExactArgs(1),
		Run:        runSendPhoto,
		HasCaption: true,
	})
	rootCmd.AddCommand(sendPhotoCmd)
}

func runSendPhoto(_ *cobra.Command, args []string) {
	runner := sendPhotoFlags.NewRunner()
	filePath := args[0]

	params := map[string]any{"file": filePath}
	sendPhotoFlags.AddToParams(params)

	result := runner.CallWithParams("send_photo", params)

	runner.PrintResult(result, func(r any) {
		FormatSuccess(r, "Photo")
	})
}
