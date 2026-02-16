// Package gift provides commands for managing star gifts.
package gift

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	transferTo    cliutil.Recipient
	transferMsgID int
	transferSlug  string
)

// TransferCmd represents the gift transfer command.
var TransferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer a star gift to another user",
	Long: `Transfer a saved star gift to another Telegram user.
Specify the gift by either --msg-id or --slug.

Example:
  agent-telegram gift transfer --to @username --msg-id 123
  agent-telegram gift transfer --to @username --slug gift_slug`,
	Args: cobra.NoArgs,
}

// AddTransferCommand adds the transfer command to the parent command.
func AddTransferCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(TransferCmd)

	TransferCmd.Flags().VarP(&transferTo, "to", "t", "Recipient (@username or ID)")
	TransferCmd.Flags().IntVar(&transferMsgID, "msg-id", 0, "Message ID of the gift")
	TransferCmd.Flags().StringVar(&transferSlug, "slug", "", "Gift slug")
	_ = TransferCmd.MarkFlagRequired("to")

	TransferCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(TransferCmd, false)
		params := map[string]any{}
		transferTo.AddToParams(params)
		if transferMsgID != 0 {
			params["msgId"] = transferMsgID
		}
		if transferSlug != "" {
			params["slug"] = transferSlug
		}
		result := runner.CallWithParams("transfer_star_gift", params)
		runner.PrintResult(result, func(result any) {
			cliutil.PrintSuccessSummary(result, "Star gift transferred successfully!")
		})
	}
}
