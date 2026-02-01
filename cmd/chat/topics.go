// Package chat provides commands for managing chats.
package chat

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"agent-telegram/internal/cliutil"
)

var (
	topicsPeer string
	topicsLimit int
)

// TopicsCmd represents the topics command.
var TopicsCmd = &cobra.Command{
	Use:     "topics",
	Short:   "List forum topics in a channel",
	Long: `List all forum topics in a Telegram channel that has enabled forum mode.

Use --peer @username or --peer username to specify the channel.
Use --limit to set the maximum number of topics to return (max 100).

Example:
  agent-telegram topics --peer @mychannel --limit 20`,
	Args: cobra.NoArgs,
}

// AddTopicsCommand adds the topics command to the root command.
func AddTopicsCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(TopicsCmd)

	TopicsCmd.Flags().StringVarP(&topicsPeer, "peer", "p", "", "Channel username (@username or username)")
	TopicsCmd.Flags().IntVarP(&topicsLimit, "limit", "l", 100, "Maximum number of topics (max 100)")
	_ = TopicsCmd.MarkFlagRequired("peer")

	TopicsCmd.Run = func(_ *cobra.Command, _ []string) {
		// Validate and sanitize limit
		if topicsLimit < 1 {
			topicsLimit = 1
		}
		if topicsLimit > 100 {
			topicsLimit = 100
		}

		runner := cliutil.NewRunnerFromCmd(TopicsCmd, true) // Always JSON output
		params := map[string]any{
			"peer":  topicsPeer,
			"limit": topicsLimit,
		}

		result := runner.CallWithParams("get_topics", params)
		//nolint:errchkjson // Output to stdout, error handling not required
		_ = json.NewEncoder(os.Stdout).Encode(result)

		// Print human-readable summary
		cliutil.PrintTopics(result, naValue)
	}
}
