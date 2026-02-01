// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// BannedCmd represents the banned command.
var BannedCmd = cliutil.NewListCommand(cliutil.ListCommandConfig{
	Use:   "banned",
	Short: "List banned users in a channel",
	Long: `List all banned users in a Telegram channel.

Use --peer @username or --peer username to specify the channel.
Use --limit to set the maximum number of banned users to return (max 200).

Example:
  agent-telegram banned --peer @mychannel --limit 20`,
	Method:    "get_banned",
	PrintFunc: cliutil.PrintBanned,
	MaxLimit:  200,
	HasOffset: true,
})

// AddBannedCommand adds the banned command to the root command.
func AddBannedCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(BannedCmd)
}
