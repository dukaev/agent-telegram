// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	sendContactJSON bool
)

// sendContactCmd represents the send-contact command.
var sendContactCmd = &cobra.Command{
	Use:   "send-contact @peer <phone> <firstName>",
	Short: "Send a contact to a Telegram peer",
	Long: `Send a contact to a Telegram user or chat.

The contact will be shared with the peer, who can then add it to their contacts.

Example: agent-telegram send-contact @user +1234567890 "John"`,
	Args: cobra.RangeArgs(3, 4),
	Run:  runSendContact,
}

func init() {
	rootCmd.AddCommand(sendContactCmd)

	sendContactCmd.Flags().BoolVarP(&sendContactJSON, "json", "j", false, "Output as JSON")
}

func runSendContact(_ *cobra.Command, args []string) {
	runner := NewRunnerFromRoot(sendContactJSON)
	peer := args[0]
	phone := args[1]
	firstName := args[2]

	lastName := ""
	if len(args) > 3 {
		lastName = args[3]
	}

	result := runner.CallWithParams("send_contact", map[string]any{
		"peer":      peer,
		"phone":     phone,
		"firstName": firstName,
		"lastName":  lastName,
	})

	runner.PrintResult(result, func(r any) {
		rMap, _ := r.(map[string]any)
		id, _ := rMap["id"].(float64)
		peer, _ := rMap["peer"].(string)
		phone, _ := rMap["phone"].(string)
		fmt.Printf("Contact sent successfully!\n")
		fmt.Printf("  Peer: @%s\n", peer)
		fmt.Printf("  Phone: %s\n", phone)
		fmt.Printf("  ID: %d\n", int64(id))
	})
}
