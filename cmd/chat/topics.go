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
	Use:   "topics [peer]",
	Short: "List forum topics in a chat",
	Long: `List all forum topics in a Telegram chat that has enabled forum mode.

Use a positional peer or --to @username to specify the chat.
Use --limit to set the maximum number of topics to return (max 100).

Example:
  agent-telegram chat topics @mybot --limit 20`,
	Args: cobra.MaximumNArgs(1),
}

// AddTopicsCommand adds the topics command to the root command.
func AddTopicsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(TopicsCmd)
	cliutil.MarkFirstArgPeer(TopicsCmd)

	TopicsCmd.Flags().VarP(&topicsTo, "to", "t", "Chat or bot peer (@username, username, or ID)")
	TopicsCmd.Flags().IntVarP(&topicsLimit, "limit", "l", cliutil.DefaultLimitMax, "Maximum number of topics (max 100)")
	TopicsCmd.Flags().IntVarP(&topicsOffset, "offset", "o", 0, "Offset for pagination")
	TopicsCmd.Run = func(_ *cobra.Command, args []string) {
		runner := cliutil.NewRunnerFromCmd(TopicsCmd, true) // Always JSON output
		if len(args) > 0 {
			_ = topicsTo.Set(args[0])
		}
		if topicsTo.Peer() == "" {
			runner.Fatal("peer is required (positional or --to)")
		}
		pag := cliutil.NewPagination(topicsLimit, topicsOffset, cliutil.PaginationConfig{
			MaxLimit: cliutil.MaxLimitStandard,
		})

		params := map[string]any{}
		topicsTo.AddToParams(params)
		pag.ToParams(params, true)

		result := runner.CallWithParams("get_topics", params)
		runner.PrintResult(result, func(r any) {
			cliutil.PrintTopics(r, naValue)
		})
	}
}
