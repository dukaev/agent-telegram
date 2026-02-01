// Package chat provides commands for managing chats.
package chat

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	promoteAdminPeer string
	promoteAdminUser string
	promoteAdminCanChangeInfo bool
	promoteAdminCanPostMessages bool
	promoteAdminCanEditMessages bool
	promoteAdminCanDeleteMessages bool
	promoteAdminCanBanUsers bool
	promoteAdminCanInviteUsers bool
	promoteAdminCanPinMessages bool
	promoteAdminCanAddAdmins bool
	promoteAdminAnonymous bool
)

// PromoteAdminCmd represents the promote-admin command.
var PromoteAdminCmd = &cobra.Command{
	Use:     "promote-admin",
	Short:   "Promote a user to admin",
	Long: `Promote a user to administrator in a Telegram channel or supergroup.

Use --peer @username or --peer username to specify the channel.
Use --user @username or --user username to specify the user to promote.
Use various flags to set admin permissions.

Example:
  agent-telegram promote-admin --peer @mychannel --user @username --can-change-info --can-ban-users`,
	Args: cobra.NoArgs,
}

// AddPromoteAdminCommand adds the promote-admin command to the root command.
func AddPromoteAdminCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(PromoteAdminCmd)

	PromoteAdminCmd.Flags().StringVarP(&promoteAdminPeer, "peer", "p", "", "Channel username (@username or username)")
	PromoteAdminCmd.Flags().StringVarP(&promoteAdminUser, "user", "u", "", "User to promote (@username or username)")
	PromoteAdminCmd.Flags().BoolVar(&promoteAdminCanChangeInfo, "can-change-info", false, "Can change info")
	PromoteAdminCmd.Flags().BoolVar(&promoteAdminCanPostMessages, "can-post-messages", false, "Can post messages")
	PromoteAdminCmd.Flags().BoolVar(&promoteAdminCanEditMessages, "can-edit-messages", false, "Can edit messages")
	PromoteAdminCmd.Flags().BoolVar(&promoteAdminCanDeleteMessages, "can-delete-messages", false, "Can delete messages")
	PromoteAdminCmd.Flags().BoolVar(&promoteAdminCanBanUsers, "can-ban-users", false, "Can ban users")
	PromoteAdminCmd.Flags().BoolVar(&promoteAdminCanInviteUsers, "can-invite-users", false, "Can invite users")
	PromoteAdminCmd.Flags().BoolVar(&promoteAdminCanPinMessages, "can-pin-messages", false, "Can pin messages")
	PromoteAdminCmd.Flags().BoolVar(&promoteAdminCanAddAdmins, "can-add-admins", false, "Can add admins")
	PromoteAdminCmd.Flags().BoolVar(&promoteAdminAnonymous, "anonymous", false, "Anonymous admin")
	_ = PromoteAdminCmd.MarkFlagRequired("peer")
	_ = PromoteAdminCmd.MarkFlagRequired("user")

	PromoteAdminCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(PromoteAdminCmd, true)
		params := map[string]any{
			"peer":            promoteAdminPeer,
			"user":            promoteAdminUser,
			"canChangeInfo":   promoteAdminCanChangeInfo,
			"canPostMessages": promoteAdminCanPostMessages,
			"canEditMessages": promoteAdminCanEditMessages,
			"canDeleteMessages": promoteAdminCanDeleteMessages,
			"canBanUsers":     promoteAdminCanBanUsers,
			"canInviteUsers":  promoteAdminCanInviteUsers,
			"canPinMessages":  promoteAdminCanPinMessages,
			"canAddAdmins":    promoteAdminCanAddAdmins,
			"anonymous":       promoteAdminAnonymous,
		}

		result := runner.CallWithParams("promote_admin", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)

		// Print human-readable summary
		r, ok := result.(map[string]any)
		if ok {
			if success, ok := r["success"].(bool); ok && success {
				fmt.Fprintf(os.Stderr, "User promoted to admin successfully\n")
			}
		}
	}
}
