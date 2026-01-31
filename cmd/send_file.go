// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var sendFileFlags SendFlags

func init() {
	sendFileCmd := sendFileFlags.NewCommand(CommandConfig{
		Use:   "send-file <file>",
		Short: "Send a file to a Telegram peer",
		Long: `Send any file to a Telegram user or chat.

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
		Args:       cobra.ExactArgs(1),
		Run:        runSendFile,
		HasCaption: true,
	})
	rootCmd.AddCommand(sendFileCmd)
}

func runSendFile(_ *cobra.Command, args []string) {
	runner := sendFileFlags.NewRunner()
	filePath := args[0]

	params := map[string]any{"file": filePath}
	sendFileFlags.AddToParams(params)

	result := runner.CallWithParams("send_file", params)

	runner.PrintResult(result, func(r any) {
		rMap, _ := r.(map[string]any)
		id, _ := rMap["id"].(float64)
		peer, _ := rMap["peer"].(string)
		fmt.Printf("File sent successfully!\n")
		fmt.Printf("  Peer: @%s\n", peer)
		fmt.Printf("  ID: %d\n", int64(id))
	})
}
