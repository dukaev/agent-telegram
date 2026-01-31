// Package send provides commands for sending messages and media.
package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	sendFileFlags SendFlags
)

// FileCmd represents the send-file command.
var FileCmd = &cobra.Command{
	Use:   "send-file <file>",
	Short: "Send a file to a Telegram peer",
	Long: `Send a generic file to a Telegram user or chat.

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
	Args: cobra.ExactArgs(1),
}

// AddFileCommand adds the file command to the root command.
func AddFileCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(FileCmd)

	FileCmd.Run = func(_ *cobra.Command, args []string) {
		runner := sendFileFlags.NewRunner()
		filePath := args[0]

		params := map[string]any{"file": filePath}
		sendFileFlags.AddToParams(params)

		result := runner.CallWithParams("send_file", params)

		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, "File")
		})
	}

	sendFileFlags.Register(FileCmd)
}
