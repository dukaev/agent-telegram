// Package gift provides commands for managing star gifts.
package gift

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	buyTo cliutil.Recipient
)

// BuyCmd represents the gift buy command.
var BuyCmd = &cobra.Command{
	Use:   "buy <slug>",
	Short: "Buy a gift from the marketplace",
	Long: `Buy a unique star gift listed for resale on the marketplace.
Payment is made in Telegram Stars.`,
	Example: `  agent-telegram gift buy SantaHat-55373
  agent-telegram gift buy SwissWatch-718 --to @username`,
	Args: cobra.ExactArgs(1),
}

// AddBuyCommand adds the buy command to the parent command.
func AddBuyCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(BuyCmd)

	BuyCmd.Flags().VarP(&buyTo, "to", "t", "Recipient (@username or ID, defaults to self)")

	BuyCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(BuyCmd, false)
		params := map[string]any{
			"slug": args[0],
		}
		if buyTo.Peer() != "" {
			params["peer"] = buyTo.Peer()
		}
		result := runner.CallWithParams("buy_resale_gift", params)
		runner.PrintResult(result, func(result any) {
			cliutil.PrintSuccessWithDuration(result, "Gift purchased successfully!", runner.LastDuration())
		})
	}
}
