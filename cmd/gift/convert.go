// Package gift provides commands for managing star gifts.
package gift

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	convertMsgID int
	convertSlug  string
)

// ConvertCmd represents the gift convert command.
var ConvertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert a star gift to stars",
	Long: `Convert a saved star gift into Telegram stars.
Specify the gift by either --msg-id or --slug.

Example:
  agent-telegram gift convert --msg-id 123
  agent-telegram gift convert --slug gift_slug`,
	Args: cobra.NoArgs,
}

// AddConvertCommand adds the convert command to the parent command.
func AddConvertCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(ConvertCmd)

	ConvertCmd.Flags().IntVar(&convertMsgID, "msg-id", 0, "Message ID of the gift")
	ConvertCmd.Flags().StringVar(&convertSlug, "slug", "", "Gift slug")

	ConvertCmd.Run = func(_ *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(ConvertCmd, false)
		params := map[string]any{}
		if convertMsgID != 0 {
			params["msgId"] = convertMsgID
		}
		if convertSlug != "" {
			params["slug"] = convertSlug
		}
		result := runner.CallWithParams("convert_star_gift", params)
		runner.PrintResult(result, func(result any) {
			cliutil.PrintSuccessSummary(result, "Star gift converted to stars successfully!")
		})
	}
}
