// Package contact provides commands for managing contacts.
package contact

import (
	"fmt"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	addPhone     string
	addFirstName string
	addLastName  string
)

// AddContactCmd represents the contact add command.
var AddContactCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add a new contact to your Telegram account",
	Long: `Add a new contact to your Telegram account using their phone number.

The phone number must include the country code (e.g., +1234567890).

Example:
  agent-telegram contact add --phone +1234567890 --first-name John
  agent-telegram contact add --phone +1234567890 --first-name John --last-name Doe`,
	Args: cobra.NoArgs,
}

// AddAddContactCommand adds the add contact command to the root command.
func AddAddContactCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(AddContactCmd)

	AddContactCmd.Flags().StringVarP(&addPhone, "phone", "p", "", "Phone number (with country code)")
	AddContactCmd.Flags().StringVarP(&addFirstName, "first-name", "f", "", "First name")
	AddContactCmd.Flags().StringVarP(&addLastName, "last-name", "l", "", "Last name (optional)")
	_ = AddContactCmd.MarkFlagRequired("phone")
	_ = AddContactCmd.MarkFlagRequired("first-name")

	AddContactCmd.Run = runAddContact
}

// runAddContact executes the add contact command.
func runAddContact(_ *cobra.Command, _ []string) {
	runner := cliutil.NewRunnerFromCmd(AddContactCmd, false)
	params := map[string]any{
		"phone":     addPhone,
		"firstName": addFirstName,
	}
	if addLastName != "" {
		params["lastName"] = addLastName
	}

	result := runner.CallWithParams("add_contact", params)
	runner.PrintResult(result, printContactResult)
}

// printContactResult prints the contact result.
func printContactResult(result any) {
	r, ok := result.(map[string]any)
	if !ok {
		fmt.Println("Contact added successfully!")
		return
	}

	contact, ok := r["contact"].(map[string]any)
	if !ok {
		fmt.Println("Contact added successfully!")
		return
	}

	firstName, _ := contact["firstName"].(string)
	lastName, _ := contact["lastName"].(string)
	username, _ := contact["username"].(string)
	phone, _ := contact["phone"].(string)
	peer, _ := contact["peer"].(string)

	// Build display name
	name := firstName
	if lastName != "" {
		if name != "" {
			name += " " + lastName
		} else {
			name = lastName
		}
	}
	if name == "" {
		name = "Contact"
	}

	fmt.Printf("Added contact: %s", name)
	if username != "" {
		fmt.Printf(" (@%s)", username)
	}
	if peer != "" {
		fmt.Printf(" [%s]", peer)
	}
	if phone != "" {
		fmt.Printf(" - %s", phone)
	}
	fmt.Println()
}
