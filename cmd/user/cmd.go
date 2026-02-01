// Package user provides commands for managing user-related operations.
package user

import (
	"github.com/spf13/cobra"
	"agent-telegram/cmd/mute"
)

// UserCmd represents the parent user command.
var UserCmd = &cobra.Command{
	GroupID: "chat",
	Use:     "user",
	Short:   "Manage user-related operations",
	Long:    `Commands for user-related operations like getting info, blocking, etc.`,
}

// AddUserCommand adds the parent user command and all its subcommands to the root command.
func AddUserCommand(rootCmd *cobra.Command) {
	// Add all subcommands to UserCmd
	AddBlockCommand(UserCmd)
	AddInfoCommand(UserCmd)
	AddMuteSubcommand(UserCmd)

	// Add the parent user command to root
	rootCmd.AddCommand(UserCmd)
}

// AddMuteSubcommand adds the mute command as a subcommand.
func AddMuteSubcommand(parentCmd *cobra.Command) {
	muteCmd := &cobra.Command{
		Use:     "mute",
		Short:   "Mute or unmute a Telegram chat",
		Long: `Mute or unmute a Telegram chat to control notifications.

Muted chats will not send you notifications for new messages.
Use --disable to unmute a previously muted chat.

Use --to @username, --to username, or --to <chat_id> to specify the chat.`,
		Args: cobra.NoArgs,
	}

	mute.SetupMuteFlags(muteCmd)
	mute.SetMuteRun(muteCmd)

	parentCmd.AddCommand(muteCmd)
}
