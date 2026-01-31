// Package cmd provides CLI commands.
package cmd

import (
	"github.com/spf13/cobra"
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
	runner := NewRunnerFromRoot(sendPhotoJSON)
	peer := args[0]
	filePath := args[1]

	result := runner.CallWithParams("send_photo", map[string]any{
		"peer": peer,
		"file": filePath,
	})

	runner.PrintResult(result, func(r any) {
		FormatSuccess(r, "Photo")
	})
}
