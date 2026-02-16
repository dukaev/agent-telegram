// Package gift provides commands for managing star gifts.
package gift

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	priceMsgID int
	priceValue int64
)

// PriceCmd represents the gift price command.
var PriceCmd = &cobra.Command{
	Use:   "price [slug]",
	Short: "Set resale price on a star gift",
	Long: `Set or update the resale price on a saved star gift.
Specify the gift by slug (positional) or --msg-id.

Example:
  agent-telegram gift price SwissWatch-718 --stars 5000
  agent-telegram gift price --msg-id 123 --stars 1000`,
	Args: cobra.MaximumNArgs(1),
}

// AddPriceCommand adds the price command to the parent command.
func AddPriceCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(PriceCmd)

	PriceCmd.Flags().IntVar(&priceMsgID, "msg-id", 0, "Message ID of the gift")
	PriceCmd.Flags().Int64Var(&priceValue, "stars", 0, "Resale price in stars")
	_ = PriceCmd.MarkFlagRequired("stars")

	PriceCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(PriceCmd, false)
		params := map[string]any{
			"price": priceValue,
		}
		if len(args) > 0 {
			params["slug"] = args[0]
		}
		if priceMsgID != 0 {
			params["msgId"] = priceMsgID
		}
		result := runner.CallWithParams("update_gift_price", params)
		runner.PrintResult(result, func(result any) {
			cliutil.PrintSuccessSummary(result, "Gift price updated successfully!")
		})
	}
}
