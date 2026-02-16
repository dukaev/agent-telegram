// Package chat provides commands for managing chats.
package chat

import (
	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	topicsTo     cliutil.Recipient
	topicsLimit  int
	topicsOffset int
)

// TopicsCmd represents the topics command.
var TopicsCmd = &cobra.Command{
	Use:     "topics",
	Short:   "List forum topics in a channel",
	Long: `List all forum topics in a Telegram channel that has enabled forum mode.

Use --to @username or --to username to specify the channel.
Use --limit to set the maximum number of topics to return (max 100).

Example:
  agent-telegram chat topics --to @mychannel --limit 20`,
	Args: cobra.NoArgs,
}

// AddTopicsCommand adds the topics command to the root command.
func AddTopicsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(TopicsCmd)

	TopicsCmd.Flags().VarP(&topicsTo, "to", "t", "Channel (@username or username)")
	TopicsCmd.Flags().IntVarP(&topicsLimit, "limit", "l", cliutil.DefaultLimitMax, "Maximum number of topics (max 100)")
	TopicsCmd.Flags().IntVarP(&topicsOffset, "offset", "o", 0, "Offset for pagination")
	_ = TopicsCmd.MarkFlagRequired("to")

	TopicsCmd.Run = func(_ *cobra.Command, _ []string) {
		pag := cliutil.NewPagination(topicsLimit, topicsOffset, cliutil.PaginationConfig{
			MaxLimit: cliutil.MaxLimitStandard,
		})

		runner := cliutil.NewRunnerFromCmd(TopicsCmd, true) // Always JSON output
		params := map[string]any{}
		topicsTo.AddToParams(params)
		pag.ToParams(params, true)

		result := runner.CallWithParams("get_topics", params)
		runner.PrintResult(result, func(r any) {
			cliutil.PrintTopics(r, naValue)
		})
	}
}
