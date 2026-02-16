package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	contactFlags     SendFlags
	contactPhone     string
	contactFirstName string
	contactLastName  string
)

// ContactCmd represents the send contact subcommand.
var ContactCmd = &cobra.Command{
	Use:   "contact <peer>",
	Short: "Send a contact",
	Long: `Send a contact card to a Telegram user or chat.

Examples:
  agent-telegram send contact @user --phone +1234567890 --first-name John
  agent-telegram send contact --to @user --phone +1234567890 --first-name John --last-name Doe`,
	Args: cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if len(args) > 0 {
			_ = contactFlags.To.Set(args[0])
		}
		contactFlags.RequirePeer()

		runner := contactFlags.NewRunner()
		params := map[string]any{
			"phone": contactPhone,
		}
		contactFlags.To.AddToParams(params)
		if contactFirstName != "" {
			params["firstName"] = contactFirstName
		}
		if contactLastName != "" {
			params["lastName"] = contactLastName
		}

		result := runner.CallWithParams("send_contact", params)
		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, "send_contact")
		})
	},
}

func addContactCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ContactCmd)
	contactFlags.RegisterOptionalTo(ContactCmd)
	ContactCmd.Flags().StringVar(&contactPhone, "phone", "", "Phone number")
	ContactCmd.Flags().StringVar(&contactFirstName, "first-name", "", "First name")
	ContactCmd.Flags().StringVar(&contactLastName, "last-name", "", "Last name")
	_ = ContactCmd.MarkFlagRequired("phone")
}
