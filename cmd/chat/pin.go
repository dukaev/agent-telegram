// Package chat provides commands for managing chats.
package chat

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	pinChatTo cliutil.Recipient
	pinChatDisable bool
)

// PinChatCmd represents the pin command.
var PinChatCmd = &cobra.Command{
	Use:    "pin",
	Short:  "Pin or unpin a chat in the dialog list",
	Long: `Pin or unpin a chat in your Telegram dialog list.

Pinned chats stay at the top of your chat list.

Use --disable to unpin a previously pinned chat.
Use --to @username, --to username, or --to <chat_id> to specify the chat.`,
	Args:    cobra.NoArgs,
}

// AddPinChatCommand adds the pin command to the root command.
func AddPinChatCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(PinChatCmd)

	PinChatCmd.Flags().VarP(&pinChatTo, "to", "t", "Recipient (@username, username, or chat ID)")
	PinChatCmd.Flags().BoolVarP(&pinChatDisable, "disable", "d", false, "Unpin the chat")
	_ = PinChatCmd.MarkFlagRequired("to")

	PinChatCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(PinChatCmd, false)
		params := map[string]any{}
		pinChatTo.AddToParams(params)
		params["disable"] = pinChatDisable

		result := runner.CallWithParams("pin_chat", params)
		runner.PrintResult(result, func(result any) {
			r, ok := result.(map[string]any)
			if !ok {
				action := "pinned"
				if pinChatDisable {
					action = "unpinned"
				}
				fmt.Printf("Chat %s successfully!\n", action)
				return
			}
			peer := cliutil.ExtractString(r, "peer")
			pinned := false
			if v, ok := r["pinned"].(bool); ok {
				pinned = v
			}
			if pinned {
				fmt.Printf("Chat %s pinned successfully!\n", peer)
			} else {
				fmt.Printf("Chat %s unpinned successfully!\n", peer)
			}
		})
	}
}
