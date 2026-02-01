// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// SlowModeCmd represents the slow-mode command.
var SlowModeCmd = cliutil.NewSimpleCommand(cliutil.SimpleCommandDef{
	Use:   "slow-mode",
	Short: "Set slow mode for a channel",
	Long: `Set slow mode (delay between messages) for a channel or supergroup.

Allowed values for --seconds: 0 (off), 10, 30, 60, 300, 900, 3600

Example:
  agent-telegram chat slow-mode --to @mychannel --seconds 30
  agent-telegram chat slow-mode --to @mychannel --seconds 0`,
	Method: "set_slow_mode",
	Flags: []cliutil.Flag{
		cliutil.ToFlag,
		{Name: "seconds", Short: "s", Usage: "Seconds between messages (0 to disable)", Type: cliutil.FlagInt},
	},
	Success: "Slow mode updated",
})

// AddSlowModeCommand adds the slow-mode command to the root command.
func AddSlowModeCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(SlowModeCmd)
}
