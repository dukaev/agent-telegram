// Package chat provides commands for managing chats.
package chat

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	subscribeChannel string
)

// SubscribeCmd represents the subscribe command.
var SubscribeCmd = &cobra.Command{
	Use:     "subscribe",
	Short:   "Subscribe to a public channel",
	Long: `Subscribe to a public Telegram channel by username.

Use --channel @username or --channel username to specify the channel.

Example:
  agent-telegram subscribe --channel @telegram`,
	Args: cobra.NoArgs,
}

// AddSubscribeCommand adds the subscribe command to the root command.
func AddSubscribeCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(SubscribeCmd)

	SubscribeCmd.Flags().StringVarP(&subscribeChannel, "channel", "c", "", "Channel username (@username or username)")
	_ = SubscribeCmd.MarkFlagRequired("channel")

	SubscribeCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(SubscribeCmd, false)
		params := map[string]any{
			"channel": subscribeChannel,
		}

		result := runner.CallWithParams("subscribe_channel", params)
		runner.PrintResult(result, func(result any) {
			r, ok := result.(map[string]any)
			if !ok {
				fmt.Println("Subscribed to channel successfully!")
				return
			}
			if title, ok := r["title"].(string); ok && title != "" {
				fmt.Printf("Subscribed to \"%s\" successfully!\n", title)
			} else {
				fmt.Println("Subscribed to channel successfully!")
			}
			if chatID, ok := r["chatId"].(float64); ok {
				fmt.Printf("Channel ID: %d\n", int64(chatID))
			}
		})
	}
}
