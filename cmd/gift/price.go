// Package gift provides commands for managing star gifts.
package gift

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	priceMsgID int
	priceSlug  string
	priceValue int64
)

// PriceCmd represents the gift price command.
var PriceCmd = &cobra.Command{
	Use:   "price",
	Short: "Set resale price on a star gift",
	Long: `Set or update the resale price on a saved star gift.
Specify the gift by either --msg-id or --slug.

Example:
  agent-telegram gift price --slug SwissWatch-718 --stars 5000
  agent-telegram gift price --msg-id 123 --stars 1000`,
	Args: cobra.NoArgs,
}

// AddPriceCommand adds the price command to the parent command.
func AddPriceCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(PriceCmd)

	PriceCmd.Flags().IntVar(&priceMsgID, "msg-id", 0, "Message ID of the gift")
	PriceCmd.Flags().StringVar(&priceSlug, "slug", "", "Gift slug")
	PriceCmd.Flags().Int64Var(&priceValue, "stars", 0, "Resale price in stars")
	_ = PriceCmd.MarkFlagRequired("stars")

	PriceCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(PriceCmd, false)
		params := map[string]any{
			"price": priceValue,
		}
		if priceMsgID != 0 {
			params["msgId"] = priceMsgID
		}
		if priceSlug != "" {
			params["slug"] = priceSlug
		}
		result := runner.CallWithParams("update_gift_price", params)
		runner.PrintResult(result, func(result any) {
			cliutil.PrintSuccessSummary(result, "Gift price updated successfully!")
		})
	}
}
