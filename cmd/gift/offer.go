// Package gift provides commands for managing star gifts.
package gift

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	offerTo       cliutil.Recipient
	offerPrice    int64
	offerDuration int
)

// OfferCmd represents the gift offer command.
var OfferCmd = &cobra.Command{
	Use:   "offer <slug>",
	Short: "Make an offer to buy someone's gift",
	Long: `Make an offer to buy a star gift owned by another user.
Specify the gift by its unique slug.`,
	Example: `  agent-telegram gift offer SwissWatch-718 --to @username --stars 5000
  agent-telegram gift offer RestlessJar-55271 --to @username --stars 10000 --duration 86400`,
	Args: cobra.ExactArgs(1),
}

// AddOfferCommand adds the offer command to the parent command.
func AddOfferCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(OfferCmd)

	OfferCmd.Flags().VarP(&offerTo, "to", "t", "Gift owner (@username or ID)")
	OfferCmd.Flags().Int64Var(&offerPrice, "stars", 0, "Offer price in stars")
	OfferCmd.Flags().IntVar(&offerDuration, "duration", 0, "Offer duration in seconds (optional)")
	_ = OfferCmd.MarkFlagRequired("to")
	_ = OfferCmd.MarkFlagRequired("stars")

	OfferCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(OfferCmd, false)
		params := map[string]any{
			"slug":  args[0],
			"price": offerPrice,
		}
		offerTo.AddToParams(params)
		if offerDuration > 0 {
			params["duration"] = offerDuration
		}
		result := runner.CallWithParams("offer_gift", params)
		runner.PrintResult(result, func(result any) {
			cliutil.PrintSuccessWithDuration(result, "Offer sent successfully!", runner.LastDuration())
		})
	}
}
