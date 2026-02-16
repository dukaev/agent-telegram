// Package cmd provides the root command and CLI configuration.
package cmd

import (
	"github.com/spf13/cobra"

	"agent-telegram/cmd/gift"
	"agent-telegram/internal/cliutil"
)

var (
	giftsLimit  int
	giftsOffset string
	giftsTo     cliutil.Recipient
)

// GiftsCmd represents the top-level gifts command (alias for gift my).
var GiftsCmd = &cobra.Command{
	GroupID: GroupIDChat,
	Use:     "gifts",
	Aliases: []string{"inventory"},
	Short:   "List your saved star gifts",
	Long:    `List saved/received star gifts. Shortcut for 'gift my'.`,
	Example: `  agent-telegram gifts
  agent-telegram gifts --limit 20
  agent-telegram gifts --to @username`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(cmd, false)
		params := map[string]any{
			"limit": giftsLimit,
		}
		if giftsOffset != "" {
			params["offset"] = giftsOffset
		}
		if giftsTo.Peer() != "" {
			params["peer"] = giftsTo.Peer()
		}
		result := runner.CallWithParams("get_saved_gifts", params)
		runner.PrintResult(result, gift.PrintSavedGifts)
	},
}

func init() {
	GiftsCmd.Flags().IntVarP(&giftsLimit, "limit", "l", cliutil.DefaultLimitLarge, "Max gifts to show")
	GiftsCmd.Flags().StringVarP(&giftsOffset, "offset", "o", "", "Offset for pagination")
	GiftsCmd.Flags().VarP(&giftsTo, "to", "t", "User whose gifts to view (@username or ID)")

	RootCmd.AddCommand(GiftsCmd)
}
