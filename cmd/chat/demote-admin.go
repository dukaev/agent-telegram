// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// DemoteAdminCmd represents the demote-admin command.
var DemoteAdminCmd = cliutil.NewSimpleCommand(cliutil.SimpleCommandDef{
	Use:     "demote-admin",
	Short:   "Demote an admin to regular user",
	Long: `Demote an administrator to a regular user.

Example:
  agent-telegram chat demote-admin --to @mychannel --user @username`,
	Method:  "demote_admin",
	Flags:   []cliutil.Flag{cliutil.ToFlag, cliutil.UserFlag},
	Success: "Admin demoted successfully",
})

// AddDemoteAdminCommand adds the demote-admin command to the root command.
func AddDemoteAdminCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(DemoteAdminCmd)
}
