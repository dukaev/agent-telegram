// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

const (
	unknownName = "Unknown"
	naValue     = "N/A"
)

// ParticipantsCmd represents the participants command.
var ParticipantsCmd = cliutil.NewListCommand(cliutil.ListCommandConfig{
	Use:   "participants",
	Short: "List participants in a chat or channel",
	Long: `List all participants in a Telegram chat or channel.

Use --to @username or --to username to specify the chat/channel.
Use --limit to set the maximum number of participants to return (max 200).

Example:
  agent-telegram chat participants --to @mychannel --limit 50`,
	Method:    "get_participants",
	PrintFunc: cliutil.PrintParticipants,
	MaxLimit:  200,
	HasOffset: true,
})

// AddParticipantsCommand adds the participants command to the root command.
func AddParticipantsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ParticipantsCmd)
}
