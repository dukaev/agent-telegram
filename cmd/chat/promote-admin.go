// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

// PromoteAdminCmd represents the promote-admin command.
var PromoteAdminCmd = cliutil.NewSimpleCommand(cliutil.SimpleCommandDef{
	Use:   "promote-admin",
	Short: "Promote a user to admin",
	Long: `Promote a user to administrator in a Telegram channel or supergroup.

Use --peer @username or --peer username to specify the channel.
Use --user @username or --user username to specify the user to promote.
Use various flags to set admin permissions.

Example:
  agent-telegram promote-admin --peer @mychannel --user @username --can-change-info --can-ban-users`,
	Method: "promote_admin",
	Flags: []cliutil.Flag{
		cliutil.PeerFlag,
		cliutil.UserFlag,
		{Name: "canChangeInfo", Short: "", Usage: "Can change info", Type: cliutil.FlagBool},
		{Name: "canPostMessages", Short: "", Usage: "Can post messages", Type: cliutil.FlagBool},
		{Name: "canEditMessages", Short: "", Usage: "Can edit messages", Type: cliutil.FlagBool},
		{Name: "canDeleteMessages", Short: "", Usage: "Can delete messages", Type: cliutil.FlagBool},
		{Name: "canBanUsers", Short: "", Usage: "Can ban users", Type: cliutil.FlagBool},
		{Name: "canInviteUsers", Short: "", Usage: "Can invite users", Type: cliutil.FlagBool},
		{Name: "canPinMessages", Short: "", Usage: "Can pin messages", Type: cliutil.FlagBool},
		{Name: "canAddAdmins", Short: "", Usage: "Can add admins", Type: cliutil.FlagBool},
		{Name: "anonymous", Short: "", Usage: "Anonymous admin", Type: cliutil.FlagBool},
	},
	Success: "User promoted to admin successfully",
})

// AddPromoteAdminCommand adds the promote-admin command to the root command.
func AddPromoteAdminCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(PromoteAdminCmd)
}
