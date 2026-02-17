// Package session provides commands for managing Telegram sessions.
package session

import (
	"github.com/spf13/cobra"
)

// SessionCmd represents the parent session command.
var SessionCmd = &cobra.Command{
	GroupID: "server",
	Use:     "session",
	Short:   "Manage Telegram session",
	Long:    `Commands for managing the Telegram session file — export for use in environment variables.`,
}

// AddSessionCommand adds the parent session command and all its subcommands to the root command.
func AddSessionCommand(rootCmd *cobra.Command) {
	AddExportCommand(SessionCmd)
	rootCmd.AddCommand(SessionCmd)
}
