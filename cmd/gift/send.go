// Package gift provides commands for managing star gifts.
package gift

import (
	"strconv"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	sendTo       cliutil.Recipient
	sendHideName bool
	sendMessage  string
)

// SendCmd represents the gift send command.
var SendCmd = &cobra.Command{
	Use:   "send <gift_name_or_id>",
	Short: "Send a star gift to a user",
	Long: `Buy a star gift from the catalog and send it to a Telegram user.
Payment is made in Telegram Stars. Specify the gift by name or numeric ID.

Example:
  agent-telegram gift send Heart --to @username
  agent-telegram gift send Heart --to @username -m "Happy birthday!"
  agent-telegram gift send 5170145012310081615 --to @username
  agent-telegram gift send Heart --to @username --hide-name`,
	Args: cobra.ExactArgs(1),
}

// AddSendCommand adds the send command to the parent command.
func AddSendCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(SendCmd)

	SendCmd.Flags().VarP(&sendTo, "to", "t", "Recipient (@username or ID)")
	SendCmd.Flags().StringVarP(&sendMessage, "msg", "m", "", "Message to attach with the gift")
	SendCmd.Flags().BoolVar(&sendHideName, "hide-name", false, "Hide your name from the recipient's profile")
	_ = SendCmd.MarkFlagRequired("to")

	SendCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(SendCmd, false)
		params := map[string]any{}
		sendTo.AddToParams(params)
		if sendMessage != "" {
			params["message"] = sendMessage
		}
		if sendHideName {
			params["hideName"] = true
		}

		// Try parsing as numeric ID, otherwise treat as name
		if id, err := strconv.ParseInt(args[0], 10, 64); err == nil {
			params["giftId"] = id
		} else {
			params["name"] = args[0]
		}

		result := runner.CallWithParams("send_star_gift", params)
		runner.PrintResult(result, func(result any) {
			cliutil.PrintSuccessSummary(result, "Star gift sent successfully!")
		})
	}
}
