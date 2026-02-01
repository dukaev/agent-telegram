// Package user provides commands for managing user-related operations.
package user

import (
	"github.com/spf13/cobra"
)

// UserCmd represents the parent user command.
var UserCmd = &cobra.Command{
	Use:     "user",
	Short:   "Manage user-related operations",
	Long:    `Commands for user-related operations like getting info, blocking, etc.`,
}

// AddUserCommand adds the parent user command and all its subcommands to the root command.
func AddUserCommand(rootCmd *cobra.Command) {
	// Add all subcommands to UserCmd
	AddBlockCommand(UserCmd)
	AddInfoCommand(UserCmd)

	// Add the parent user command to root
	rootCmd.AddCommand(UserCmd)
}
