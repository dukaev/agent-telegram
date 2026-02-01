// Package contact provides commands for managing contacts.
package contact

import (
	"github.com/spf13/cobra"
)

// ContactCmd represents the parent contact command.
var ContactCmd = &cobra.Command{
	Use:     "contact",
	Short:   "Manage Telegram contacts",
	Long:    `Commands for managing Telegram contacts - add, delete, and list.`,
}

// AddContactCommand adds the parent contact command and all its subcommands to the root command.
func AddContactCommand(rootCmd *cobra.Command) {
	// Add all subcommands to ContactCmd
	AddListContactsCommand(ContactCmd)
	AddAddContactCommand(ContactCmd)
	AddDeleteContactCommand(ContactCmd)

	// Add the parent contact command to root
	rootCmd.AddCommand(ContactCmd)
}
