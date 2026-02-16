// Package gift provides commands for managing star gifts.
package gift

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var convertMsgID int

// ConvertCmd represents the gift convert command.
var ConvertCmd = &cobra.Command{
	Use:     "convert [slug]",
	Aliases: []string{"cashout"},
	Short:   "Convert a star gift to stars",
	Long: `Convert a saved star gift into Telegram stars.
Specify the gift by slug (positional) or --msg-id.`,
	Example: `  agent-telegram gift convert SantaHat-55373
  agent-telegram gift convert --msg-id 123`,
	Args: cobra.MaximumNArgs(1),
}

// AddConvertCommand adds the convert command to the parent command.
func AddConvertCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ConvertCmd)

	ConvertCmd.Flags().IntVar(&convertMsgID, "msg-id", 0, "Message ID of the gift")

	ConvertCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(ConvertCmd, false)
		params := map[string]any{}
		if len(args) > 0 {
			params["slug"] = args[0]
		}
		if convertMsgID != 0 {
			params["msgId"] = convertMsgID
		}
		result := runner.CallWithParams("convert_star_gift", params)
		runner.PrintResult(result, func(result any) {
			cliutil.PrintSuccessWithDuration(result, "Star gift converted to stars successfully!", runner.LastDuration())
		})
	}
}
