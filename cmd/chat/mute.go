// Package chat provides commands for managing chats.
// revive:disable:duplicated-code // Mute and archive commands follow the same pattern with different operations
package chat

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	muteTo      cliutil.Recipient
	muteDisable bool
)

// MuteCmd represents the mute command.
var MuteCmd = &cobra.Command{
	GroupID: "chat",
	Use:     "mute",
	Short:   "Mute or unmute a Telegram chat",
	Long: `Mute or unmute a Telegram chat to control notifications.

Muted chats will not send you notifications for new messages.
Use --disable to unmute a previously muted chat.

Use --to @username, --to username, or --to <chat_id> to specify the chat.`,
	Args: cobra.NoArgs,
}

// AddMuteCommand adds the mute command to the root command.
func AddMuteCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(MuteCmd)

	MuteCmd.Flags().VarP(&muteTo, "to", "t", "Recipient (@username, username, or chat ID)")
	MuteCmd.Flags().BoolVarP(&muteDisable, "disable", "d", false, "Unmute the chat")
	_ = MuteCmd.MarkFlagRequired("to")

	MuteCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(MuteCmd, false)
		params := map[string]any{}
		muteTo.AddToParams(params)

		var method string
		var successMsg string
		if muteDisable {
			method = "unmute"
			successMsg = "Chat unmuted successfully!"
		} else {
			method = "mute"
			successMsg = "Chat muted successfully!"
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
