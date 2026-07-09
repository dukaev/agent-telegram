// Package user provides commands for managing user-related operations.
package user

import (
	"agent-telegram/cmd/mute"
	"github.com/spf13/cobra"
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
	parentCmd.AddCommand(mute.NewMuteCommand())
}
