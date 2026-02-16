// Package gift provides commands for managing star gifts.
package gift

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	sendTo       cliutil.Recipient
	sendSlug     string
	sendPrice    int64
	sendDuration int
)

// SendCmd represents the gift send command.
var SendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a star gift to a user",
	Long: `Buy and send a star gift to a Telegram user.

Example:
  agent-telegram gift send --to @username --slug gift_slug --price 100`,
	Args: cobra.NoArgs,
}

// AddSendCommand adds the send command to the parent command.
func AddSendCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(SendCmd)

	SendCmd.Flags().VarP(&sendTo, "to", "t", "Recipient (@username or ID)")
	SendCmd.Flags().StringVar(&sendSlug, "slug", "", "Gift slug")
	SendCmd.Flags().Int64Var(&sendPrice, "price", 0, "Price in stars")
	SendCmd.Flags().IntVar(&sendDuration, "duration", 0, "Duration (optional)")
	_ = SendCmd.MarkFlagRequired("to")
	_ = SendCmd.MarkFlagRequired("slug")
	_ = SendCmd.MarkFlagRequired("price")

	SendCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(SendCmd, false)
		params := map[string]any{
			"slug":  sendSlug,
			"price": sendPrice,
		}
		sendTo.AddToParams(params)
		if sendDuration > 0 {
			params["duration"] = sendDuration
		}
		result := runner.CallWithParams("send_star_gift", params)
		runner.PrintResult(result, func(result any) {
			cliutil.PrintSuccessSummary(result, "Star gift sent successfully!")
		})
	}
}
