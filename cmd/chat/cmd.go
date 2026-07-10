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
	ChatCmd.AddCommand(PinChatCmd, JoinChatCmd, SubscribeCmd)
	AddTopicsCommand(ChatCmd)
	ChatCmd.AddCommand(MuteCmd, ArchiveCmd, CreateGroupCmd, CreateChannelCmd, EditTitleCmd,
		SetPhotoCmd, DeletePhotoCmd, LeaveCmd, InviteCmd, ParticipantsCmd, AdminsCmd,
		BannedCmd, PromoteAdminCmd, DemoteAdminCmd)
	AddInviteLinkCommand(ChatCmd)
	AddListCommand(ChatCmd)
	AddOpenCommand(ChatCmd)
	AddInfoCommand(ChatCmd)
	ChatCmd.AddCommand(SlowModeCmd)
	AddPermissionsCommand(ChatCmd)
	AddKeyboardCommand(ChatCmd)

	// Add the parent chat command to root
	rootCmd.AddCommand(ChatCmd)
}
