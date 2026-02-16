package send

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	pollSubFlags   SendFlags
	pollSubOptions []string
)

// PollCmd represents the send poll subcommand.
var PollCmd = &cobra.Command{
	Use:   "poll <peer> <question>",
	Short: "Send a poll",
	Long: `Send a poll to a Telegram user or chat.

Examples:
  agent-telegram send poll @user "What's your favorite?" --option "A" --option "B" --option "C"
  agent-telegram send poll --to @user "Yes or no?" --option Yes --option No`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(_ *cobra.Command, args []string) {
		var question string
		switch len(args) {
		case 2:
			_ = pollSubFlags.To.Set(args[0])
			question = args[1]
		case 1:
			if pollSubFlags.To.Peer() != "" {
				question = args[0]
			} else {
				_ = pollSubFlags.To.Set(args[0])
			}
		}
		pollSubFlags.RequirePeer()

		runner := pollSubFlags.NewRunner()
		params := map[string]any{
			"question": question,
		}
		pollSubFlags.To.AddToParams(params)

		if len(pollSubOptions) > 0 {
			optionMaps := make([]map[string]string, len(pollSubOptions))
			for i, opt := range pollSubOptions {
				optionMaps[i] = map[string]string{"text": opt}
			}
			params["options"] = optionMaps
		}

		result := runner.CallWithParams("send_poll", params)
		runner.PrintResult(result, func(r any) {
			cliutil.FormatSuccess(r, "send_poll")
		})
	},
}

func addPollCommand(parentCmd *cobra.Command) {
	parentCmd.AddCommand(PollCmd)
	pollSubFlags.RegisterOptionalTo(PollCmd)
	PollCmd.Flags().StringSliceVar(&pollSubOptions, "option", nil, "Poll options (can be used multiple times)")
}
