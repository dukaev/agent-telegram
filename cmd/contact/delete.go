// Package contact provides commands for managing contacts.
package contact

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	deleteUsername string
)

// DeleteContactCmd represents the contact delete command.
var DeleteContactCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete a contact from your Telegram account",
	Long: `Delete a contact from your Telegram account by username.

You can specify the username with or without the @ prefix.

Example:
  agent-telegram contact delete --username john
  agent-telegram contact delete --username @john`,
	Args: cobra.NoArgs,
}

// AddDeleteContactCommand adds the delete contact command to the root command.
func AddDeleteContactCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(DeleteContactCmd)

	DeleteContactCmd.Flags().StringVarP(&deleteUsername, "username", "u", "", "Username to delete")
	_ = DeleteContactCmd.MarkFlagRequired("username")

	DeleteContactCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(DeleteContactCmd, false)
		params := map[string]any{
			"username": deleteUsername,
		}

		result := runner.CallWithParams("delete_contact", params)
		runner.PrintResult(result, func(result any) {
			r, ok := result.(map[string]any)
			if !ok {
				fmt.Printf("Contact deleted successfully: %s\n", deleteUsername)
				return
			}

			success, _ := r["success"].(bool)
			if success {
				fmt.Printf("Contact deleted successfully: %s\n", deleteUsername)
			} else {
				fmt.Printf("Failed to delete contact: %s\n", deleteUsername)
			}
		})
	}
}
