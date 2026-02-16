// Package gift provides commands for managing star gifts.
package gift

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	transferTo    cliutil.Recipient
	transferMsgID int
)

// TransferCmd represents the gift transfer command.
var TransferCmd = &cobra.Command{
	Use:   "transfer [slug]",
	Short: "Transfer a star gift to another user",
	Long: `Transfer a saved star gift to another Telegram user.
Specify the gift by slug (positional) or --msg-id.

Example:
  agent-telegram gift transfer SantaHat-55373 --to @username
  agent-telegram gift transfer --to @username --msg-id 123`,
	Args: cobra.MaximumNArgs(1),
}

// AddTransferCommand adds the transfer command to the parent command.
func AddTransferCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(TransferCmd)

	TransferCmd.Flags().VarP(&transferTo, "to", "t", "Recipient (@username or ID)")
	TransferCmd.Flags().IntVar(&transferMsgID, "msg-id", 0, "Message ID of the gift")
	_ = TransferCmd.MarkFlagRequired("to")

	TransferCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(TransferCmd, false)
		params := map[string]any{}
		transferTo.AddToParams(params)
		if len(args) > 0 {
			params["slug"] = args[0]
		}
		if transferMsgID != 0 {
			params["msgId"] = transferMsgID
		}
		result := runner.CallWithParams("transfer_star_gift", params)
		runner.PrintResult(result, func(result any) {
			cliutil.PrintSuccessSummary(result, "Star gift transferred successfully!")
		})
	}
}
