// Package chat provides commands for managing chats.
package chat

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	archiveTo      cliutil.Recipient
	archiveDisable bool
)

// ArchiveCmd represents the archive command.
var ArchiveCmd = &cobra.Command{
	GroupID: "chat",
	Use:     "archive",
	Short:   "Archive or unarchive a Telegram chat",
	Long: `Archive or unarchive a Telegram chat.

Archived chats are moved to the Archived folder and hidden from the main chat list.
Use --disable to unarchive a previously archived chat.

Use --to @username, --to username, or --to <chat_id> to specify the chat.`,
	Args: cobra.NoArgs,
}

// AddArchiveCommand adds the archive command to the root command.
func AddArchiveCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ArchiveCmd)

	ArchiveCmd.Flags().VarP(&archiveTo, "to", "t", "Recipient (@username, username, or chat ID)")
	ArchiveCmd.Flags().BoolVarP(&archiveDisable, "disable", "d", false, "Unarchive the chat")
	_ = ArchiveCmd.MarkFlagRequired("to")

	ArchiveCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(ArchiveCmd, false)
		params := map[string]any{}
		archiveTo.AddToParams(params)

		var method string
		var successMsg string
		if archiveDisable {
			method = "unarchive"
			successMsg = "Chat unarchived successfully!"
		} else {
			method = "archive"
			successMsg = "Chat archived successfully!"
		}

		result := runner.CallWithParams(method, params)
		runner.PrintResult(result, func(result any) {
			r, ok := result.(map[string]any)
			if !ok {
				fmt.Printf("%s\n", successMsg)
				return
			}
			peer := cliutil.ExtractString(r, "peer")
			fmt.Printf("%s\n", successMsg)
			fmt.Printf("  Peer: %s\n", peer)
		})
	}
}
