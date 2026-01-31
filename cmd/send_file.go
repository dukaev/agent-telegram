// Package cmd provides CLI commands.
//nolint:dupl // Similar structure to send-video but for different command
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	sendFileJSON   bool
	sendFileCaption string
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
	runner := NewRunnerFromRoot(sendFileJSON)
	peer := args[0]
	filePath := args[1]

	result := runner.CallWithParams("send_file", map[string]any{
		"peer":    peer,
		"file":    filePath,
		"caption": sendFileCaption,
	})

	runner.PrintResult(result, func(r any) {
		rMap, _ := r.(map[string]any)
		id, _ := rMap["id"].(float64)
		peer, _ := rMap["peer"].(string)
		fmt.Printf("File sent successfully!\n")
		fmt.Printf("  Peer: @%s\n", peer)
		fmt.Printf("  ID: %d\n", int64(id))
	})
}
