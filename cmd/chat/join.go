// Package chat provides commands for managing chats.
package chat

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	joinInviteLink string
)

// JoinChatCmd represents the join command.
var JoinChatCmd = &cobra.Command{
	Use:     "join",
	Short:   "Join a chat or channel using an invite link",
	Long: `Join a Telegram chat or channel using an invite link.

Supports various invite link formats:
  - https://t.me/+hash
  - https://t.me/joinchat/hash
  - tg://join?invite=hash
  - +hash
  - hash

Example:
  agent-telegram join https://t.me/+abc123`,
	Args: cobra.NoArgs,
}

// AddJoinChatCommand adds the join command to the root command.
func AddJoinChatCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(JoinChatCmd)

	JoinChatCmd.Flags().StringVarP(&joinInviteLink, "link", "l", "", "Invite link to join")
	_ = JoinChatCmd.MarkFlagRequired("link")

	JoinChatCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(JoinChatCmd, false)
		params := map[string]any{
			"inviteLink": joinInviteLink,
		}

		result := runner.CallWithParams("join_chat", params)
		runner.PrintResult(result, func(result any) {
			r, ok := result.(map[string]any)
			if !ok {
				fmt.Println("Joined chat successfully!")
				return
			}
			if chatID, ok := r["chatId"].(float64); ok {
				fmt.Printf("Joined chat successfully! Chat ID: %d\n", int64(chatID))
			} else {
				fmt.Println("Joined chat successfully!")
			}
			if title, ok := r["title"].(string); ok && title != "" {
				fmt.Printf("Title: %s\n", title)
			}
		})
	}
}
