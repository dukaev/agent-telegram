// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"
)

// ChatCmd represents the parent chat command.
var ChatCmd = &cobra.Command{
	GroupID: "chat",
	Use:     "chat",
	Short:   "Manage Telegram chats and channels",
	Long:    `Commands for managing Telegram chats, channels, groups, and their settings.`,
}

// AddChatCommand adds the parent chat command and all its subcommands to the root command.
func AddChatCommand(rootCmd *cobra.Command) {
	// Add all subcommands to ChatCmd
	AddPinChatCommand(ChatCmd)
	AddJoinChatCommand(ChatCmd)
	AddSubscribeCommand(ChatCmd)
	AddTopicsCommand(ChatCmd)
	AddMuteCommand(ChatCmd)
	AddArchiveCommand(ChatCmd)
	AddCreateGroupCommand(ChatCmd)
	AddCreateChannelCommand(ChatCmd)
	AddEditTitleCommand(ChatCmd)
	AddSetPhotoCommand(ChatCmd)
	AddDeletePhotoCommand(ChatCmd)
	AddLeaveCommand(ChatCmd)
	AddInviteCommand(ChatCmd)
	AddParticipantsCommand(ChatCmd)
	AddAdminsCommand(ChatCmd)
	AddBannedCommand(ChatCmd)
	AddPromoteAdminCommand(ChatCmd)
	AddDemoteAdminCommand(ChatCmd)
	AddInviteLinkCommand(ChatCmd)
	AddListCommand(ChatCmd)
	AddOpenCommand(ChatCmd)
	AddInfoCommand(ChatCmd)

	// Add the parent chat command to root
	rootCmd.AddCommand(ChatCmd)
}
