// Package send provides commands for sending messages and media.
package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	sendContactFlags SendFlags
)

// ContactCmd represents the send-contact command.
var ContactCmd = &cobra.Command{
	Use:   "send-contact <phone> <first_name> [last_name]",
	Short: "Send a contact to a Telegram peer",
	Long: `Send a contact (vCard) to a Telegram user or chat.

Use --to @username, --to username, or --to <chat_id> to specify the recipient.`,
	Args: cobra.RangeArgs(2, 3),
}

// AddContactCommand adds the contact command to the root command.
func AddContactCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(ContactCmd)

	ContactCmd.Run = func(_ *cobra.Command, args []string) {
		runner := sendContactFlags.NewRunner()
		phone := args[0]
		firstName := args[1]
		lastName := ""
		if len(args) > 2 {
			lastName = args[2]
		}

		params := map[string]any{
			"phone":      phone,
			"firstName":  firstName,
			"lastName":   lastName,
		}
		sendContactFlags.AddToParams(params)

		result := runner.CallWithParams("send_contact", params)

		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, "Contact")
		})
	}

	sendContactFlags.RegisterWithoutCaption(ContactCmd)
}
