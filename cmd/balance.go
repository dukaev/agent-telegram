// Package cmd provides the root command and CLI configuration.
package cmd

import (
	"github.com/spf13/cobra"

	"agent-telegram/cmd/gift"
	"agent-telegram/internal/cliutil"
)

// BalanceCmd represents the top-level balance command (alias for gift balance).
var BalanceCmd = &cobra.Command{
	GroupID: GroupIDChat,
	Use:     "balance",
	Short:   "Show stars and TON balance",
	Long:    `Show your current Telegram Stars and TON balance. Shortcut for 'gift balance'.`,
	Example: `  agent-telegram balance`,
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		runner := cliutil.NewRunnerFromCmd(cmd, false)
		result := runner.CallWithParams("get_balance", map[string]any{})
		runner.PrintResult(result, gift.PrintBalance)
	},
}

func init() {
	RootCmd.AddCommand(BalanceCmd)
}
