// Package cmd provides CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var sendContactFlags SendFlags

func init() {
	sendContactCmd := sendContactFlags.NewCommand(CommandConfig{
		Use:   "send-contact <phone> <firstName> [lastName]",
		Short: "Send a contact to a Telegram peer",
		Long: `Send a contact to a Telegram user or chat.

The contact will be shared with the peer, who can then add it to their contacts.

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
		Args: cobra.RangeArgs(2, 3),
		Run:  runSendContact,
	})
	rootCmd.AddCommand(sendContactCmd)
}

func runSendContact(_ *cobra.Command, args []string) {
	runner := sendContactFlags.NewRunner()
	phone := args[0]
	firstName := args[1]

	lastName := ""
	if len(args) > 2 {
		lastName = args[2]
	}

	params := map[string]any{
		"phone":     phone,
		"firstName": firstName,
		"lastName":  lastName,
	}
	sendContactFlags.AddToParams(params)

	result := runner.CallWithParams("send_contact", params)

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
